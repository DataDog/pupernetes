// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package options

import (
	"encoding/json"
	"github.com/golang/glog"
)

type Clean struct {
	common
	Etcd      bool `json:"etcd,omitempty"`
	Binaries  bool `json:"binaries,omitempty"`
	Manifests bool `json:"manifests,omitempty"`
	Kubelet   bool `json:"kubelet,omitempty"`
	Secrets   bool `json:"secrets,omitempty"`
	Network   bool `json:"network,omitempty"`
	Systemd   bool `json:"systemd,omitempty"`
	Kubectl   bool `json:"kubectl,omitempty"`
	Mounts    bool `json:"mounts,omitempty"`
}

func NewCleanOptions(cleanString string) *Clean {
	return newOptions(cleanString, &Clean{}).(*Clean)
}

func (c *Clean) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		glog.Errorf("Cannot marshal the clean options: %v", err)
		return ""
	}
	return string(b)
}
