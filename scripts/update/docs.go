// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package main

import (
	"flag"
	"os"
	"path"

	"github.com/golang/glog"
	"github.com/spf13/cobra/doc"

	"fmt"
	"github.com/DataDog/pupernetes/cmd/cli"
	"github.com/DataDog/pupernetes/pkg/run/state"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"os/exec"
	"sort"
	"strings"
)

func init() {
	flag.CommandLine.Parse([]string{})
	flag.Lookup("alsologtostderr").Value.Set("true")
	flag.Lookup("v").Value.Set("2")
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		glog.Exitln(err)
	}
	docDir := path.Join(cwd, "docs")
	_, err = os.Stat(docDir)
	if err != nil {
		glog.Exitf("Cannot create markdown in %s", docDir)
	}
	command, _ := cli.NewCommand()
	err = doc.GenMarkdownTree(command, docDir)
	if err != nil {
		glog.Exitln(err)
	}
	files, err := ioutil.ReadDir(docDir)
	if err != nil {
		glog.Exitln(err)
	}
	// Remove cobra footer to avoid no-op diff on docs
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !strings.HasSuffix(f.Name(), ".md") {
			continue
		}
		b, err := exec.Command("bash", "-c", fmt.Sprintf(`grep -v "^###### Auto generated by spf13/cobra on" %s`, path.Join(docDir, f.Name()))).CombinedOutput()
		if err != nil {
			glog.Exitf("%s %v", string(b), err)
		}
		n, err := os.OpenFile(path.Join(docDir, f.Name()), os.O_TRUNC|os.O_WRONLY, 0444)
		if err != nil {
			glog.Errorf("Cannot open: %v", err)
			continue
		}
		_, err = n.Write(b)
		n.Close()
		if err != nil {
			glog.Errorf("Cannot write: %v", err)
			continue
		}
		glog.Infof("Successfully generated %s", path.Join(docDir, f.Name()))
	}
	glog.Infof("Generated command line documentation in %s", docDir)

	_, err = state.NewState()
	if err != nil {
		glog.Exitf("%v", err)
	}
	metrics, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		glog.Exitf("%s", err)
	}

	var metricsToWrite []string
	for _, m := range metrics {
		metricsToWrite = append(metricsToWrite, fmt.Sprintf("%q,%q,%q\n", m.GetName(), m.GetType(), m.GetHelp()))
	}

	metricFile, err := os.OpenFile(path.Join(docDir, "metrics.csv"), os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		glog.Exitf("%s", err)
	}
	defer metricFile.Close()
	metricFile.WriteString("name,type,help\n")
	sort.Strings(metricsToWrite)
	for _, elt := range metricsToWrite {
		metricFile.WriteString(elt)
	}
	metricFile.Sync()
	glog.Infof("Generated metrics file in %s", metricFile.Name())

}
