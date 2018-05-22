package run

import (
	"os"

	"github.com/coreos/go-systemd/daemon"
	systemdutil "github.com/coreos/go-systemd/util"
	"github.com/golang/glog"
)

// notifySystemd call sd_notify(3) if running on systemd platform and in a systemd service
func notifySystemd() error {
	if !systemdutil.IsRunningSystemd() {
		glog.V(3).Info("Not running on systemd platform")
		return nil
	}

	inService, err := systemdutil.RunningFromSystemService()
	cgoDisabled := err == systemdutil.ErrNoCGO
	if err != nil && !cgoDisabled {
		glog.Errorf("Fail to identify if running in systemd service: %s", err)
		return err
	}
	if err == nil && !inService {
		glog.V(2).Info("Not running in systemd service, skipping the notify")
		return nil
	}

	sent, err := daemon.SdNotify(false, "READY=1")
	if err != nil {
		glog.Errorf("Failed to notify systemd for readiness: %v", err)
		return err
	}
	if cgoDisabled {
		glog.V(2).Infof("Compiled with CGO_ENABLED=0, unable to ack the systemd notify. PPID is %d", os.Getppid())
		return nil
	}
	if !sent {
		glog.Warning("Forgot to set Type=notify in systemd service file ? PPID is %d", os.Getppid())
		return nil
	}
	glog.V(2).Infof("Systemd notify sent")
	return nil
}
