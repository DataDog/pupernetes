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

func (e *Environment) extractRunc() error {
	glog.V(2).Infof("Copying %s", e.binaryRunc.archivePath)
	b, err := exec.Command("cp", "-v", e.binaryRunc.archivePath, e.binaryRunc.binaryABSPath).CombinedOutput()
	output := string(b)
	if err != nil {
		glog.Errorf("Cannot cp %s, %s: %v", e.binaryEtcd.archivePath, output, err)
		_ = e.binaryRunc.removeArchive()
		return err
	}
	_, err = os.Stat(e.binaryRunc.binaryABSPath)
	if err != nil {
		glog.Errorf("Unexpected error: %v after cp %s", err, output)
	}
	err = os.Chmod(e.binaryRunc.binaryABSPath, 0700)
	if err != nil {
		glog.Errorf("Cannot chmod +x %s: %v", e.binaryRunc.binaryABSPath, err)
		return err
	}
	glog.V(2).Infof("Successfully cp %s: %s", e.binaryRunc.archivePath, output)
	return err
}

func (e *Environment) setupBinaryRunc() error {
	_, err := os.Stat(e.binaryRunc.binaryABSPath)
	if err == nil && e.binaryRunc.isUpToDate() {
		glog.V(4).Infof("Runc already setup and up to date: %s", e.binaryRunc.binaryABSPath)
		return nil
	}
	err = e.binaryRunc.download()
	if err != nil {
		return err
	}
	err = e.extractRunc()
	if err != nil {
		return err
	}
	if !e.binaryRunc.isUpToDate() {
		return fmt.Errorf("runc %s is outdated", e.binaryEtcd.binaryABSPath)
	}
	return nil
}
