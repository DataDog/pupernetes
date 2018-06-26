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

func (e *Environment) extractEtcd() error {
	glog.V(2).Infof("Extracting %s", e.binaryEtcd.archivePath)
	b, err := exec.Command("tar", "-C", e.binABSPath, "-xzvf", e.binaryEtcd.archivePath, "--strip-components=1",
		fmt.Sprintf("etcd-v%s-linux-amd64/etcd", e.binaryEtcd.version)).CombinedOutput()
	output := string(b)
	if err != nil {
		glog.Errorf("Cannot untar %s, %s: %v", e.binaryEtcd.archivePath, output, err)
		e.binaryEtcd.removeArchive()
		return err
	}
	_, err = os.Stat(e.binaryEtcd.binaryABSPath)
	if err != nil {
		glog.Errorf("Unexpected error: %v after untar %s", err, output)
	}
	glog.V(2).Infof("Successfully untar %s: %s", e.binaryEtcd.archivePath, output)
	return err
}

func (e *Environment) setupBinaryEtcd() error {
	_, err := os.Stat(e.binaryEtcd.binaryABSPath)
	if err == nil && e.binaryEtcd.isUpToDate() {
		glog.V(4).Infof("Etcd already setup and up to date: %s", e.binaryEtcd.binaryABSPath)
		return nil
	}
	err = e.binaryEtcd.download()
	if err != nil {
		return err
	}
	err = e.extractEtcd()
	if err != nil {
		return err
	}
	if !e.binaryEtcd.isUpToDate() {
		return fmt.Errorf("etcd %s is outdated", e.binaryEtcd.binaryABSPath)
	}

	return nil
}
