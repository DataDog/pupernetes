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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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
}

func NewRunner(env *setup.Environment) *Runtime {
	var zero int64 = 0
	sigChan := make(chan os.Signal, 2)

	run := &Runtime{
		env:     env,
		SigChan: sigChan,
		httpClient: &http.Client{
			Timeout: time.Millisecond * 500,
		},
		state:         &State{},
		runTimeout:    config.ViperConfig.GetDuration("timeout"),
		waitKubeletGC: config.ViperConfig.GetDuration("gc"),
		kubeDeleteOption: &v1.DeleteOptions{
			GracePeriodSeconds: &zero,
		},
	}
	signal.Notify(run.SigChan, syscall.SIGTERM, syscall.SIGINT)
	run.api = api.NewAPI(run.SigChan, run.DeleteAPIManifests, run.state.IsReady)
	return run
}

func (r *Runtime) Run() error {
	glog.Infof("Timeout for this current run is %s", r.runTimeout.String())
	timeout := time.NewTimer(r.runTimeout)

	go r.api.ListenAndServe()

	defer timeout.Stop()
	err := r.startUnit(fmt.Sprintf("%setcd.service", config.ViperConfig.GetString("systemd-unit-prefix")))
	if err != nil {
		return err
	}
	err = r.startUnit(fmt.Sprintf("%skubelet.service", config.ViperConfig.GetString("systemd-unit-prefix")))
	if err != nil {
		return err
	}

	// TODO check the state of p8s-kubelet.service few seconds after: because it doesn't use sd_notify(3)

	probeChan := time.NewTicker(time.Second * 2)
	defer probeChan.Stop()

	displayChan := time.NewTicker(time.Second * 10)
	defer displayChan.Stop()

	readinessChan := time.NewTicker(time.Second * 1)
	defer readinessChan.Stop()

	kubeletProbeURL := fmt.Sprintf("http://127.0.0.1:%d/healthz", r.env.GetKubeletHealthzPort())
	for {
		select {
		case sig := <-r.SigChan:
			glog.Warningf("Signal received: %q, propagating ...", sig.String())
			signal.Reset(syscall.SIGTERM, syscall.SIGINT)
			return r.Stop()

		case <-timeout.C:
			glog.Warningf("Timeout %s reached, stopping ...", r.runTimeout.String())
			r.SigChan <- syscall.SIGTERM

		case <-probeChan.C:
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

		case <-displayChan.C:
			r.runDisplay()

		case <-readinessChan.C:
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
				// TODO because it's almost a deadlock if the user didn't set a short --timeout
				glog.Errorf("Cannot apply manifests in %s", r.env.GetManifestsABSPathToApply())
				continue
			}
			// Mark the current state as ready
			r.state.setReady()
			glog.V(2).Infof("Pupernetes is ready")
			readinessChan.Stop()
		}
	}
}

func (r *Runtime) runDisplay() {
	if !r.state.IsReady() {
		glog.V(8).Infof("Skipping display")
		return
	}
	pods, err := r.GetKubeletRunningPods()
	if err != nil {
		glog.Warningf("Cannot runDisplay some state: %v", err)
		return
	}
	r.state.setKubeletPodRunning(len(pods))
}
