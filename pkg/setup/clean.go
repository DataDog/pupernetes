// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"fmt"
	"github.com/DataDog/pupernetes/pkg/config"
	"github.com/golang/glog"
)

func remove(path string) error {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		glog.V(4).Infof("Not existing path: %s", path)
		return nil
	}
	glog.V(4).Infof("Removing %s ...", path)
	err = os.RemoveAll(path)
	if err != nil {
		glog.Errorf("Unexpected error during cleanup of %s: %v", path, err)
		return err
	}
	glog.Infof("Removed %s", path)
	return nil
}

func (e *Environment) cleanMounts() error {
	b, err := exec.Command("mount").CombinedOutput()
	output := string(b)
	if err != nil {
		glog.Errorf("Cannot get mount list: %v, %s", err, output)
		return err
	}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		mountPath := fields[2]
		if !strings.HasPrefix(mountPath, e.kubeletRootDir) {
			continue
		}
		glog.V(4).Infof("Removing mount %s", mountPath)
		err := syscall.Unmount(mountPath, 0)
		if err != nil {
			glog.Errorf("Cannot umount %s: %v", mountPath, err)
			continue
		}
		glog.Infof("Removed mount %s", mountPath)
	}
	return nil
}

func (e *Environment) Clean() error {
	var toRemove []string

	if e.cleanOptions.None {
		glog.V(4).Infof("Cleanup skipped")
		return nil
	}

	_, err := os.Stat(e.GetHyperkubePath())
	if e.cleanOptions.Kubectl && err != nil {
		glog.Warningf("Cannot use kubectl: %v", err)
	}
	if e.cleanOptions.Kubectl && err == nil {
		for _, elt := range [][]string{
			{"delete-context", defaultKubectlContextName},
			{"delete-cluster", defaultKubectlClusterName},
			{"unset", "current-context"},
		} {
			b, err := exec.Command(e.GetHyperkubePath(),
				"kubectl",
				"config",
				elt[0],
				elt[1],
			).CombinedOutput()
			output := string(b)
			if err != nil && !strings.Contains(output, ", not in ") {
				glog.Errorf("Cannot exec kubectl: %s", output)
			}
		}
	}
	if e.cleanOptions.Iptables && err == nil {
		// this command can fail, it's a non issue
		b, err := exec.Command(e.GetHyperkubePath(), "proxy", "--cleanup").CombinedOutput()
		glog.V(5).Infof("Iptables clean: %s, %v", string(b), err)
	}

	if e.cleanOptions.Etcd {
		toRemove = append(toRemove, e.etcdDataABSPath)
	}
	if e.cleanOptions.Manifests {
		toRemove = append(toRemove, e.manifestTemplatesABSPath, e.manifestStaticPodABSPath, e.manifestAPIABSPath, e.manifestConfigABSPath, e.manifestSystemdUnit)
	}
	if e.cleanOptions.Binaries {
		toRemove = append(toRemove, e.binABSPath)
	}
	if e.cleanOptions.Secrets {
		toRemove = append(toRemove, e.secretsABSPath)
	}
	if e.cleanOptions.Network {
		toRemove = append(toRemove, e.networkABSPath)
	}
	if e.cleanOptions.Systemd {
		toRemove = append(toRemove, path.Join(UnitPath, fmt.Sprintf("%setcd.service", config.ViperConfig.GetString("systemd-unit-prefix"))))
		toRemove = append(toRemove, path.Join(UnitPath, fmt.Sprintf("%skubelet.service", config.ViperConfig.GetString("systemd-unit-prefix"))))
	}
	if e.cleanOptions.All {
		toRemove = append(toRemove, e.rootABSPath)
	}
	if e.cleanOptions.Mounts {
		e.cleanMounts()
	}
	if e.cleanOptions.Kubelet {
		// don't do it twice
		if !e.cleanOptions.Mounts {
			e.cleanMounts()
		}
		toRemove = append(toRemove, e.kubeletRootDir)
		toRemove = append(toRemove, KubeletCRILogPath)
	}

	for _, dir := range toRemove {
		err := remove(dir)
		if err != nil {
			if dir == e.kubeletRootDir && strings.Contains(err.Error(), "device or resource busy") {
				glog.Warningf("Mounts still present in %s ?", e.kubeletRootDir)
				continue
			}
			return err
		}
	}

	glog.Infof("Cleanup successfully finished")
	return nil
}
