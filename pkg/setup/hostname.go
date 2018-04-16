// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"unicode"

	"github.com/golang/glog"
)

const (
	validHostnameRegex     = `[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	invalidHostnameMessage = `a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is ` + validHostnameRegex
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

func (e *Environment) setupHostname() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	if hostname == "" {
		return fmt.Errorf("empty hostname")
	}
	if hostname == "localhost" {
		return fmt.Errorf("invalid hostname: %q", hostname)
	}
	if !isValidHostname(hostname) {
		return fmt.Errorf("invalid hostname: %q, %s", hostname, invalidHostnameMessage)
	}
	_, err = net.LookupHost(hostname)
	if err != nil {
		return err
	}
	e.hostname = hostname

	glog.V(4).Infof("Using hostname: %q", e.hostname)
	return nil
}
