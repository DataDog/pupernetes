// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/golang/glog"
)

type depBinary struct {
	archivePath   string
	binaryABSPath string
	archiveURL    string
	version       string
}

type exeBinary struct {
	depBinary
	commandVersion []string
}

const downloadBinaryRetryDelay = time.Second * 5

func (x *exeBinary) isUpToDate() bool {
	b, err := exec.Command(x.binaryABSPath, x.commandVersion...).CombinedOutput()
	output := strings.Trim(string(b), "\n")
	if err != nil {
		glog.Errorf("Cannot check version for %s: %s, %v", x.binaryABSPath, output, err)
		return false
	}
	glog.V(4).Infof("Binary %s version: %s, wanted: %s", x.binaryABSPath, strings.Split(output, "\n"), x.version)
	upToDate := strings.Contains(output, x.version)
	if upToDate {
		return true
	}
	glog.V(2).Infof("Need to update the binary %s version: %s, wanted: %s", x.binaryABSPath, strings.Split(output, "\n"), x.version)
	return false
}

func (d *depBinary) downloadToFile() error {
	glog.V(2).Infof("Downloading the archive %s to %s", d.archiveURL, d.archivePath)
	client := &http.Client{Timeout: time.Minute * 15}
	resp, err := client.Get(d.archiveURL)
	if err != nil {
		glog.Errorf("Cannot download %s: %v", d.archiveURL, err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("status code: %d", resp.StatusCode)
		glog.Errorf("Cannot download %s, status code != 200, %s", d.archiveURL, err)
		return err
	}

	f, err := os.OpenFile(d.archivePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0400)
	if err != nil {
		if !os.IsNotExist(err) {
			glog.Errorf("Cannot open %s: %v", d.archivePath, err)
			return err
		}
	}
	defer f.Close()
	dst := bufio.NewWriter(f)

	_, err = io.Copy(dst, resp.Body)
	if err != nil {
		glog.Errorf("Cannot write %s in %s: %v", resp.Request.URL.RawQuery, f.Name(), err)
		return err
	}
	glog.V(2).Infof("Successfully downloaded %s to %s", d.archiveURL, d.archivePath)
	return dst.Flush()
}

func (d *depBinary) download() error {
	_, err := os.Stat(d.archivePath)
	if err == nil {
		glog.V(2).Infof("Archive already here: %s", d.archivePath)
		return nil
	}
	err = d.downloadToFile()
	if err != nil {
		glog.Errorf("Fail to download %s: %v", d.archiveURL, err)
		glog.Infof("Retrying to download in %s ...", downloadBinaryRetryDelay.String())
		time.Sleep(downloadBinaryRetryDelay)
		// we don't need to delete the file as we open it with O_TRUNC
		return d.downloadToFile()
	}
	return nil
}

func (d *depBinary) removeArchive() error {
	err := os.Remove(d.archivePath)
	if err != nil {
		glog.Infof("Cannot remove the archive %s: %v", d.archivePath, err)
		return err
	}
	glog.V(2).Infof("Removed %s", d.archivePath)
	return nil
}
