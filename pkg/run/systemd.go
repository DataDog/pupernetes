// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package run

import (
	"fmt"
	"time"

	"github.com/golang/glog"
)

func executeSystemdAction(unitName string, systemdAction func(string, string, chan<- string) (int, error)) error {
	statusChan := make(chan string)
	defer close(statusChan)
	_, err := systemdAction(unitName, "replace", statusChan)
	if err != nil {
		glog.Errorf("Cannot execute systemd action on %s: %v", unitName, err)
		return err
	}
	timeout := time.After(time.Second * 5)
	for {
		select {
		case s := <-statusChan:
			glog.V(3).Infof("Status of %s job: %q", unitName, s)
			if s == "done" {
				return nil
			}
		case <-timeout:
			err := fmt.Errorf("timeout awaiting for %s unit to be done", unitName)
			glog.Errorf("%v", err)
			return err
		}
	}
}

func (r *Runtime) startUnit(unitName string) error {
	glog.Infof("Starting systemd unit: %s ...", unitName)
	return executeSystemdAction(unitName, r.env.GetDBUSClient().StartUnit)
}

func (r *Runtime) stopUnit(unitName string) error {
	glog.Infof("Stopping systemd unit: %s ...", unitName)
	return executeSystemdAction(unitName, r.env.GetDBUSClient().StopUnit)
}
