// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"io/ioutil"
	"net/http"
	"path"

	"github.com/coreos/go-systemd/dbus"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
)

func (e *Environment) GetHyperkubePath() string {
	return e.binaryHyperkube.binaryABSPath
}

func (e *Environment) GetHostname() string {
	return e.hostname
}

func (e *Environment) GetDBUSClient() *dbus.Conn {
	return e.dbusClient
}

// TODO conf this
func (e *Environment) GetKubeletHealthzPort() int {
	return 10248
}

func (e *Environment) GetKubeconfigAuthPath() string {
	return e.kubeConfigAuthPath
}

func (e *Environment) GetKubeconfigInsecurePath() string {
	return e.kubeConfigInsecurePath
}

func (e *Environment) GetStaticPodPaths() ([]string, error) {
	files, err := ioutil.ReadDir(e.manifestStaticPodABSPath)
	if err != nil {
		glog.Errorf("Cannot read %s: %v", e.manifestStaticPodABSPath, err)
		return nil, err
	}
	var manifestABSPaths []string
	for _, elt := range files {
		eltABS := path.Join(e.manifestStaticPodABSPath, elt.Name())
		glog.V(4).Infof("Inventoring static pod manifest: %s", eltABS)
		manifestABSPaths = append(manifestABSPaths, eltABS)
	}
	glog.V(4).Infof("Found %d static pod manifests", len(manifestABSPaths))
	return manifestABSPaths, nil
}

func (e *Environment) GetKubernetesClient() *kubernetes.Clientset {
	return e.clientSet
}

func (e *Environment) IsDrainingPods() bool {
	return e.drainOptions.Pods
}

func (e *Environment) IsWaitingKubeletGC() bool {
	return e.drainOptions.KubeletGC
}

func (e *Environment) IsSkippingStop() bool {
	return e.drainOptions.None
}

func (e *Environment) IsCleaningIptables() bool {
	return e.drainOptions.Iptables
}

func (e *Environment) GetManifestsABSPathToApply() string {
	return e.manifestAPIABSPath
}

func (e *Environment) GetKubeletPodListReq() *http.Request {
	return e.podListRequest
}

func (e *Environment) GetKubeletClient() *http.Client {
	return e.kubeletClient
}

func (e *Environment) GetResolvConfPath() string {
	return path.Join(e.networkABSPath, "resolv.conf")
}

// GetPublicIP should panic if nil
func (e *Environment) GetPublicIP() string {
	return e.outboundIP.String()
}

func (e *Environment) GetSystemdUnits() []string {
	return e.systemdUnitNames
}
