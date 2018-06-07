package wait

import (
	"fmt"
	"github.com/DataDog/pupernetes/pkg/logging"
	"github.com/coreos/go-systemd/dbus"
	"github.com/golang/glog"
	"strings"
	"time"
)

type Wait struct {
	systemdUnitName       string
	timeout, loggingSince time.Duration
}

func NewWaiter(systemdUnitName string, timeout, loggingSince time.Duration) *Wait {
	if !strings.HasSuffix(systemdUnitName, ".service") {
		systemdUnitName = ".service"
	}
	return &Wait{
		systemdUnitName: systemdUnitName,
		timeout:         timeout,
		loggingSince:    loggingSince,
	}
}

func getSystemdUnitState(conn *dbus.Conn, unitName string) (string, error) {
	units, err := conn.ListUnitsByNames([]string{unitName})
	if err != nil {
		glog.Errorf("Cannot list units: %v", err)
		return "", err
	}
	if len(units) == 0 {
		err := fmt.Errorf("cannot find %s systemd unit", unitName)
		glog.Errorf("Empty result: %v", err)
		return "", err
	}
	if len(units) != 1 {
		err := fmt.Errorf("invalid number of systemd unit: %d", len(units))
		glog.Errorf("Too much results: %v", err)
		return "", err
	}
	glog.V(4).Infof("%s sub state is %s", units[0].Name, units[0].SubState)
	return units[0].SubState, nil
}

func (w *Wait) Wait() error {
	conn, err := dbus.NewSystemdConnection()
	if err != nil {
		glog.Errorf("Cannot connect to dbus: %v", err)
		return err
	}
	defer conn.Close()

	ts := time.Now().Add(-w.loggingSince)
	jt, err := logging.NewJournalTailer(w.systemdUnitName, ts, true)
	if err != nil {
		glog.Errorf("Cannot start the journal tailer on %s: %v", w.systemdUnitName, err)
		return err
	}
	defer jt.StopTail()
	err = jt.StartTail()
	if err != nil {
		glog.Errorf("Cannot start journal tailing: %v", err)
		return err
	}
	timeout := time.NewTimer(w.timeout)
	defer timeout.Stop()

	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()
	for {
		select {
		case <-timeout.C:
			err := fmt.Errorf("timeout while waiting for %s systemd unit", w.systemdUnitName)
			glog.Errorf("Unexpected timeout: %v", err)
			return err

		case <-tick.C:
			state, err := getSystemdUnitState(conn, w.systemdUnitName)
			if err != nil {
				return err
			}
			if state == "running" {
				glog.Infof("Systemd unit %s is %s", w.systemdUnitName, state)
				return nil
			}
			if state == "dead" {
				err = fmt.Errorf("systemd job %s is %s", w.systemdUnitName, state)
				glog.Errorf("Unexpected state of systemd job: %v", err)
				return err
			}
		}
	}
}
