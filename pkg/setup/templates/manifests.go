// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package templates

const (
	// ManifestStaticPod path for static pods
	ManifestStaticPod = "/manifest-static-pod"
	// ManifestAPI path for kube-apiserver manifests
	ManifestAPI = "/manifest-api"
	// ManifestConfig path for configuration related to Kubernetes
	// like kubeconfig, audit files, ...
	ManifestConfig = "/manifest-config"
	// ManifestSystemdUnit path where systemd units are stored
	ManifestSystemdUnit = "/manifest-systemd-unit"
)

// Manifest represent a file to be rendered in a destination
type Manifest struct {
	Name        string
	Destination string
	Content     []byte
}

// Manifests is the map catalog where all Kubernetes major.minor are stored
var Manifests map[string][]Manifest

// TODO add a layer for flavor like, http, https

func init() {
	Manifests = make(map[string][]Manifest)

	Manifests["1.13"] = manifest1o13
	Manifests["1.12"] = manifest1o12
	Manifests["1.11"] = manifest1o11
	Manifests["1.10"] = manifest1o10
	Manifests["1.9"] = manifest1o9
	Manifests["1.8"] = manifest1o8
	Manifests["1.7"] = manifest1o7
	Manifests["1.6"] = manifest1o6
	Manifests["1.5"] = manifest1o5
}
