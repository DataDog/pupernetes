// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package config

import (
	"github.com/spf13/viper"
	"time"
)

var ViperConfig = viper.New()

func init() {
	ViperConfig.SetDefault("hyperkube-version", "1.10.1")
	ViperConfig.SetDefault("vault-version", "0.9.5")
	ViperConfig.SetDefault("etcd-version", "3.1.11")
	ViperConfig.SetDefault("cni-version", "0.7.0")

	ViperConfig.SetDefault("kubernetes-cluster-ip-range", "192.168.254.0/24")
	ViperConfig.SetDefault("bind-address", "127.0.0.1:8989")
	ViperConfig.SetDefault("kubelet-root-dir", "/var/lib/e2e-kubelet")
	ViperConfig.SetDefault("systemd-unit-prefix", "e2e-")

	ViperConfig.SetDefault("root-path", "") // TODO not used yet

	ViperConfig.SetDefault("clean", "etcd,mounts,iptables")
	ViperConfig.SetDefault("drain", "all")
	ViperConfig.SetDefault("timeout", time.Hour*6)
	ViperConfig.SetDefault("gc", time.Second*60)

	ViperConfig.SetDefault("kubelet-cadvisor-port", 0)
}
