// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"time"
	"unicode"

	"github.com/golang/glog"
)

const (
	validHostnameRegex     = `[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	invalidHostnameMessage = `a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is ` + validHostnameRegex

	awsMetadataEndpoint = "http://169.254.169.254/latest/meta-data"
	awsLocalHostname    = awsMetadataEndpoint + "/local-hostname"
	awsPublicHostname   = awsMetadataEndpoint + "/public-hostname"
	awsLocalIpv4        = awsMetadataEndpoint + "/local-ipv4"
)

var hostnameRegex = regexp.MustCompile(validHostnameRegex)

func isValidHostname(h string) bool {
	runes := []rune(h)
	for c := 0; c < len(runes); c++ {
		if unicode.IsLetter(runes[c]) && unicode.IsUpper(runes[c]) {
			return false
		}
	}
	return hostnameRegex.MatchString(h)
}

func checkHostname(hostname string) error {
	glog.V(4).Infof("Validating hostname %q ...", hostname)
	if hostname == "" {
		glog.Errorf("Invalid empty hostname")
		return fmt.Errorf("empty hostname")
	}
	if hostname == "localhost" {
		glog.Errorf("Invalid hostname: %q", hostname)
		return fmt.Errorf("invalid hostname: %q", hostname)
	}
	if !isValidHostname(hostname) {
		glog.Errorf("Invalid hostname: %q", hostname)
		return fmt.Errorf("invalid hostname: %q, %s", hostname, invalidHostnameMessage)
	}
	_, err := net.LookupHost(hostname)
	if err == nil {
		glog.V(2).Infof("Using hostname: %q", hostname)
		return nil
	}
	glog.Errorf("Fail to lookup host: %s", err)
	return err
}

func getAWSHostname(url string) (string, error) {
	glog.V(2).Infof("Trying AWS hostname on %s ...", url)
	c := &http.Client{Timeout: time.Second}
	resp, err := c.Get(url)
	if err != nil {
		glog.Errorf("Fail to reach AWS metadata: %v", err)
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected status code on %s: %d", url, resp.StatusCode)
		glog.Errorf("Cannot GET AWS endpoint: %v", err)
		return "", err
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Cannot read the AWS metadata response: %v", err)
		return "", err
	}
	hostname := string(b)
	return hostname, checkHostname(hostname)
}

func (e *Environment) setupHostname() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	if checkHostname(hostname) == nil {
		e.hostname = hostname
		return nil
	}

	for _, ep := range []string{
		awsLocalHostname,
		awsPublicHostname,
		awsLocalIpv4,
	} {
		hostname, err = getAWSHostname(ep)
		if err == nil {
			e.hostname = hostname
			return nil
		}
	}
	return err
}
