package logging

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/golang/glog"
)

type JournalTailer struct {
	unitName      string
	cmd           *exec.Cmd
	cmdLine       string
	stdoutPipe    io.ReadCloser
	stdoutScanner *bufio.Scanner
}

func NewJournalTailer(unitName string) (*JournalTailer, error) {
	s := []string{
		"journalctl",
		"-o",
		"cat",
		"-fu",
		unitName,
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
	err := j.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		glog.Errorf("Unexpected error during stopping command %s: %v", j.cmdLine, err)
		return err
	}
	j.cmd.Wait()
	glog.V(3).Infof("Journal tailing successfully stopped")
	return j.stdoutPipe.Close()
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
