// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package run

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DataDog/pupernetes/pkg/config"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"os/exec"
)

func (r *Runtime) getNamespaces() (*corev1.NamespaceList, error) {
	ns, err := r.env.GetKubernetesClient().CoreV1().Namespaces().List(v1.ListOptions{})
	if err != nil {
		glog.Errorf("Unexpected error during get namespaces: %v", err)
		return nil, err
	}
	glog.V(4).Infof("Listed %d namespaces", len(ns.Items))
	return ns, nil
}

func (r *Runtime) isAPIServerHookDone() bool {
	r.state.RLock()
	defer r.state.RUnlock()
	return r.state.apiServerHookDone
}

func (r *Runtime) gracefulDeleteManifests() error {
	glog.Infof("Graceful deleting API manifests ...")
	ns, err := r.getNamespaces()
	if err == nil {
		r.DeleteAPIManifests(ns)
	}

	stateTicker := time.NewTicker(1 * time.Second)
	defer stateTicker.Stop()
	timeout := time.NewTimer(15 * time.Second)
	defer timeout.Stop()
	for {
		select {
		case <-stateTicker.C:
			stillRunningPods, err := r.GetKubeletRunningPods()
			if err != nil {
				if r.isAPIServerHookDone() {
					continue
				}
				err = fmt.Errorf("cannot poll pods, RBAC may not deployed")
				glog.Errorf("Stop called too early: %v", err)
				return err
			}
			staticPods := r.SearchStaticPods(stillRunningPods)
			if len(staticPods) == len(stillRunningPods) {
				glog.V(2).Infof("Kubelet run only %d static pods", len(staticPods))
				return nil
			}
			glog.V(3).Infof("Kubelet still reports %d API Pods", len(stillRunningPods)-len(staticPods))

		case <-timeout.C:
			err := fmt.Errorf("timeout reached during delete manifests")
			glog.Errorf("Unexpected %v", err)
			return err
		}
	}
}

// TODO maybe see how long it is to use kubectl drain (API or exec) (but drain command is limited - ignore daemonsets)
func (r *Runtime) drainingPods() error {
	if !r.env.IsDrainingPods() {
		glog.Infof("Skipping the draining pod phase")
		return nil
	}
	glog.Infof("Draining kubelet's pods ...")

	err := r.gracefulDeleteManifests()
	if err != nil {
		glog.Warningf("Failed to handle a graceful delete of API resources: %v", err)
	}

	stillRunningPods, err := r.GetKubeletStaticPods()
	if err != nil {
		glog.Warningf("Cannot get the static pod still running: %v", err)
	} else {
		glog.V(4).Infof("%d static pods are running before stopping the kubelet", len(stillRunningPods))
	}

	staticPodPaths, err := r.env.GetStaticPodPaths()
	if err != nil {
		glog.Errorf("Cannot get static pod paths: %v", err)
		return err
	}
	for _, absPath := range staticPodPaths {
		err = os.Remove(absPath)
		if err != nil {
			glog.Warningf("Unexpected error during rm %s: %v", absPath, err)
		}
		glog.V(4).Infof("Removed %s", absPath)
	}

	stateTicker := time.NewTicker(1 * time.Second)
	defer stateTicker.Stop()
	timeoutDelay := 15 * time.Second
	if !r.isAPIServerHookDone() {
		timeoutDelay = timeoutDelay / 3
		glog.Warningf("APIserver hooks aren't deployed, RBAC-less ? Lowering the timeout to %s for static pods polling", timeoutDelay.String())
	}
	timeout := time.NewTimer(timeoutDelay)
	defer timeout.Stop()
	for {
		select {
		case <-stateTicker.C:
			remainStaticPods, err := r.GetKubeletStaticPods()
			if err != nil {
				continue
			}
			if len(remainStaticPods) == 0 {
				glog.V(2).Infof("Kubelet doesn't run any pod, waiting %s for the kubelet's gc or SIGINT", r.waitKubeletGC.String())
				sigChan := make(chan os.Signal)
				signal.Notify(sigChan, syscall.SIGINT)
				select {
				case <-sigChan:
					glog.Info("Skipping the GC period, want to garbage collect manually ?")
				case <-time.After(r.waitKubeletGC):
					glog.Info("GC period reached")
				}
				close(sigChan)
				signal.Reset(syscall.SIGINT)
				return nil
			}
			glog.V(2).Infof("Kubelet still have static pods running: %d", len(remainStaticPods))
		case <-timeout.C:
			err := fmt.Errorf("timeout reached during kubelet stop")
			glog.Errorf("Cannot properly delete static pods: %v", err)
			return err
		}
	}
}

func (r *Runtime) cleanIptables() error {
	if !r.env.IsCleaningIptables() {
		glog.Infof("Skipping iptables clean")
		return nil
	}
	b, err := exec.Command(r.env.GetHyperkubePath(), "proxy", "--cleanup").CombinedOutput()
	output := string(b)
	if err != nil {
		glog.V(4).Infof("Issue during kube-proxy --cleanup: %s, %v", output, err)
		return err
	}
	return nil
}

func (r *Runtime) Stop() error {
	if r.env.IsSkippingStop() {
		glog.Infof("Skipping stop")
		return nil
	}

	err := r.drainingPods()
	if err != nil {
		glog.Errorf("Failed to stop kubelet: %v", err)
	}

	err = r.stopUnit(fmt.Sprintf("%skubelet.service", config.ViperConfig.GetString("systemd-unit-prefix")))
	if err != nil {
		return err
	}

	err = r.stopUnit(fmt.Sprintf("%setcd.service", config.ViperConfig.GetString("systemd-unit-prefix")))
	if err != nil {
		return err
	}

	// ignore any error here
	r.cleanIptables()
	return nil
}
