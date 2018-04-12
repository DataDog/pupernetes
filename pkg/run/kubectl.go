// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package run

import (
	"github.com/golang/glog"
	"os/exec"
)

func (r *Runtime) applyManifests() error {
	glog.Infof("Calling kubectl create -f %s ...", r.env.GetManifestsABSPathToApply())
	b, err := exec.Command(r.env.GetHyperkubePath(), "kubectl", "--kubeconfig", r.env.GetKubeconfigInsecurePath(), "apply", "-f", r.env.GetManifestsABSPathToApply()).CombinedOutput()
	output := string(b)
	if err != nil {
		glog.Errorf("Cannot apply manifests %v:\n%s", err, output)
		return err
	}
	glog.V(2).Infof("Successfully applied manifests:\n%s", output)
	return nil
}
