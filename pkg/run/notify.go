package run

import (
	"github.com/coreos/go-systemd/daemon"
	systemdutil "github.com/coreos/go-systemd/util"
	"github.com/golang/glog"
)

// notifySystemd call sd_notify(3) if running on systemd platform and in a systemd service
func notifySystemd() error {
	if !systemdutil.IsRunningSystemd() {
		glog.V(2).Info("Not running on systemd platform")
		return nil
	}

	inService, err := systemdutil.RunningFromSystemService()
	if err != nil {
		glog.Errorf("Fail to identify if running in systemd service: %s", err)
		return err
	}
	if !inService {
		glog.V(2).Info("Not running in systemd service, skipping the notify")
		return nil
	}

	sent, err := daemon.SdNotify(false, "READY=1")
	if err != nil {
		glog.Errorf("Failed to notify systemd for readiness: %v", err)
		return err
	}
	if !sent {
		glog.Warning("Forgot to set Type=notify in systemd service file?")
		return nil
	}
	glog.V(2).Infof("Systemd notify sent")
	return nil
}
