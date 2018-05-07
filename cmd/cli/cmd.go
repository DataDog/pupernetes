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
	"github.com/DataDog/pupernetes/pkg/job"
	"github.com/DataDog/pupernetes/pkg/options"
	"github.com/DataDog/pupernetes/pkg/run"
	"github.com/DataDog/pupernetes/pkg/setup"
)

const programName = "pupernetes"

func NewCommand() (*cobra.Command, *int) {
	var verbose int
	var exitCode int

	rootCommand := &cobra.Command{
		Use:   fmt.Sprintf("%s command line", programName),
		Short: "Use this command to clean setup and run a Kubernetes local environment",
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

# Setup and run the environment as a systemd service:
# Get logs with "journalctl -o cat -efu %s" 
# Get status with "systemctl status %s --no-pager" 
%s run state/ --%s %s
`,
			programName,
			programName,
			programName,
			programName,
			programName,
			config.ViperConfig.GetString("systemd-job-name"), config.ViperConfig.GetString("systemd-job-name"), programName, config.JobTypeKey, config.JobSystemd,
		),
		Run: func(cmd *cobra.Command, args []string) {
			// Manage self start in systemd
			jobType := config.ViperConfig.GetString(config.JobTypeKey)
			if jobType == config.JobSystemd {
				glog.V(2).Infof("Self starting as systemd unit %s.service ...", config.ViperConfig.GetString("systemd-job-name"))
				err := job.RunSystemdJob(args[0])
				if err != nil {
					glog.Errorf("Command returns error: %v", err)
					exitCode = 1
				}
				return
			}
			if jobType != config.JobForeground {
				glog.Warningf("Invalid value for --%s=%s, continuing as %q", config.JobTypeKey, jobType, config.JobForeground)
			}

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

	rootCommand.PersistentFlags().String("kubectl-link", config.ViperConfig.GetString("kubectl-link"), "Path to create a kubectl link")
	config.ViperConfig.BindPFlag("kubectl-link", rootCommand.PersistentFlags().Lookup("kubectl-link"))

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

	runCommand.PersistentFlags().String("systemd-job-name", config.ViperConfig.GetString("systemd-job-name"), "unit name used when running as systemd service")
	config.ViperConfig.BindPFlag("systemd-job-name", runCommand.PersistentFlags().Lookup("systemd-job-name"))

	runCommand.PersistentFlags().String(config.JobTypeKey, config.ViperConfig.GetString(config.JobTypeKey), fmt.Sprintf("type of job: %s or %s", config.JobForeground, config.JobSystemd))
	config.ViperConfig.BindPFlag(config.JobTypeKey, runCommand.PersistentFlags().Lookup(config.JobTypeKey))

	runCommand.PersistentFlags().String("cloud-provider", config.ViperConfig.GetString("cloud-provider"), "cloud provider for the kubelet")
	config.ViperConfig.BindPFlag("cloud-provider", runCommand.PersistentFlags().Lookup("cloud-provider"))

	return rootCommand, &exitCode
}
