// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package run

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DataDog/pupernetes/pkg/api"
	"github.com/DataDog/pupernetes/pkg/config"
	"github.com/DataDog/pupernetes/pkg/setup"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"path"
	"sort"
	"strconv"
	"strings"
)

const (
	appProbeThreshold = 10
)

type Runtime struct {
	env *setup.Environment

	api *http.Server

	SigChan          chan os.Signal
	httpClient       *http.Client
	state            *State
	runTimeout       time.Duration
	waitKubeletGC    time.Duration
	kubeDeleteOption *v1.DeleteOptions

	ApplyChan chan struct{}
}

func NewRunner(env *setup.Environment) *Runtime {
	var zero int64 = 0

	run := &Runtime{
		env:     env,
		SigChan: make(chan os.Signal, 2),
		httpClient: &http.Client{
			Timeout: time.Millisecond * 500,
		},
		state: &State{
			kubeAPIServerRestartNb: -1,
		},
		runTimeout:    config.ViperConfig.GetDuration("timeout"),
		waitKubeletGC: config.ViperConfig.GetDuration("gc"),
		kubeDeleteOption: &v1.DeleteOptions{
			GracePeriodSeconds: &zero,
		},
		ApplyChan: make(chan struct{}),
	}
	signal.Notify(run.SigChan, syscall.SIGTERM, syscall.SIGINT)
	run.api = api.NewAPI(run.SigChan, run.DeleteAPIManifests, run.state.IsReady, run.ApplyChan)
	return run
}

func (r *Runtime) Run() error {
	defer close(r.ApplyChan)

	glog.Infof("Timeout for this current run is %s", r.runTimeout.String())
	timeoutTimer := time.NewTimer(r.runTimeout)
	defer timeoutTimer.Stop()

	go r.api.ListenAndServe()

	err := r.startUnit(fmt.Sprintf("%setcd.service", config.ViperConfig.GetString("systemd-unit-prefix")))
	if err != nil {
		return err
	}
	err = r.startUnit(fmt.Sprintf("%skube-apiserver.service", config.ViperConfig.GetString("systemd-unit-prefix")))
	if err != nil {
		return err
	}
	err = r.startUnit(fmt.Sprintf("%skubelet.service", config.ViperConfig.GetString("systemd-unit-prefix")))
	if err != nil {
		return err
	}

	// TODO check the state of p8s-kubelet.service few seconds after: because it doesn't use sd_notify(3)

	probeTick := time.NewTicker(time.Second * 2)
	defer probeTick.Stop()

	displayTick := time.NewTicker(time.Second * 2)
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
			signal.Reset(syscall.SIGTERM, syscall.SIGINT)
			return r.Stop()

		case <-timeoutTimer.C:
			glog.Warningf("Timeout %s reached, stopping ...", r.runTimeout.String())
			r.SigChan <- syscall.SIGTERM

		case <-probeTick.C:
			err = r.httpProbe(kubeletProbeURL)
			if err == nil {
				continue
			}
			failures := r.state.getKubeletProbeFail()
			if failures >= appProbeThreshold {
				glog.Warningf("Probing failed, stopping ...")
				// display some helpers to investigate:
				glog.Infof("Investigate the kubelet logs with: journalctl -u %skubelet.service -o cat -e --no-pager", config.ViperConfig.GetString("systemd-unit-prefix"))
				glog.Infof("Investigate the kubelet status with: systemctl status %skubelet.service -l --no-pager", config.ViperConfig.GetString("systemd-unit-prefix"))
				// Propagate a stop
				r.SigChan <- syscall.SIGTERM
				continue
			}
			r.state.incKubeletProbeFail()
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
				glog.Errorf("Cannot apply manifests in %s", r.env.GetManifestsABSPathToApply())
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
				glog.Errorf("Cannot apply manifests in %s", r.env.GetManifestsABSPathToApply())
			}

		case <-readinessTick.C:
			if r.state.IsReady() {
				// In case of lags during the kubectl apply
				continue
			}
			// Check if the kube-apiserver is healthy
			err = r.httpProbe("http://127.0.0.1:8080/healthz")
			if err != nil {
				r.state.setAPIServerProbeLastError(err.Error())
				continue
			}
			// kubectl apply -f manifests-api
			err := r.applyManifests()
			if err != nil {
				// TODO do we trigger an exit at some point
				// TODO because it's almost a deadlock if the user didn't set a short --timeoutTimer
				glog.Errorf("Cannot apply manifests in %s", r.env.GetManifestsABSPathToApply())
				continue
			}
			// Mark the current state as ready
			r.state.setReady()
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
	r.state.setKubeletLogsPodRunning(len(podLogs))
	if !r.state.IsReady() {
		for _, pod := range podLogs {
			if !pod.IsDir() {
				continue
			}
			// Static POD id is a hash of the spec, the hash doesn't contain traditional -
			if strings.ContainsRune(pod.Name(), rune('-')) {
				continue
			}
			containers, err := ioutil.ReadDir(setup.KubeletCRILogPath + pod.Name())
			if err != nil {
				glog.Errorf("Unexpected error: %v", err)
				continue
			}
			for _, container := range containers {
				if !container.IsDir() {
					continue
				}
				containerABSPath := path.Join(setup.KubeletCRILogPath, pod.Name(), container.Name())
				logs, err := ioutil.ReadDir(containerABSPath)
				if err != nil {
					glog.Errorf("Unexpected error: %v", err)
					continue
				}
				if len(logs) == 0 {
					glog.V(2).Infof("Kubernetes apiserver not running yet")
					continue
				}
				var logFilenames []string
				for _, log := range logs {
					logFilenames = append(logFilenames, log.Name())
				}
				sort.Strings(logFilenames)
				latestLog := logFilenames[len(logFilenames)-1]
				if !strings.HasSuffix(latestLog, ".log") {
					continue
				}
				restartCount, err := strconv.Atoi(latestLog[:len(latestLog)-4])
				if err != nil {
					glog.Errorf("Cannot parse the kube-apiserver restart count: %v", err)
					continue
				}
				r.state.setKubeAPIServerRestartNb(restartCount)
			}
		}
		return
	}
	pods, err := r.GetKubeletRunningPods()
	if err != nil {
		glog.Warningf("Cannot runDisplay some state: %v", err)
		return
	}
	r.state.setKubeletAPIPodRunning(len(pods))
}
