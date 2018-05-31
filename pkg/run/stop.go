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

	"github.com/DataDog/pupernetes/pkg/logging"
	"github.com/DataDog/pupernetes/pkg/util"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"os/exec"
	"strings"
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
	return r.state.ready
}

func (r *Runtime) gracefulDeleteAPIResources() error {
	glog.Infof("Graceful deleting API resources ...")
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
				glog.V(2).Infof("Kubelet run %d static pods", len(staticPods))
				return nil
			}
			glog.V(3).Infof("Kubelet still reports %d API Pods", len(stillRunningPods)-len(staticPods))

		case <-timeout.C:
			err := fmt.Errorf("timeout reached during delete API resources")
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

	err := r.gracefulDeleteAPIResources()
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
			err := fmt.Errorf("timeout reached during pod draining")
			glog.Errorf("Cannot properly delete pods: %v", err)
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
		glog.V(5).Infof("Issue during kube-proxy --cleanup: %s, %v", output, err)
		return err
	}
	return nil
}

func (r *Runtime) startJournalTailers() ([]*logging.JournalTailer, error) {
	failed, err := r.probeUnitStatuses()
	if err != nil && len(failed) == 0 {
		glog.Errorf("Probe units in failed: %v", err)
		return nil, err
	}
	if len(failed) == 0 {
		glog.V(2).Infof("All systemd units are healthy")
		return nil, nil
	}
	// Display the logs of the failed units
	var journalTailers []*logging.JournalTailer
	var errs []string

	for _, unitName := range failed {
		jt, err := logging.NewJournalTailer(unitName, r.runTimestamp)
		if err != nil {
			msg := fmt.Sprintf("cannot create journal tailer for %s: %v", unitName, err)
			errs = append(errs, msg)
			glog.Errorf("Unexpected error: %s", msg)
			continue
		}
		journalTailers = append(journalTailers, jt)
		err = jt.StartTail()
		if err != nil {
			msg := fmt.Sprintf("cannot start journal tailer for %s: %v", unitName, err)
			errs = append(errs, msg)
			glog.Errorf("Unexpected error: %s", msg)
		}
	}
	if len(errs) == 0 {
		return journalTailers, nil
	}
	return journalTailers, fmt.Errorf("failed to start journal tailers: %s", strings.Join(errs, ", "))
}

func stopJournalTailers(journalTailers []*logging.JournalTailer) error {
	if len(journalTailers) == 0 {
		glog.V(3).Infof("No journal tailers to stop")
		return nil
	}
	// let time to display logs
	time.Sleep(time.Second * 2)

	var errs []string
	for _, jt := range journalTailers {
		err := jt.StopTail()
		if err != nil {
			glog.Errorf("Fail to stop the journal tailer: %v", err)
			errs = append(errs, err.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("fail to stop journal tailers: %v", strings.Join(errs, ", "))
}

func (r *Runtime) Stop() error {
	if r.env.IsSkippingStop() {
		glog.Infof("Skipping stop")
		return nil
	}

	var errs []string

	err := r.drainingPods()
	if err != nil {
		glog.Errorf("Failed to drain the node: %v", err)
		errs = append(errs, err.Error())
	}

	journalTailers, err := r.startJournalTailers()
	if err != nil {
		glog.Errorf("Fail to start journalTailers: %v", err)
		errs = append(errs, err.Error())
	}
	if len(journalTailers) > 0 {
		errs = append(errs, "systemd units unhealthy")
	}

	err = stopJournalTailers(journalTailers)
	if err != nil {
		errs = append(errs, err.Error())
	}

	for _, u := range r.env.GetSystemdUnits() {
		err = util.StopUnit(r.env.GetDBUSClient(), u)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	// iptables always fail
	r.cleanIptables()
	if len(errs) == 0 {
		return nil
	}
	err = fmt.Errorf("errors during stop: %s", strings.Join(errs, ", "))
	glog.Errorf("Unexpected errors: %v", err)
	return err
}
