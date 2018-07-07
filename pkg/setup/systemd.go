// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-systemd/dbus"
	unit2 "github.com/coreos/go-systemd/unit"
	"github.com/golang/glog"
)

const (
	// UnitPath is the systemd target where systemd links are created
	UnitPath             = "/run/systemd/system/"
	customSystemdSection = "X-p8s"
)

var (
	fieldsToCompare = []string{
		"ExecStart",
		//"RootPath", TODO
	}
)

func getUnitOptions(unitABSPath string) ([]*unit2.UnitOption, error) {
	f, err := os.OpenFile(unitABSPath, os.O_RDONLY, 0)
	if err != nil {
		glog.Errorf("Cannot open %v", err)
		return nil, err
	}
	defer f.Close()
	opts, err := unit2.Deserialize(f)
	if err != nil {
		glog.Errorf("Cannot deserialize %s: %v", unitABSPath, err)
		return nil, err
	}
	glog.V(4).Infof("Deserialized %s with %d items", unitABSPath, len(opts))
	return opts, nil
}

func pushUnitInMap(opts []*unit2.UnitOption) map[string]string {
	m := make(map[string]string, 2)
	for _, elt := range opts {
		if elt.Section == customSystemdSection && elt.Name == "RootPath" {
			m[elt.Name] = elt.Value
		} else if elt.Section == "Service" && elt.Name == "ExecStart" {
			m[elt.Name] = elt.Value
		}
	}
	return m
}

func (e *Environment) isUnitUpToDate(onDiskOpts, currentOpts []*unit2.UnitOption) bool {
	disk := pushUnitInMap(onDiskOpts)
	current := pushUnitInMap(currentOpts)

	for _, field := range fieldsToCompare {
		diskValue, ok := disk[field]
		if !ok {
			glog.Warningf("%s isn't in the systemd unit on disk", field)
			return false
		}
		// we supposed the current is fine
		if diskValue != current[field] {
			glog.Warningf("On disk unit is different than the generated one")
			dSplited := strings.Split(diskValue, " ")
			cSplited := strings.Split(current[field], " ")
			if len(dSplited) == len(cSplited) {
				for i := 0; i < len(dSplited); i++ {
					if dSplited[i] == cSplited[i] {
						continue
					}
					glog.Warningf("Diff disk: %q, current: %q", dSplited[i], cSplited[i])
				}
			}
			return false
		}
	}
	glog.V(4).Infof("Unit on disk matched the current one")
	return true
}

func statExecStart(opts []*unit2.UnitOption) error {
	for _, elt := range opts {
		if elt.Section != "Service" {
			continue
		}
		if elt.Name != "ExecStart" {
			continue
		}
		commandLine := strings.Split(elt.Value, " ")
		_, err := os.Stat(commandLine[0])
		// TODO maybe check if executable ?
		return err
	}
	return fmt.Errorf("cannot find ExecStart in systemd options")
}

func (e *Environment) linkSystemdUnit(unitOpt []*unit2.UnitOption, manifestUnitName, unitName string) error {
	unitABSPath := path.Join(UnitPath, unitName)
	_, err := os.Stat(unitABSPath)
	if err == nil {
		glog.V(2).Infof("Already created systemd unit: %s, untouched", unitName)

		// Validate the content
		runSystemdSystemUnit, err := getUnitOptions(unitABSPath)
		if err != nil {
			return err
		}
		if !e.isUnitUpToDate(runSystemdSystemUnit, unitOpt) {
			if e.cleanOptions.Systemd {
				err = fmt.Errorf("non uptodate systemd unit %s", unitABSPath)
				glog.Errorf("Unexpected error: %v", err)
				return err
			}
			glog.Warningf(`The already created unit %q doesn't match the generated one, used clean options are %q use instead "%s,systemd"`, unitName, e.cleanOptions.StringCLI(), e.cleanOptions.StringCLI())
		}
		err = statExecStart(runSystemdSystemUnit)
		if err != nil {
			glog.Errorf("Current ExecStart in %s unit is incorrect: %v", unitABSPath, err)
			return err
		}
		return nil
	}

	// Link
	glog.V(4).Infof("Creating systemd unit %s ...", unitName)
	err = os.Symlink(manifestUnitName, unitABSPath)
	if err != nil {
		glog.Errorf("Fail to link systemd unit %s -> %s: %v", manifestUnitName, unitABSPath, err)
		return err
	}
	glog.V(4).Infof("Successfully linked systemd unit %s -> %s", manifestUnitName, unitABSPath)
	return nil
}

func (e *Environment) createEnd2EndSection() []*unit2.UnitOption {
	return []*unit2.UnitOption{
		{
			Section: customSystemdSection,
			Name:    "RootPath",
			Value:   e.rootABSPath,
		},
		{
			Section: customSystemdSection,
			Name:    "HyperkubeVersion",
			Value:   e.binaryHyperkube.version,
		},
		{
			Section: customSystemdSection,
			Name:    "EtcdVersion",
			Value:   e.binaryEtcd.version,
		},
		{
			Section: customSystemdSection,
			Name:    "VaultVersion",
			Value:   e.binaryVault.version,
		},
		{
			Section: customSystemdSection,
			Name:    "CNIVersion",
			Value:   e.binaryCNI.version,
		},
		{
			Section: customSystemdSection,
			Name:    "ContainerdVersion",
			Value:   e.binaryContainerd.version,
		},
		{
			Section: customSystemdSection,
			Name:    "RuncVersion",
			Value:   e.binaryRunc.version,
		},
		{
			Section: customSystemdSection,
			Name:    "Timestamp",
			Value:   strconv.Itoa(int(time.Now().Unix())),
		},
	}
}

func (e *Environment) createUnitFromTemplate(unitName string) error {
	manifestUnitName := path.Join(e.manifestSystemdUnit, unitName)
	fd, err := os.OpenFile(manifestUnitName, os.O_RDONLY, 0)
	if err != nil {
		glog.Errorf("Cannot read %s: %v", manifestUnitName, err)
		return err
	}
	defer fd.Close()
	unitOptions, err := unit2.Deserialize(fd)
	if err != nil {
		glog.Errorf("Unexpected error during parsing s: %v", manifestUnitName, err)
		return err
	}
	// TODO see how to insert e.systemdEnd2EndSection
	err = e.linkSystemdUnit(unitOptions, manifestUnitName, unitName)
	if err != nil {
		return err
	}
	return nil
}

func (e *Environment) setupSystemd() error {
	conn, err := dbus.NewSystemdConnection()
	if err != nil {
		glog.Errorf("Cannot connect to dbus: %v", err)
		return err
	}
	e.dbusClient = conn

	for _, u := range e.systemdUnitNames {
		glog.V(4).Infof("Creating systemd unit %s ...", u)
		err = e.createUnitFromTemplate(u)
		if err != nil {
			return err
		}
	}

	err = conn.Reload()
	if err != nil {
		glog.Errorf("Cannot daemon-reload: %v", err)
		return err
	}
	return nil
}
