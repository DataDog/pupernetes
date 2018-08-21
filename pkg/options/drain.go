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

	"github.com/DataDog/pupernetes/pkg/util/sets"
	"github.com/fatih/structs"
	"github.com/golang/glog"
)

var drainOptions = sets.NewString(getOptionNames(&Drain{})...) // does not include "all" or "none"

// Drain is a convenient key / value bool struct to store which components
// should be drained.
type Drain struct {
	common
	Pods      bool `json:"pods,omitempty"`
	KubeletGC bool `json:"kubeletgc,omitempty"`
	Iptables  bool `json:"iptables,omitempty"`
}

// NewDrainOptions instantiate a new Drain from the drainString
// The drain string is lowercase drain options comma separated like: pods,iptables ...
func NewDrainOptions(drainString string) *Drain {
	opts := newOptions(drainString, drainOptions)
	return &Drain{
		common:    common{opts.Has("all"), opts.Has("none")},
		Pods:      opts.Has("pods"),
		KubeletGC: opts.Has("kubeletgc"),
		Iptables:  opts.Has("iptables"),
	}
}

// StringJSON represents the clean options as a JSON
func (c *Drain) StringJSON() string {
	b, err := json.Marshal(c)
	if err != nil {
		glog.Errorf("Cannot marshal the drain options: %v", err)
		return ""
	}
	return string(b)
}

// StringCLI returns the drain options as a command line representation
func (c *Drain) StringCLI() string {
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
