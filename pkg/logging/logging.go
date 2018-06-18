package logging

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/golang/glog"
)

type JournalTailer struct {
	// setup
	unitName   string
	cmdLine    []string
	cmdLineStr string

	// state
	cmd           *exec.Cmd
	stdoutPipe    io.ReadCloser
	stdoutScanner *bufio.Scanner

	running bool
	mu      sync.RWMutex
}

func NewJournalTailer(unitName string, since time.Time, follow bool) (*JournalTailer, error) {
	s := []string{
		"journalctl",
		"-o",
		"cat",
		"-u",
		unitName,
		"--no-pager",
	}

	// Manage legacy systemd
	d, err := dbus.New()
	if err != nil {
		return nil, err
	}
	_, err = d.SystemState()
	if err != nil {
		glog.Warningf("Running over an old systemd platform, cannot use journalctl --since %s", since.Format("15:04:05"))
	} else {
		s = append(s, "-S", since.Format("15:04:05"))
	}

	// tailing
	if follow {
		s = append(s, "-f")
	}

	return &JournalTailer{
		unitName:   unitName,
		cmdLine:    s,
		cmdLineStr: strings.Join(s, " "),
	}, nil
}

func (j *JournalTailer) StopTail() error {
	glog.V(3).Infof("Stopping command %s ...", j.GetCommandLine())
	err := j.stdoutPipe.Close()
	if err != nil {
		glog.Warningf("Unexpected error during pipe closing: %v", err)
	}
	err = j.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		glog.Errorf("Unexpected error during stopping command %s: %v", j.GetCommandLine(), err)
		return err
	}
	j.cmd.Wait()
	glog.V(3).Infof("Journal tailing successfully stopped")
	return nil
}

func (j *JournalTailer) IsRunning() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.running
}

func (j *JournalTailer) RestartTail() error {
	glog.V(2).Infof("Restarting Journal Tailer of %s", j.GetUnitName())
	err := j.StopTail()
	if err != nil {
		glog.Errorf("Cannot restart: %v", err)
		return err
	}
	return j.StartTail()
}

func (j *JournalTailer) display() {
	unitName := j.GetUnitName()
	glog.V(4).Infof("Logging of %s started", unitName)
	for j.stdoutScanner.Scan() {
		fmt.Fprintf(os.Stderr, "[%s] %s\n", unitName, j.stdoutScanner.Text())
	}
	glog.V(4).Infof("Logging of %s stopped", j.GetCommandLine())
	j.mu.Lock()
	defer j.mu.Unlock()
	j.running = false
}

func (j *JournalTailer) StartTail() error {
	var err error

	j.mu.Lock()
	defer j.mu.Unlock()

	glog.V(3).Infof("Starting command %s ...", j.cmdLineStr)
	j.cmd = exec.Command(j.cmdLine[0], j.cmdLine[1:]...)

	// stdout logging
	j.stdoutPipe, err = j.cmd.StdoutPipe()
	if err != nil {
		glog.Errorf("Unexpected error during pipe: %v", err)
		return err
	}
	j.stdoutScanner = bufio.NewScanner(bufio.NewReader(j.stdoutPipe))
	go j.display()

	// exec
	err = j.cmd.Start()
	if err != nil {
		glog.Errorf("Cannot start the command %s: %v", j.cmdLineStr, err)
		return err
	}

	j.running = true
	glog.V(3).Infof("Command %s started as pid %d", j.cmdLineStr, j.cmd.Process.Pid)
	return nil
}

func (j *JournalTailer) Wait() error {
	err := j.cmd.Wait()
	if err == nil {
		glog.V(2).Infof("Journal tailing of %s stopped, get them again with: %s", j.GetUnitName(), j.GetCommandLine())
		return nil
	}
	glog.Errorf("Journal tailing of %s stopped with unexpected error: %s", j.GetUnitName(), err)
	return err
}

func (j *JournalTailer) GetUnitName() string {
	return j.unitName
}

func (j *JournalTailer) GetCommandLine() string {
	return j.cmdLineStr
}
