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
)

func (e *Environment) extractVault() error {
	glog.V(2).Infof("Extracting %s", e.binaryVault.archivePath)
	b, err := exec.Command("unzip", e.binaryVault.archivePath, "-d", e.binABSPath).CombinedOutput()
	output := string(b)
	if err != nil {
		glog.Errorf("Cannot unzip %s, %s: %v", e.binaryVault.archivePath, output, err)
		_ = e.binaryVault.removeArchive()
		return err
	}
	_, err = os.Stat(e.binaryVault.binaryABSPath)
	if err != nil {
		glog.Errorf("Unexpected error: %v after unzip %s", err, output)
	}
	glog.V(2).Infof("Successfully unzip %s: %s", e.binaryVault.archivePath, output)
	return err
}

func (e *Environment) setupBinaryVault() error {
	_, err := os.Stat(e.binaryVault.binaryABSPath)
	if err == nil && e.binaryVault.isUpToDate() {
		glog.V(4).Infof("Vault already setup and up to date: %s", e.binaryVault.binaryABSPath)
		return nil
	}
	err = e.binaryVault.download()
	if err != nil {
		return err
	}
	err = e.extractVault()
	if err != nil {
		return err
	}
	if !e.binaryVault.isUpToDate() {
		return fmt.Errorf("vault %s is outdated", e.binaryVault.binaryABSPath)
	}
	return nil
}
