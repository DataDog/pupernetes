// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package util

import (
	"fmt"
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/golang/glog"
	"strings"
)

func executeSystemdAction(unitName string, systemdAction func(string, string, chan<- string) (int, error)) error {
	statusChan := make(chan string)
	defer close(statusChan)
	_, err := systemdAction(unitName, "replace", statusChan)
	if err != nil {
		glog.Errorf("Cannot execute systemd action on %s: %v", unitName, err)
		return err
	}
	timeout := time.After(time.Minute * 1)
	for {
		select {
		case s := <-statusChan:
			glog.V(3).Infof("Status of %s job: %q", unitName, s)
			if s == "done" {
				return nil
			}
		case <-timeout:
			err := fmt.Errorf("timeout awaiting for %s unit to be done", unitName)
			glog.Errorf("Unexpected error: %v", err)
			return err
		}
	}
}

func StartUnit(d *dbus.Conn, unitName string) error {
	glog.Infof("Starting systemd unit: %s ...", unitName)
	return executeSystemdAction(unitName, d.StartUnit)
}

func StopUnit(d *dbus.Conn, unitName string) error {
	glog.Infof("Stopping systemd unit: %s ...", unitName)
	return executeSystemdAction(unitName, d.StopUnit)
}

func GetUnitStates(d *dbus.Conn, unitNames []string) ([]dbus.UnitStatus, error) {
	var units []dbus.UnitStatus

	allUnits, err := d.ListUnits()
	if err != nil {
		glog.Errorf("Cannot ListUnits: %v", err)
		return nil, err
	}
	intersect := make(map[string]dbus.UnitStatus)
	for _, elt := range allUnits {
		if !strings.HasSuffix(elt.Name, ".service") {
			continue
		}
		intersect[elt.Name] = elt
	}
	for _, wantedUnit := range unitNames {
		unit, ok := intersect[wantedUnit]
		if !ok {
			glog.V(2).Infof("cannot find %s in actual running units", wantedUnit)
			continue
		}
		glog.V(4).Infof("Found wanted unit %s", wantedUnit)
		units = append(units, unit)
	}
	return units, nil
}
