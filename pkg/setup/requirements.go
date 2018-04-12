// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/coreos/go-systemd/dbus"
	"github.com/golang/glog"
)

func checkCommand(command string, args ...string) error {
	err := exec.Command(command, args...).Run()
	if err != nil {
		glog.Errorf("Unexpected error on cmd %s: %v", command, err)
		return err
	}
	return nil
}

func checkRequirements() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root")
	}
	err := checkCommand("tar", "--version")
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
	err = checkCommand("openssl", "genrsa")
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
