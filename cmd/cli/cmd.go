// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package cli

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"github.com/DataDog/pupernetes/pkg/config"
	"github.com/DataDog/pupernetes/pkg/options"
	"github.com/DataDog/pupernetes/pkg/run"
	"github.com/DataDog/pupernetes/pkg/setup"
)

const programName = "pupernetes"

func NewCommand() (*cobra.Command, *int) {
	var verbose int
	var exitCode int

	rootCommand := &cobra.Command{
		Use:   fmt.Sprintf("%s testing command line", programName),
		Short: "Use this command to manage a Kubernetes testing environment",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			flag.Lookup("alsologtostderr").Value.Set("true")
			flag.Lookup("v").Value.Set(strconv.Itoa(verbose))
		},
	}

	setupCommand := &cobra.Command{
		SuggestFor: []string{"prepare"},
		Use:        "setup [directory]",
		Short:      "Setup the environment",
		Args:       cobra.ExactArgs(1), // basePathDirectory
		Example:    fmt.Sprintf("%s setup state/", programName),
		Run: func(cmd *cobra.Command, args []string) {
			env, err := setup.NewConfigSetup(args[0])
			if err != nil {
				glog.Errorf("Command returns error: %v", err)
				exitCode = 1
				return
			}
			err = env.Clean()
			if err != nil {
				glog.Errorf("Command returns error: %v", err)
				exitCode = 1
				return
			}
			err = env.Setup()
			if err != nil {
				glog.Errorf("Command returns error: %v", err)
				exitCode = 1
				return
			}
		},
	}

	runCommand := &cobra.Command{
		SuggestFor: []string{"start"},
		Use:        "run [directory]",
		Short:      fmt.Sprintf("%s and run the environment", setupCommand.Name()),
		Args:       cobra.ExactArgs(1), // basePathDirectory
		Example: fmt.Sprintf(`
# Setup and run the environment with the default options: 
%s run state/

# Clean all the environment, setup and run the environment:
%s run state/ -c all

# Clean everything but the binaries, setup and run the environment:
%s run state/ -c etcd,kubectl,kubelet,manifests,network,secrets,systemd,mounts

# Setup and run the environment with a 5 minutes timeout: 
%s run state/ --timeout 5m

# Setup and run the environment, then garantee a kubelet garbage collection during the drain phase: 
%s run state/ --gc 1m
`, programName, programName, programName, programName, programName),
		Run: func(cmd *cobra.Command, args []string) {
			env, err := setup.NewConfigSetup(args[0])
			if err != nil {
				glog.Errorf("Command returns error: %v", err)
				exitCode = 1
				return
			}
			err = env.Clean()
			if err != nil {
				glog.Errorf("Command returns error: %v", err)
				exitCode = 1
				return
			}
			err = env.Setup()
			if err != nil {
				glog.Errorf("Command returns error: %v", err)
				exitCode = 1
				return
			}
			err = run.NewRunner(env).Run()
			if err != nil {
				glog.Errorf("Command returns error: %v", err)
				exitCode = 2
				return
			}
		},
	}

	cleanCommand := &cobra.Command{
		SuggestFor: []string{"remove", "delete"},
		Use:        "clean [directory]",
		Short:      fmt.Sprintf("Clean the environment created by %s and altered by a %s", setupCommand.Name(), runCommand.Name()),
		Args:       cobra.ExactArgs(1), // basePathDirectory
		Example: fmt.Sprintf(`
# Clean the environment default:
%s clean state/

# Clean everything:
%s clean state/ -c all

# Clean the etcd data-dir, the network configuration and the secrets:
%s clean state/ -c etcd,network,secrets
`, programName, programName, programName),
		Run: func(cmd *cobra.Command, args []string) {
			env, err := setup.NewConfigSetup(args[0])
			if err != nil {
				glog.Errorf("Command returns error: %v", err)
				exitCode = 1
				return
			}
			err = env.Clean()
			if err != nil {
				glog.Errorf("Command returns error: %v", err)
				exitCode = 1
				return
			}
		},
	}

	// root
	rootCommand.PersistentFlags().IntVarP(&verbose, "verbose", "v", 2, "verbose level")

	rootCommand.PersistentFlags().String("hyperkube-version", config.ViperConfig.GetString("hyperkube-version"), "hyperkube version")
	config.ViperConfig.BindPFlag("hyperkube-version", rootCommand.PersistentFlags().Lookup("hyperkube-version"))

	rootCommand.PersistentFlags().String("vault-version", config.ViperConfig.GetString("vault-version"), "vault version")
	config.ViperConfig.BindPFlag("vault-version", rootCommand.PersistentFlags().Lookup("vault-version"))

	rootCommand.PersistentFlags().String("etcd-version", config.ViperConfig.GetString("etcd-version"), "etcd version")
	config.ViperConfig.BindPFlag("etcd-version", rootCommand.PersistentFlags().Lookup("etcd-version"))

	rootCommand.PersistentFlags().String("cni-version", config.ViperConfig.GetString("cni-version"), "container network interface (cni) version")
	config.ViperConfig.BindPFlag("cni-version", rootCommand.PersistentFlags().Lookup("cni-version"))

	rootCommand.PersistentFlags().String("kubelet-root-dir", config.ViperConfig.GetString("kubelet-root-dir"), "directory path for managing kubelet files")
	config.ViperConfig.BindPFlag("kubelet-root-dir", rootCommand.PersistentFlags().Lookup("kubelet-root-dir"))

	rootCommand.PersistentFlags().String("systemd-unit-prefix", config.ViperConfig.GetString("systemd-unit-prefix"), "prefix for systemd unit name")
	config.ViperConfig.BindPFlag("systemd-unit-prefix", rootCommand.PersistentFlags().Lookup("systemd-unit-prefix"))

	rootCommand.PersistentFlags().Int("kubelet-cadvisor-port", config.ViperConfig.GetInt("kubelet-cadvisor-port"), "enable kubelet cAdvisor on port")
	config.ViperConfig.BindPFlag("kubelet-cadvisor-port", rootCommand.PersistentFlags().Lookup("kubelet-cadvisor-port"))

	rootCommand.PersistentFlags().StringP("clean", "c", config.ViperConfig.GetString("clean"), fmt.Sprintf("clean options before %s: %s", setupCommand.Name(), options.GetOptionNames(options.Clean{})))
	config.ViperConfig.BindPFlag("clean", rootCommand.PersistentFlags().Lookup("clean"))

	// clean
	rootCommand.AddCommand(cleanCommand)

	// setup
	rootCommand.AddCommand(setupCommand)

	// run
	rootCommand.AddCommand(runCommand)

	runCommand.PersistentFlags().StringP("drain", "d", config.ViperConfig.GetString("drain"), fmt.Sprintf("drain options after %s: %s", runCommand.Name(), options.GetOptionNames(options.Drain{})))
	config.ViperConfig.BindPFlag("drain", runCommand.PersistentFlags().Lookup("drain"))

	runCommand.PersistentFlags().Duration("timeout", config.ViperConfig.GetDuration("timeout"), fmt.Sprintf("timeout for %s", runCommand.Name()))
	config.ViperConfig.BindPFlag("timeout", runCommand.PersistentFlags().Lookup("timeout"))

	runCommand.PersistentFlags().Duration("gc", config.ViperConfig.GetDuration("gc"), fmt.Sprintf("grace period for the kubelet GC trigger when draining %s, no-op if not draining", runCommand.Name()))
	config.ViperConfig.BindPFlag("gc", runCommand.PersistentFlags().Lookup("gc"))

	runCommand.PersistentFlags().String("bind-address", config.ViperConfig.GetString("bind-address"), "bind address for the API ip:port")
	config.ViperConfig.BindPFlag("bind-address", runCommand.PersistentFlags().Lookup("bind-address"))

	return rootCommand, &exitCode
}
