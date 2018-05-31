// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package templates

const (
	ManifestStaticPod   = "/manifest-static-pod"
	ManifestAPI         = "/manifest-api"
	ManifestConfig      = "/manifest-config"
	ManifestSystemdUnit = "/manifest-systemd-unit"
)

type Manifest struct {
	Name        string
	Destination string
	Content     []byte
}

var Manifests map[string][]Manifest

// TODO add a layer for flavor like, http, https

func init() {
	Manifests = make(map[string][]Manifest)

	Manifests["1.11"] = manifest1o11
	Manifests["1.10"] = manifest1o10
	Manifests["1.9"] = manifest1o9
	Manifests["1.8"] = manifest1o8
	Manifests["1.7"] = manifest1o7
	Manifests["1.6"] = manifest1o6
	Manifests["1.5"] = manifest1o5
}
