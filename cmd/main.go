// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package main

import (
	"flag"
	"os"

	"github.com/DataDog/pupernetes/cmd/cli"
	"github.com/golang/glog"
)

func init() {
	flag.CommandLine.Parse([]string{})
}

func main() {
	command, exitCode := cli.NewCommand()
	err := command.Execute()
	if err != nil {
		glog.Exitf("Error while running the command: %v", err)
	}
	if *exitCode != 0 {
		glog.Errorf("Exiting on error: %d", *exitCode)
		os.Exit(*exitCode)
	}
}
