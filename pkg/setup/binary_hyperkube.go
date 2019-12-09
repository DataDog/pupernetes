// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/Masterminds/semver"

	"github.com/golang/glog"
)

func (e *Environment) extractHyperkube() error {
	glog.V(2).Infof("Extracting %s", e.binaryHyperkube.archivePath)
	b, err := exec.Command("tar", "-C", e.binABSPath, "-xzvf", e.binaryHyperkube.archivePath, "--strip-components=3").CombinedOutput()
	output := string(b)
	if err != nil {
		glog.Errorf("Cannot untar %s, %s: %v", e.binaryHyperkube.archivePath, output, err)
		_ = e.binaryHyperkube.removeArchive()
		return err
	}

	// After 1.17 hyperkube binary has been removed from Kubernetes
	// It's replaced by a shell script forwarding calls to each binary
	hyperkubeMissingCheck, err := semver.NewConstraint(">=1.17.0-rc")
	if err != nil {
		return err
	}

	if hyperkubeMissingCheck.Check(e.kubeVersion) {
		f, err := os.OpenFile(e.binaryHyperkube.binaryABSPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			glog.Errorf("Cannot write shell hyperkube replacement: %v", err)
			return err
		}

		hyperkubeTpl, err := template.New("hyperkubeTemplate").Parse(hyperkubeShellTpl)
		if err != nil {
			glog.Errorf("Cannot generate hyperkube template replacement: %v", err)
			return err
		}

		tplReplace := map[string]string{
			"binABSPath": e.binABSPath,
		}

		err = hyperkubeTpl.Execute(f, tplReplace)
		if err != nil {
			glog.Errorf("Error while generating hyperkube shell: %v", err)
			return err
		}

		defer f.Close()
	}

	_, err = os.Stat(e.binaryHyperkube.binaryABSPath)
	if err != nil {
		glog.Errorf("Unexpected error: %v after untar %s", err, output)
		return err
	}
	glog.V(2).Infof("Successfully untar %s: %s", e.binaryHyperkube.archivePath, output)
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
