// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package requirements

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cloudfoundry/gosigar"
	"github.com/coreos/go-systemd/dbus"
	"github.com/golang/glog"
)

const (
	requiredMemory uint64 = 4e9 // 4GB
)

func checkCommand(command string, args ...string) error {
	err := exec.Command(command, args...).Run()
	if err != nil {
		glog.Errorf("Requirement failure: unexpected error on cmd %s: %v", command, err)
		return err
	}
	return nil
}

func checkResources() error {
	mem := sigar.Mem{}
	err := mem.Get()
	if err != nil {
		glog.Errorf("Unexpected error during check resources: %v", err)
		return err
	}
	glog.V(3).Infof("System has %d bytes as total memory", mem.Total)

	if mem.Total >= requiredMemory {
		return nil
	}
	err = fmt.Errorf("not enough memory: %d bytes are needed, currently %d bytes", requiredMemory, mem.Total)
	glog.Errorf("Requirement failure: %v", err)
	return err
}

// CheckRequirements returns an error if the hard coded requirements are not satisfied
// TODO configure this
func CheckRequirements() error {
	if os.Geteuid() != 0 {
		err := fmt.Errorf("must run as root")
		glog.Errorf("Requirement failure: %v", err)
		return err
	}
	err := checkResources()
	if err != nil {
		return err
	}
	err = checkCommand("tar", "--version")
	if err != nil {
		return err
	}
	err = checkCommand("unzip")
	if err != nil {
		return err
	}
	err = checkCommand("systemctl", "--version")
	if err != nil {
		return err
	}
	// this binary is required by CNI
	err = checkCommand("nsenter", "--version")
	if err != nil {
		return err
	}
	err = checkCommand("iptables", "--version")
	if err != nil {
		return err
	}
	conn, err := dbus.NewSystemdConnection()
	if err != nil {
		glog.Errorf("Cannot connect to dbus: %v", err)
		return err
	}
	defer conn.Close()
	err = checkCommand("systemd-resolve", "--status", "--no-pager")
	if err != nil {
		// TODO find a way to avoid this on old systemd platform
		glog.Warningf("Cannot use systemd as resolver, fallback to common /etc/resolv.conf: may have side-effects")
	}
	return nil
}
