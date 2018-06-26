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

func (e *Environment) extractHyperkube() error {
	glog.V(2).Infof("Extracting %s", e.binaryHyperkube.archivePath)
	b, err := exec.Command("tar", "-C", e.binABSPath, "-xzvf", e.binaryHyperkube.archivePath, "--strip-components=3", "kubernetes/server/bin/hyperkube").CombinedOutput()
	output := string(b)
	if err != nil {
		glog.Errorf("Cannot untar %s, %s: %v", e.binaryHyperkube.archivePath, output, err)
		e.binaryHyperkube.removeArchive()
		return err
	}
	_, err = os.Stat(e.binaryHyperkube.binaryABSPath)
	if err != nil {
		glog.Errorf("Unexpected error: %v after untar %s", err, output)
	}
	glog.V(2).Infof("Successfully untar %s: %s", e.binaryHyperkube.archivePath, output)
	wd, err := os.Getwd()
	if err != nil {
		glog.Errorf("Cannot get wd: %v", err)
		return err
	}
	defer os.Chdir(wd)
	err = os.Chdir(e.binABSPath)
	if err != nil {
		glog.Errorf("Cannot chdir: %v", err)
		return err
	}
	b, err = exec.Command(e.GetHyperkubePath(), "--make-symlinks").CombinedOutput()
	output = string(b)
	if err != nil {
		glog.Warningf("Cannot generate links for hyperkube: %v %s", err, output)
		return nil
	}
	glog.V(2).Infof("Generated links for hyperkube in %s", e.binABSPath)
	return nil
}

func (e *Environment) setupBinaryHyperkube() error {
	_, err := os.Stat(e.binaryHyperkube.binaryABSPath)
	if err == nil && e.binaryHyperkube.isUpToDate() {
		glog.V(4).Infof("Hyperkube already setup and up to date: %s", e.binaryHyperkube.binaryABSPath)
		return nil
	}
	err = e.binaryHyperkube.download()
	if err != nil {
		return err
	}
	err = e.extractHyperkube()
	if err != nil {
		return err
	}
	if !e.binaryHyperkube.isUpToDate() {
		return fmt.Errorf("hyperkube %s is outdated", e.binaryHyperkube.binaryABSPath)
	}

	return nil
}
