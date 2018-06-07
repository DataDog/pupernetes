// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package run

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
)

func (r *Runtime) GetKubeletPods() ([]v1.Pod, error) {
	resp, err := r.env.GetKubeletClient().Do(r.env.GetKubeletPodListReq())
	if err != nil {
		glog.Errorf("Cannot get PodList: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected non 200 status code: %d", resp.StatusCode)
		glog.Errorf("Cannot get kubelet /pods, %v", err)
		return nil, err
	}
	deco := json.NewDecoder(resp.Body)
	podList := &v1.PodList{}
	err = deco.Decode(podList)
	if err != nil {
		glog.Errorf("Unexpected error during unmarshaling PodList: %v", err)
		return nil, err
	}
	return podList.Items, nil
}

func (r *Runtime) GetKubeletRunningPods() ([]v1.Pod, error) {
	pods, err := r.GetKubeletPods()
	if err != nil {
		return nil, err
	}
	var runningPods []v1.Pod
	for _, po := range pods {
		if po.Status.Phase == "Running" {
			runningPods = append(runningPods, po)
			continue
		}
		if po.Annotations["kubernetes.io/config.source"] == "file" {
			runningPods = append(runningPods, po)
			continue
		}
	}
	return runningPods, nil
}

func (r *Runtime) GetKubeletStaticPods() ([]v1.Pod, error) {
	pods, err := r.GetKubeletPods()
	if err != nil {
		return nil, err
	}
	return r.SearchStaticPods(pods), nil
}

func (r *Runtime) SearchStaticPods(pods []v1.Pod) []v1.Pod {
	var staticPods []v1.Pod
	for _, pod := range pods {
		if pod.Annotations["kubernetes.io/config.source"] != "file" {
			glog.V(5).Infof("Skipping non static Pod %s", pod.Name)
			continue
		}
		glog.V(5).Infof("Found static pod in the PodList: %s", pod.Name)
		staticPods = append(staticPods, pod)
	}
	return staticPods
}
