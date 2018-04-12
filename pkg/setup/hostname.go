// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"fmt"
	"net"
	"os"

	"github.com/golang/glog"
)

func (e *Environment) setupHostname() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	if hostname == "" {
		return fmt.Errorf("empty hostname")
	}
	if hostname == "localhost" {
		return fmt.Errorf("invalid hostname: %q", hostname)
	}
	_, err = net.LookupHost(hostname)
	if err != nil {
		return err
	}
	e.hostname = hostname

	glog.V(4).Infof("Using hostname: %q", e.hostname)
	return nil
}
