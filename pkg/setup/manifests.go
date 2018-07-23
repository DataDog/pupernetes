// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/golang/glog"

	defaultTemplates "github.com/DataDog/pupernetes/pkg/setup/templates"
)

func createManifest(filePath string, content []byte) error {
	glog.V(4).Infof("Creating default template in %s", filePath)
	err := ioutil.WriteFile(filePath, content, 0444)
	if err != nil {
		glog.Errorf("Cannot create manifest file: %s")
		return err
	}
	return nil
}

func (e *Environment) populateDefaultTemplates() error {
	glog.V(4).Infof("Creating default templates, if needed ...")
	for _, manifest := range defaultTemplates.Manifests[e.templateVersion] {
		filePath := path.Join(e.manifestTemplatesABSPath, manifest.Destination, manifest.Name)
		_, err := os.Stat(filePath)
		if err == nil {
			glog.V(4).Infof("Default template already here, not creating: %s", filePath)
			continue
		}
		err = createManifest(filePath, manifest.Content)
		if err != nil {
			glog.Errorf("Cannot create manifest %s: %v", manifest.Name, err)
			return err
		}
	}
	return nil
}

func (e *Environment) renderTemplates(category string) error {
	sourceDir := path.Join(e.manifestTemplatesABSPath, category)
	files, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		glog.Errorf("Cannot list the content of: %s, %v", sourceDir, err)
		return err
	}
	b, err := json.Marshal(&e.templateMetadata)
	if err != nil {
		glog.Errorf("Cannot marshal template metadata: %v", err)
		return err
	}
	glog.V(4).Infof("Rendering templates with the following metadata: %v", string(b))
	prefix := ""
	if category == defaultTemplates.ManifestSystemdUnit {
		prefix = e.systemdUnitPrefix
		glog.V(4).Infof("Currently rendering %s with file prefix %q", category, prefix)
	}
	for _, f := range files {
		p := path.Join(sourceDir, f.Name())

		b, err := ioutil.ReadFile(p)
		if err != nil {
			glog.Errorf("Cannot read the file %s: %v", p, err)
			return err
		}
		tmpl, err := template.New(f.Name()).Parse(string(b))
		if err != nil {
			glog.Errorf("Cannot parse template from %s: %v", f.Name(), err)
			return err
		}

		destPath := path.Join(e.rootABSPath, category, prefix+f.Name())
		glog.V(4).Infof("Rendering manifest %s to %s", f.Name(), destPath)
		dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0444)
		if err != nil {
			glog.Errorf("Cannot openfile %s: %v", destPath, err)
			return err
		}
		err = tmpl.Execute(dest, e.templateMetadata)
		if err != nil {
			glog.Errorf("Cannot render template %s: %v", f.Name(), err)
			return err
		}
	}
	glog.V(4).Infof("Successfully render all templates")
	return nil
}

func (e *Environment) setupManifests() error {
	glog.V(2).Infof("Using template collection of Kubernetes %s", e.templateVersion)
	_, ok := defaultTemplates.Manifests[e.templateVersion]
	if !ok {
		err := fmt.Errorf("manifest collection for %s isn't provided", e.templateVersion)
		glog.Errorf("Cannot setup manifests: %v", err)
		return err
	}
	err := e.populateDefaultTemplates()
	if err != nil {
		return err
	}
	for _, t := range []string{
		defaultTemplates.ManifestSystemdUnit,
		defaultTemplates.ManifestStaticPod,
		defaultTemplates.ManifestConfig,
		defaultTemplates.ManifestAPI,
	} {
		err = e.renderTemplates(t)
		if err != nil {
			return err
		}
	}
	return nil
}
