// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package options

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"github.com/golang/glog"
	"sort"
	"strings"
)

type Drain struct {
	common
	Pods      bool `json:"pods,omitempty"`
	KubeletGC bool `json:"kubeletgc,omitempty"`
	Iptables  bool `json:"iptables,omitempty"`
}

func NewDrainOptions(drainString string) *Drain {
	return newOptions(drainString, &Drain{}).(*Drain)
}

func (c *Drain) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		glog.Errorf("Cannot marshal the drain options: %v", err)
		return ""
	}
	return string(b)
}

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
