// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/golang/glog"
)

func (e *Environment) extractCNI() error {
	glog.V(2).Infof("Extracting %s", e.binaryCNI.archivePath)
	b, err := exec.Command("tar", "-C", e.binABSPath, "-xzvf", e.binaryCNI.archivePath).CombinedOutput()
	output := string(b)
	if err != nil {
		glog.Errorf("Cannot untar %s, %s: %v", e.binaryCNI.archivePath, output, err)
		_ = e.binaryCNI.removeArchive()
		return err
	}
	_, err = os.Stat(e.binaryCNI.binaryABSPath)
	if err != nil {
		glog.Errorf("Unexpected error: %v after untar %s", err, output)
	}
	output = strings.Replace(output, "./", "", -1)
	output = strings.Replace(output, "\n", " ", -1)
	output = strings.TrimPrefix(output, " ")
	glog.V(2).Infof("Successfully untar %s: %s", e.binaryCNI.archivePath, output)
	return err
}

func (e *Environment) setupBinaryCNI() error {
	_, err := os.Stat(e.binaryCNI.binaryABSPath)
	if err == nil {
		glog.V(4).Infof("CNI already setup: %s", e.binaryCNI.binaryABSPath)
		return nil
	}
	err = e.binaryCNI.download()
	if err != nil {
		return err
	}
	err = e.extractCNI()
	if err != nil {
		return err
	}
	_, err = os.Stat(e.binaryCNI.binaryABSPath)
	if err != nil {
		return fmt.Errorf("cni %s is not here: %v", e.binaryCNI.binaryABSPath, err)
	}

	return nil
}
