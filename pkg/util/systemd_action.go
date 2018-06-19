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

const systemdStatusTimeout = time.Minute * 1

type sytemdAction struct {
	unitName      string
	systemdAction func(string, string, chan<- string) (int, error)
	dbusConn      *dbus.Conn

	// legacy
	expectedSubState []string
	getUnitStates    func(d *dbus.Conn, unitNames []string) ([]dbus.UnitStatus, error)
}

func (sd *sytemdAction) legacySystemdPoller(statusChan chan string) error {
	unitNames := []string{sd.unitName}

	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()

	timeout := time.After(systemdStatusTimeout)
	glog.V(2).Infof("Starting legacy-poll status of %s for %s ...", sd.unitName, strings.Join(sd.expectedSubState, ", "))
	for {
		select {
		case s := <-statusChan:
			err := fmt.Errorf("shouldn't received a systemd status: %q", s)
			glog.Errorf("Polling in legacy mode: %v", err)
			return err

		case <-ticker.C:
			unitStates, err := sd.getUnitStates(sd.dbusConn, unitNames)
			if err != nil {
				return err
			}
			if len(unitStates) == 0 {
				return nil
			}
			u := unitStates[0]
			for _, ex := range sd.expectedSubState {
				if u.SubState == ex {
					return nil
				}
			}
			glog.V(3).Infof("Status of %s is %s", sd.unitName, u.SubState)

		case <-timeout:
			err := fmt.Errorf("timeout awaiting for %s unit to be done", sd.unitName)
			glog.Errorf("Unexpected error: %v", err)
			return err
		}
	}
}

func (sd *sytemdAction) executeSystemdAction() error {
	statusChan := make(chan string)
	defer close(statusChan)

	_, err := sd.systemdAction(sd.unitName, "replace", statusChan)
	if err != nil {
		glog.Errorf("Cannot execute systemd action on %s: %v", sd.unitName, err)
		return err
	}
	_, err = sd.dbusConn.SystemState()
	if err != nil {
		glog.Warningf("Running over an old systemd platform: %v, fallback to a custom status poller ...", err)
		return sd.legacySystemdPoller(statusChan)
	}

	timeout := time.After(systemdStatusTimeout)
	for {
		select {
		case s := <-statusChan:
			glog.V(3).Infof("Status of %s job: %q", sd.unitName, s)
			if s == "done" {
				return nil
			}
		case <-timeout:
			err = fmt.Errorf("timeout awaiting for %s unit to be done", sd.unitName)
			glog.Errorf("Unexpected error: %v", err)
			return err
		}
	}
}

// StartUnit call dbus to start the given unit name
func StartUnit(d *dbus.Conn, unitName string) error {
	sd := &sytemdAction{
		unitName:      unitName,
		dbusConn:      d,
		systemdAction: d.StartUnit,
		// legacy
		expectedSubState: []string{"running"},
		getUnitStates:    MustGetUnitStates,
	}
	return sd.executeSystemdAction()
}

// StopUnit call dbus to stop the given unit name
func StopUnit(d *dbus.Conn, unitName string) error {
	sd := &sytemdAction{
		unitName:      unitName,
		dbusConn:      d,
		systemdAction: d.StopUnit,
		// legacy
		expectedSubState: []string{"dead", "failed"},
		getUnitStates:    GetUnitStates,
	}
	return sd.executeSystemdAction()
}

// GetUnitStates returns the dbus UnitStates of unit names passed in parameter
func GetUnitStates(d *dbus.Conn, unitNames []string) ([]dbus.UnitStatus, error) {
	var units []dbus.UnitStatus

	allUnits, err := d.ListUnits()
	if err != nil {
		glog.Errorf("Cannot ListUnits: %v", err)
		return nil, err
	}
	intersectName := make(map[string][]dbus.UnitStatus)
	for _, elt := range allUnits {
		if !strings.HasSuffix(elt.Name, ".service") {
			continue
		}
		// Note that units may be known by multiple names at the same time
		intersectName[elt.Name] = append(intersectName[elt.Name], elt)
	}
	for _, wantedUnit := range unitNames {
		unitStatuses, ok := intersectName[wantedUnit]
		if !ok {
			glog.V(2).Infof("cannot find %s in actual running units", wantedUnit)
			continue
		}
		glog.V(4).Infof("Found wanted unitStatus %s", wantedUnit)
		units = append(units, unitStatuses[0])
	}
	return units, nil
}

// MustGetUnitStates returns the dbus UnitStates of unit names passed in parameter,
// an error is returned if the UnitStatus return more or less than asked
func MustGetUnitStates(d *dbus.Conn, unitNames []string) ([]dbus.UnitStatus, error) {
	units, err := GetUnitStates(d, unitNames)
	if err != nil {
		return nil, err
	}
	if len(units) != len(unitNames) {
		err := fmt.Errorf("invalid number of unitStatuses %d, wanted %d", len(units), len(unitNames))
		glog.Errorf("Cannot find units: %s %v", strings.Join(unitNames, ","), err)
		return nil, err
	}
	return units, nil
}
