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

// GetHyperkubePath returns the hyperkube binary abstract path
func (e *Environment) GetHyperkubePath() string {
	return e.binaryHyperkube.binaryABSPath
}

// GetHostname returns the currently used hostname
func (e *Environment) GetHostname() string {
	return e.hostname
}

// GetDBUSClient returns a connected dbus client
func (e *Environment) GetDBUSClient() *dbus.Conn {
	return e.dbusClient
}

// GetKubeletHealthzPort returns the healthz kubelet port
// TODO conf this
func (e *Environment) GetKubeletHealthzPort() int {
	return 10248
}

// GetKubeconfigAuthPath returns the kube-config abstract path
// containing secrets to connect to the kube-apiserver
func (e *Environment) GetKubeconfigAuthPath() string {
	return e.kubeConfigAuthPath
}

// GetKubeconfigInsecurePath returns the kube-config abstract path
// to reach the kube-apiserver without any authN/Z
func (e *Environment) GetKubeconfigInsecurePath() string {
	return e.kubeConfigInsecurePath
}

// GetStaticPodPaths returns the abstract path where static pods are stored
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

// GetKubernetesClient returns a configured kubernetes client
func (e *Environment) GetKubernetesClient() *kubernetes.Clientset {
	return e.clientSet
}

// IsDrainingPods is a getter over the drainOptions
func (e *Environment) IsDrainingPods() bool {
	return e.drainOptions.Pods
}

// IsWaitingKubeletGC is a getter over the drainOptions
func (e *Environment) IsWaitingKubeletGC() bool {
	return e.drainOptions.KubeletGC
}

// IsSkippingStop is a getter over the drainOptions
func (e *Environment) IsSkippingStop() bool {
	return e.drainOptions.None
}

// IsCleaningIptables is a getter over the drainOptions
func (e *Environment) IsCleaningIptables() bool {
	return e.drainOptions.Iptables
}

// GetManifestsPathToApply returns the abstract path where Kubernetes manifests to apply are
func (e *Environment) GetManifestsPathToApply() string {
	return e.manifestAPIABSPath
}

// GetKubeletPodListReq returns an http request to query the kubelet API
func (e *Environment) GetKubeletPodListReq() *http.Request {
	return e.podListRequest
}

// GetKubeletClient returns a http client intended to query the kubelet API
func (e *Environment) GetKubeletClient() *http.Client {
	return e.kubeletClient
}

// GetResolvConfPath returns an abstract path of the resolv.conf file
func (e *Environment) GetResolvConfPath() string {
	return path.Join(e.networkConfigABSPath, "resolv.conf")
}

// GetPublicIP returns the public ip address
func (e *Environment) GetPublicIP() string {
	if e.outboundIP == nil {
		return ""
	}
	return e.outboundIP.String()
}

// GetSystemdUnits returns the systemd units to manage:
// - etcd
// - kube-apiserver
// - kubelet
func (e *Environment) GetSystemdUnits() []string {
	return e.systemdUnitNames
}

// GetSystemdUnitPrefix returns the prefix used with systemd units
func (e *Environment) GetSystemdUnitPrefix() string {
	return e.systemdUnitPrefix
}

// GetDNSClusterIP returns the dns cluster ip
func (e *Environment) GetDNSClusterIP() string {
	if e.dnsClusterIP == nil {
		return ""
	}
	return e.dnsClusterIP.String()
}
