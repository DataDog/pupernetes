// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package run

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/DataDog/pupernetes/pkg/api"
	"github.com/DataDog/pupernetes/pkg/logging"
	"github.com/DataDog/pupernetes/pkg/run/state"
	"github.com/DataDog/pupernetes/pkg/setup"
	"github.com/DataDog/pupernetes/pkg/util"
)

const (
	appProbeThreshold = 10
)

// Runtime is the main state to execute a managed pupernetes Run
type Runtime struct {
	env *setup.Environment

	api *http.Server

	SigChan              chan os.Signal
	httpClient           *http.Client
	state                *state.State
	runTimeout           time.Duration
	DNSQueryForReadiness string
	waitKubeletGC        time.Duration
	kubeDeleteOption     *v1.DeleteOptions

	runTimestamp       time.Time
	journalTailerMutex sync.RWMutex
	journalTailers     map[string]*logging.JournalTailer

	ApplyChan chan struct{}
}

// NewRunner instantiate a new Runtimer with the given Environment
func NewRunner(env *setup.Environment, runTimeout, waitKubeletGC time.Duration, DNSQueryForReadiness string) (*Runtime, error) {
	var zero int64

	s, err := state.NewState()
	if err != nil {
		glog.Errorf("Cannot create the runner: %v", err)
		return nil, err
	}

	run := &Runtime{
		env:     env,
		state:   s,
		SigChan: make(chan os.Signal, 2),
		httpClient: &http.Client{
			Timeout: time.Millisecond * 500,
		},
		runTimeout:           runTimeout,
		DNSQueryForReadiness: DNSQueryForReadiness,
		waitKubeletGC:        waitKubeletGC,
		kubeDeleteOption: &v1.DeleteOptions{
			GracePeriodSeconds: &zero,
		},
		journalTailers: make(map[string]*logging.JournalTailer),
		runTimestamp:   time.Now(),
		ApplyChan:      make(chan struct{}),
	}
	run.api = api.NewAPI(run.SigChan, run.DeleteAPIManifests, run.state.IsReady, run.ApplyChan)
	return run, nil
}

// Run daemonise pupernetes
func (r *Runtime) Run() error {
	// the associated signal.Reset is defer in r.Stop method
	signal.Notify(r.SigChan, syscall.SIGTERM, syscall.SIGINT)

	defer close(r.ApplyChan)

	glog.Infof("Timeout for this current run is %s", r.runTimeout.String())
	timeoutTimer := time.NewTimer(r.runTimeout)
	defer timeoutTimer.Stop()

	go r.api.ListenAndServe()

	for _, u := range r.env.GetSystemdUnits() {
		err := util.StartUnit(r.env.GetDBUSClient(), u)
		if err != nil {
			return r.Stop(err)
		}
	}

	probeTick := time.NewTicker(time.Second * 2)
	defer probeTick.Stop()

	displayTick := time.NewTicker(time.Second * 5)
	defer displayTick.Stop()

	readinessTick := time.NewTicker(time.Second * 1)
	defer readinessTick.Stop()

	sigStopChan := make(chan os.Signal, 2)
	defer close(sigStopChan)
	signal.Notify(sigStopChan, syscall.SIGTSTP)

	kubeletProbeURL := fmt.Sprintf("http://127.0.0.1:%d/healthz", r.env.GetKubeletHealthzPort())
	for {
		select {
		case sig := <-r.SigChan:
			glog.Warningf("Signal received: %q, propagating ...", sig.String())
			return r.Stop(nil)

		case <-timeoutTimer.C:
			glog.Warningf("Timeout %s reached, stopping ...", r.runTimeout.String())
			return r.Stop(fmt.Errorf("timeout reached during run phase: %s", r.runTimeout.String()))

		case <-probeTick.C:
			_, err := r.probeUnitStatuses()
			if err != nil {
				r.SigChan <- syscall.SIGTERM
				continue
			}
			err = r.httpProbe(kubeletProbeURL)
			if err == nil {
				continue
			}
			failures := r.state.GetKubeletProbeFail()
			if failures >= appProbeThreshold {
				glog.Warningf("Probing failed, stopping ...")
				// display some helpers to investigate:
				glog.Infof("Investigate the kubelet logs with: journalctl -u %skubelet.service -o cat -e --no-pager", r.env.GetSystemdUnitPrefix())
				glog.Infof("Investigate the kubelet status with: systemctl status %skubelet.service -l --no-pager", r.env.GetSystemdUnitPrefix())
				// Propagate a stop
				return r.Stop(fmt.Errorf("failure threshold reached %d/%d", failures, appProbeThreshold))
			}
			r.state.IncKubeletProbeFailures()
			glog.Warningf("Kubelet probe threshold is %d/%d", failures+1, appProbeThreshold)

		case <-displayTick.C:
			r.runDisplay()

		case <-r.ApplyChan:
			if !r.state.IsReady() {
				glog.Warningf("Cannot re-apply when not ready, retry later")
				continue
			}
			// kubectl apply -f manifests-api
			err := r.applyManifests()
			if err != nil {
				glog.Errorf("Cannot apply manifests in %s", r.env.GetManifestsPathToApply())
			}

		case <-sigStopChan:
			if !r.state.IsReady() {
				glog.Warningf("Cannot re-apply when not ready, retry later")
				continue
			}
			err := r.gracefulDeleteAPIResources()
			if err != nil {
				glog.Errorf("Cannot reset API resources: %v", err)
				continue
			}
			// kubectl apply -f manifests-api
			err = r.applyManifests()
			if err != nil {
				glog.Errorf("Cannot apply manifests in %s", r.env.GetManifestsPathToApply())
			}

		case <-readinessTick.C:
			if r.state.IsReady() {
				// In case of lags during the kubectl apply
				continue
			}
			// Check if the kube-apiserver is healthy
			err := r.httpProbe("http://127.0.0.1:8080/healthz")
			if err != nil {
				r.state.SetAPIServerProbeLastError(err.Error())
				continue
			}
			// kubectl apply -f manifests-api
			err = r.applyManifests()
			if err != nil {
				// TODO do we trigger an exit at some point
				// TODO because it's almost a deadlock if the user didn't set a short --timeoutTimer
				glog.Errorf("Cannot apply manifests in %s", r.env.GetManifestsPathToApply())
				continue
			}
			err = r.checkInClusterDNS()
			if err != nil {
				continue
			}
			// Mark the current state as ready
			r.state.SetReady()
			glog.V(2).Infof("Pupernetes is ready")
			readinessTick.Stop()
		}
	}
}

func (r *Runtime) runDisplay() {
	podLogs, err := ioutil.ReadDir(setup.KubeletCRILogPath)
	if err != nil {
		glog.Errorf("Cannot read dir: %v", err)
		return
	}
	r.state.SetKubeletLogsPodRunning(len(podLogs))
	if !r.state.IsKubectlApplied() {
		return
	}
	pods, err := r.GetKubeletRunningPods()
	if err != nil {
		glog.Warningf("Cannot display the current state: %v", err)
		return
	}
	r.state.SetKubeletAPIPodRunning(len(pods))
}
