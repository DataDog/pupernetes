package job

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	dbus2 "github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/unit"
	"github.com/golang/glog"

	"github.com/DataDog/pupernetes/pkg/config"
	"github.com/DataDog/pupernetes/pkg/logging"
)

const (
	unitPath = "/run/systemd/system/"
)

func createExecStart(givenRootPath string, argv []string, wd string) (string, error) {
	copyArgv := make([]string, len(argv))
	copy(copyArgv, argv)
	if !path.IsAbs(copyArgv[0]) {
		copyArgv[0] = path.Join(wd, copyArgv[0])
	}

	for i := 0; i < len(copyArgv); i++ {
		if copyArgv[i] == givenRootPath && !path.IsAbs(givenRootPath) {
			copyArgv[i] = path.Join(wd, givenRootPath)
		}
		// replace --job-type=systemd by --job-type=fg
		if copyArgv[i] == fmt.Sprintf("--%s=%s", config.JobTypeKey, config.JobSystemd) {
			copyArgv[i] = strings.Replace(copyArgv[i], config.JobSystemd, config.JobForeground, -1)
			break
		}
		// replace --job-type systemd by --job-type fg
		if copyArgv[i] != "--"+config.JobTypeKey {
			continue
		}
		if copyArgv[i+1] != config.JobSystemd {
			err := fmt.Errorf("invalid command line: %s", strings.Join(copyArgv, " "))
			glog.Errorf("Unexpected error: %s", err)
			return "", err
		}
		copyArgv[i+1] = config.JobForeground
	}
	return strings.Join(copyArgv, " "), nil
}

// RunSystemdJob creates and starts a systemd unit with the current command line as ExecStart
func RunSystemdJob(givenRootPath string) error {
	dbus, err := dbus2.NewSystemdConnection()
	if err != nil {
		glog.Errorf("Cannot connect to dbus: %v", err)
		return err
	}
	defer dbus.Close()

	unitName := config.ViperConfig.GetString("systemd-job-name")
	if !strings.HasSuffix(unitName, ".service") {
		unitName = unitName + ".service"
	}
	units, err := dbus.ListUnitsByNames([]string{unitName})
	if err != nil {
		glog.Errorf("Cannot get the status of %s: %v", unitName, err)
		return err
	}
	for _, u := range units {
		glog.V(3).Infof("Unit %q with load state %q is %q", u.Name, u.LoadState, u.SubState)
		switch u.SubState {
		case "running":
			err = fmt.Errorf("%s is already running", u.Name)
			glog.Warningf("Nothing to do: %s is already running: systemctl status %s --full", u.Name, u.Name)
			return nil
		case "start":
			err = fmt.Errorf("%s is already starting", u.Name)
			glog.Warningf("Nothing to do: %v: systemctl status %s --full", err, u.Name)
			return err
		case "stop-sigterm":
			err = fmt.Errorf("%s is stopping stop-sigterm", u.Name)
			glog.Warningf("Please retry later: %v: systemctl status %s --full", err, u.Name)
			return err
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		glog.Errorf("Unexpected error during get current working directory: %v", err)
		return err
	}

	execStart, err := createExecStart(givenRootPath, os.Args, wd)
	if err != nil {
		glog.Errorf("Cannot create ExecStart command: %v", err)
		return err
	}
	glog.V(2).Infof("Creating an unit with ExecStart=%s", execStart)
	sdP8s := []*unit.UnitOption{
		{
			Section: "Unit",
			Name:    "Description",
			Value:   "github.com/DataDog/pupernetes",
		},
		{
			Section: "Unit",
			Name:    "After",
			Value:   "network.target",
		},
		{
			Section: "Service",
			Name:    "ExecStart",
			Value:   execStart,
		},
		{
			Section: "Service",
			Name:    "Type",
			Value:   "notify",
		},
		{
			Section: "Service",
			Name:    "Restart",
			Value:   "no",
		},
	}

	// Write the unit on disk
	unitABSPath := path.Join(unitPath, unitName)
	glog.V(2).Infof("Creating systemd unit %s ...", unitName)
	c := unit.Serialize(sdP8s)
	b, err := ioutil.ReadAll(c)
	if err != nil {
		glog.Errorf("Cannot create systemd unit: %v", err)
		return err
	}
	err = ioutil.WriteFile(unitABSPath, b, 0444)
	if err != nil {
		glog.Errorf("Fail to write systemd unit %s: %v", unitABSPath, err)
		return err
	}
	glog.V(2).Infof("Successfully wrote systemd unit %s", unitABSPath)

	err = dbus.Reload()
	if err != nil {
		glog.Errorf("Failed to daemon-reload: %v", err)
		return err
	}

	jt, err := logging.NewJournalTailer(unitName, time.Now())
	if err != nil {
		glog.Errorf("Unexpected error while creating journal-tail of %s: %v", unitName, err)
		return err
	}
	err = jt.StartTail()
	if err != nil {
		glog.Errorf("Cannot start the journal-tail of %s: %v", unitName, err)
		return err
	}

	// Start the unit
	statusChan := make(chan string)
	defer close(statusChan)
	_, err = dbus.StartUnit(unitName, "replace", statusChan)
	if err != nil {
		glog.Errorf("Cannot start %s: %v", unitName, err)
		return err
	}

	// Poll the status of the started unit
	timeout := time.After(time.Minute * 5)

	sigChan := make(chan os.Signal, 2)
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	displayChan := time.NewTicker(5 * time.Second)
	defer displayChan.Stop()

	glog.V(2).Infof("Polling the status of %s ... SIGTERM or SIGINT to interrupt", unitName)
	for {
		select {
		case s := <-statusChan:
			glog.V(2).Infof("Status of %s job: %q", unitName, s)
			if s != "done" {
				continue
			}
			return jt.StopTail()

		case <-timeout:
			err = fmt.Errorf("timeout awaiting for %s unit to be done", unitName)
			tearDownErr := jt.StopTail()
			if tearDownErr != nil {
				err = fmt.Errorf("%s + %s", err, tearDownErr)
			}
			return err

		case <-sigChan:
			glog.V(2).Infof("Stop polling for the status of %s", unitName)

			return jt.StopTail()

		case <-displayChan.C:
			glog.V(3).Infof("Still polling the status of %s ...", unitName)
		}
	}
}
