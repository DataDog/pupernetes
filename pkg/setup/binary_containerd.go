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

func (e *Environment) extractContainerd() error {
	glog.V(2).Infof("Extracting %s", e.binaryContainerd.archivePath)
	b, err := exec.Command("tar", "-C", e.binABSPath, "-xzvf", e.binaryContainerd.archivePath, "--strip-components=1").CombinedOutput()
	output := string(b)
	if err != nil {
		glog.Errorf("Cannot untar %s, %s: %v", e.binaryEtcd.archivePath, output, err)
		_ = e.binaryContainerd.removeArchive()
		return err
	}
	_, err = os.Stat(e.binaryContainerd.binaryABSPath)
	if err != nil {
		glog.Errorf("Unexpected error: %v after untar %s", err, output)
	}
	glog.V(2).Infof("Successfully untar %s: %s", e.binaryContainerd.archivePath, output)
	return err
}

func (e *Environment) setupBinaryContainerd() error {
	_, err := os.Stat(e.binaryContainerd.binaryABSPath)
	if err == nil && e.binaryContainerd.isUpToDate() {
		glog.V(4).Infof("Containerd already setup and up to date: %s", e.binaryContainerd.binaryABSPath)
		return nil
	}
	err = e.binaryContainerd.download()
	if err != nil {
		return err
	}
	err = e.extractContainerd()
	if err != nil {
		return err
	}
	if !e.binaryContainerd.isUpToDate() {
		return fmt.Errorf("containerd %s is outdated", e.binaryEtcd.binaryABSPath)
	}
	return nil
}
