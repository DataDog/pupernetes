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

	"github.com/DataDog/pupernetes/cmd/cli"
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
	glog.Infof("Generated command line documentation in %s", docDir)
}
