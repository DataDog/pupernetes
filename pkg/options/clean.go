// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package options

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/structs"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/sets"
)

var cleanOptions = sets.NewString(getOptionNames(&Clean{})...) // does not include "all" or "none"

// Clean is a convenient key / value bool struct to store which components
// should be cleaned.
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
	Iptables  bool `json:"iptables,omitempty"`
	Logs      bool `json:"logs,omitempty"`
}

// NewCleanOptions instantiate a new Clean from the cleanString and keepString
// The clean string is lowercase clean options comma separated like: etcd,binaries ...
// keepString takes precedence over the clean one
func NewCleanOptions(cleanString, keepString string) *Clean {
	var opts sets.String
	if keepString == "" {
		opts = newOptions(cleanString, cleanOptions)
	} else {
		keepOptions := newOptions(keepString, cleanOptions)
		opts = cleanOptions.Difference(keepOptions) // keep is the opposite of clean
	}

	glog.V(3).Infof("Clean options are %q", opts.UnsortedList())

	return &Clean{
		common:    common{opts.Has("all"), opts.Has("none")},
		Etcd:      opts.Has("etcd"),
		Binaries:  opts.Has("binaries"),
		Manifests: opts.Has("manifests"),
		Kubelet:   opts.Has("kubelet"),
		Secrets:   opts.Has("secrets"),
		Network:   opts.Has("network"),
		Systemd:   opts.Has("systemd"),
		Kubectl:   opts.Has("kubectl"),
		Mounts:    opts.Has("mounts"),
		Iptables:  opts.Has("iptables"),
		Logs:      opts.Has("logs"),
	}
}

// StringJSON represents the clean options as a JSON
func (c *Clean) StringJSON() string {
	b, err := json.Marshal(c)
	if err != nil {
		glog.Errorf("Cannot marshal the clean options: %v", err)
		return ""
	}
	return string(b)
}

// StringCLI returns the clean options as a command line representation
func (c *Clean) StringCLI() string {
	m := structs.Map(c)
	var cli []string
	for k, v := range m {
		b, ok := v.(bool)
		if ok && b {
			cli = append(cli, strings.ToLower(fmt.Sprintf("%v", k)))
		}
	}
	if len(cli) == len(m) {
		return "all"
	}
	sort.Strings(cli)
	return strings.Join(cli, ",")
}
