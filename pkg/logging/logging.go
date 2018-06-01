package logging

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/golang/glog"
)

type JournalTailer struct {
	unitName      string
	cmd           *exec.Cmd
	cmdLine       string
	stdoutPipe    io.ReadCloser
	stdoutScanner *bufio.Scanner
}

func NewJournalTailer(unitName string, since time.Time, follow bool) (*JournalTailer, error) {
	s := []string{
		"journalctl",
		"-o",
		"cat",
		"-S",
		since.Format("15:04:05"),
		"-u",
		unitName,
		"--no-pager",
	}
	if follow {
		s = append(s, "-f")
	}
	c := exec.Command(s[0], s[1:]...)

	stdoutPipe, err := c.StdoutPipe()
	if err != nil {
		glog.Errorf("Unexpected error during pipe: %v", err)
		return nil, err
	}

	return &JournalTailer{
		unitName:      unitName,
		cmd:           c,
		cmdLine:       strings.Join(s, " "),
		stdoutPipe:    stdoutPipe,
		stdoutScanner: bufio.NewScanner(bufio.NewReader(stdoutPipe)),
	}, nil
}

func (j *JournalTailer) StopTail() error {
	glog.V(3).Infof("Stopping command %s ...", j.cmdLine)
	err := j.stdoutPipe.Close()
	if err != nil {
		glog.Warningf("Unexpected error during pipe closing: %v", err)
	}
	err = j.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		glog.Errorf("Unexpected error during stopping command %s: %v", j.cmdLine, err)
		return err
	}
	j.cmd.Wait()
	glog.V(3).Infof("Journal tailing successfully stopped")
	return nil
}

func (j JournalTailer) display() {
	glog.V(4).Infof("Logging of %s started", j.unitName)
	for j.stdoutScanner.Scan() {
		fmt.Fprintf(os.Stderr, "[%s] %s\n", j.unitName, j.stdoutScanner.Text())
	}
	glog.V(4).Infof("Logging of %s stopped", j.unitName)
}

func (j *JournalTailer) StartTail() error {
	glog.V(3).Infof("Starting command %s ...", j.cmdLine)
	go j.display()
	err := j.cmd.Start()
	if err != nil {
		glog.Errorf("Cannot start the command %s: %v", j.cmdLine, err)
		return err
	}
	glog.V(3).Infof("Command %s started as pid %d", j.cmdLine, j.cmd.Process.Pid)
	return nil
}

func (j *JournalTailer) Wait() error {
	err := j.cmd.Wait()
	if err == nil {
		glog.V(2).Infof("Journal tailing of %s stopped, get them again with: %s", j.unitName, j.cmdLine)
		return nil
	}
	glog.Errorf("Journal tailing of %s stopped with unexpected error: %s", j.unitName, err)
	return err
}

func (j *JournalTailer) GetUnitName() string {
	return j.unitName
}

func (j *JournalTailer) GetCommandLine() string {
	return j.cmdLine
}
