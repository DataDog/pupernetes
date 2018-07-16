// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package config

import (
	"github.com/spf13/viper"
	"time"
)

// ViperConfig is a global variable for the viper configuration
// TODO move it to the cmd package and make it private see https://github.com/DataDog/pupernetes/issues/40
var ViperConfig = viper.New()

const (
	// JobTypeKey is the key for daemon types
	JobTypeKey = "job-type"

	// JobSystemd is the value to daemonise in a systemd unit .service
	JobSystemd = "systemd"

	// JobForeground is the value to daemonise the current process
	JobForeground = "fg"

	defaultAPIAddress = "127.0.0.1:8989"
)

func init() {
	ViperConfig.SetDefault("hyperkube-version", "1.10.3")
	ViperConfig.SetDefault("vault-version", "0.9.5")
	ViperConfig.SetDefault("etcd-version", "3.1.11")
	ViperConfig.SetDefault("cni-version", "0.7.0")

	ViperConfig.SetDefault("download-timeout", time.Minute*30)

	ViperConfig.SetDefault("kubernetes-cluster-ip-range", "192.168.254.0/24")
	ViperConfig.SetDefault("bind-address", defaultAPIAddress)
	ViperConfig.SetDefault("api-address", defaultAPIAddress)
	ViperConfig.SetDefault("kubelet-root-dir", "/var/lib/p8s-kubelet")
	ViperConfig.SetDefault("systemd-unit-prefix", "p8s-")

	ViperConfig.SetDefault("kubectl-link", "")
	ViperConfig.SetDefault("vault-root-token", "")

	ViperConfig.SetDefault("clean", "etcd,kubelet,logs,mounts,iptables")
	ViperConfig.SetDefault("drain", "all")
	ViperConfig.SetDefault("run-timeout", time.Hour*7)
	ViperConfig.SetDefault("skip-probes", false)
	ViperConfig.SetDefault("gc", time.Second*60)

	// The supported job-type are "fg" and "systemd"
	ViperConfig.SetDefault(JobTypeKey, JobForeground)

	ViperConfig.SetDefault("systemd-job-name", "pupernetes")

	ViperConfig.SetDefault("apply", false)

	ViperConfig.SetDefault("logging-since", time.Minute*5)
	ViperConfig.SetDefault("unit-to-watch", "pupernetes.service")
	ViperConfig.SetDefault("wait-timeout", time.Minute*15)
	ViperConfig.SetDefault("client-timeout", time.Minute*1)
	ViperConfig.SetDefault("kubeconfig-path", "")
	ViperConfig.SetDefault("dns-queries", []string{"coredns.kube-system.svc.cluster.local."})
	ViperConfig.SetDefault("dns-check", false)
}
