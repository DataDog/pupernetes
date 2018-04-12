// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package options

import (
	"encoding/json"
	"github.com/golang/glog"
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
