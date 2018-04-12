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

type State struct {
	sync.RWMutex

	apiServerHookLastError string
	apiServerHookDone      bool
	kubeletProbeFailureNb  int
	kubeletPodsRunningNb   int
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
	run.api = api.NewAPI(run.SigChan, run.DeleteAPIManifests)
	return run
}

func (r *Runtime) httpProbe(url string) error {
	resp, err := r.httpClient.Get(url)
	if err != nil {
		glog.V(5).Infof("HTTP probe %s failed: %v", url, err)
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Unexpected error when reading body of %s: %s", url, err)
		return err
	}
	content := string(b)
	defer resp.Body.Close()
	glog.V(10).Infof("%s %q", url, content)
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status code for %s: %d", url, resp.StatusCode)
		glog.V(5).Infof("HTTP probe %s failed: %v", url, err)
		return err
	}
	return nil
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

	// TODO check the state of e2e-kubelet.service few seconds after: because it doesn't use sd_notify(3)

	probeChan := time.NewTicker(time.Second * 2)
	defer probeChan.Stop()

	displayChan := time.NewTicker(time.Second * 10)
	defer displayChan.Stop()

	apiserverHookChan := time.NewTicker(time.Second * 1)
	defer apiserverHookChan.Stop()

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
			r.state.Lock()
			if r.state.kubeletProbeFailureNb < appProbeThreshold {
				r.state.kubeletProbeFailureNb++
				glog.Warningf("Kubelet probe threshold is %d/%d", r.state.kubeletProbeFailureNb, appProbeThreshold)
				r.state.Unlock()
				continue
			}
			r.state.Unlock()
			glog.Warningf("Probing failed, stopping ...")
			glog.Infof("Investigate the kubelet logs with: journalctl -u %skubelet.service -o cat -e --no-pager", config.ViperConfig.GetString("systemd-unit-prefix"))
			glog.Infof("Investigate the kubelet status with: systemctl status %skubelet.service -l --no-pager", config.ViperConfig.GetString("systemd-unit-prefix"))
			r.SigChan <- syscall.SIGTERM

		case <-displayChan.C:
			r.runDisplay()

		case <-apiserverHookChan.C:
			if r.runAPIServerHook() {
				glog.V(2).Infof("Kubernetes apiserver hooks done")
				apiserverHookChan.Stop()
			}
		}
	}
}

func (r *Runtime) runDisplay() {
	r.state.Lock()
	defer r.state.Unlock()
	if r.state.apiServerHookDone == false {
		glog.V(8).Infof("Skipping display")
		return
	}
	pods, err := r.GetKubeletRunningPods()
	if err != nil {
		glog.Warningf("Cannot runDisplay some state: %v", err)
		return
	}
	if len(pods) != r.state.kubeletPodsRunningNb {
		glog.Infof("Kubelet is running %d pods", len(pods))
		r.state.kubeletPodsRunningNb = len(pods)
	}
}

func (r *Runtime) runAPIServerHook() bool {
	r.state.Lock()
	defer r.state.Unlock()
	if r.state.apiServerHookDone {
		glog.V(7).Infof("Manifests already applied")
		return false
	}
	err := r.httpProbe(fmt.Sprintf("http://127.0.0.1:8080/healthz"))
	if err != nil {
		if r.state.apiServerHookLastError != err.Error() {
			r.state.apiServerHookLastError = err.Error()
			glog.Warningf("Kubenertes apiserver not ready yet: %v", err)
		}
		return false
	}
	// deploy trough the kube-apiserver
	err = r.applyManifests()
	if err != nil {
		// TODO do we trigger an exit at some point
		// TODO because it's almost a deadlock if the user didn't set a short --timeout
		glog.Errorf("Cannot apply manifests in %s", r.env.GetManifestsABSPathToApply())
		return false
	}
	r.state.apiServerHookDone = true
	return r.state.apiServerHookDone
}
