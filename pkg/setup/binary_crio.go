// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/golang/glog"

	"github.com/DataDog/pupernetes/pkg/config"
)

func (e *Environment) extractCrio() error {
	glog.V(2).Infof("Extracting %s", e.binaryCrio.archivePath)
	cmd := exec.Command("ar", "x", e.binaryCrio.archivePath)
	cmd.Dir = e.binABSPath
	b, err := cmd.CombinedOutput()
	output := string(b)
	if err != nil {
		glog.Errorf("Cannot extract %s, %s: %v", e.binaryCrio.archivePath, output, err)
		_ = e.binaryCrio.removeArchive()
		return err
	}
	// todo extract data.tar.xz and get `usr/bin/crio`
	b, err = exec.Command("tar", "-C", e.binABSPath, "-xJf", " data.tar.xz").CombinedOutput()
	_, err = os.Stat(e.binaryCrio.binaryABSPath)
	if err != nil {
		glog.Errorf("Unexpected error: %v after untar %s", err, output)
		return err
	}
	glog.V(2).Infof("Successfully untar %s: %s", e.binaryCrio.archivePath, output)
	return err
}

func (e *Environment) setupBinaryCrio() error {
	if e.containerRuntimeInterface != config.CRICrio {
		glog.V(2).Infof("Skipping the download of CRI-o: using %q", e.containerRuntimeInterface)
		return nil
	}
	_, err := os.Stat(e.binaryCrio.binaryABSPath)
	if err == nil && e.binaryCrio.isUpToDate() {
		glog.V(4).Infof("CRI-o already setup and up to date: %s", e.binaryCrio.binaryABSPath)
		return nil
	}
	err = e.binaryCrio.download()
	if err != nil {
		return err
	}
	err = e.extractCrio()
	if err != nil {
		return err
	}
	if !e.binaryCrio.isUpToDate() {
		return fmt.Errorf("CRI-o %s is outdated", e.binaryEtcd.binaryABSPath)
	}
	return nil
}
