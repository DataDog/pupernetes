// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package run

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/miekg/dns"
	"os/exec"
	"strings"
	"time"
)

func (r *Runtime) applyManifests() error {
	glog.Infof("Calling kubectl apply -f %s ...", r.env.GetManifestsPathToApply())
	b, err := exec.Command(r.env.GetHyperkubePath(), "kubectl", "--kubeconfig", r.env.GetKubeconfigInsecurePath(), "apply", "-f", r.env.GetManifestsPathToApply()).CombinedOutput()
	output := string(b)
	if err != nil {
		glog.Errorf("Cannot apply manifests %v:\n%s", err, output)
		return err
	}
	glog.V(2).Infof("Successfully applied manifests:\n%s", output)
	return nil
}

func (r *Runtime) checkInClusterDNS() error {
	if r.conf.ReadinessDNSQueries == nil {
		glog.V(2).Infof("No dns query supplied for readiness condition, skipping")
		return nil
	}
	c := dns.Client{Timeout: time.Millisecond * 500}
	for _, query := range r.conf.ReadinessDNSQueries {
		if !strings.HasSuffix(query, ".") {
			// dns: domain must be fully qualified
			query = query + "."
		}
		dnsMessage := &dns.Msg{}
		dnsMessage.SetQuestion(query, dns.TypeA)
		resp, _, err := c.Exchange(dnsMessage, r.env.GetDNSClusterIP()+":53")
		if err != nil {
			glog.V(4).Infof("Cannot run DNS query: %v", err)
			// err message can be like:
			// - read udp 192.168.1.12:60449->192.168.254.2:53: i/o timeout
			// - write udp 192.168.1.12:42766->192.168.254.2:53: write: operation not permitted
			i := strings.Index(err.Error(), "->")
			if i == -1 {
				// log all messages if the basic parsing failed,
				// this is not ideal but enough for this use case
				i = 0
				glog.Warningf("DNS error: %v, this is blocking the readiness", err)
			}
			r.state.SetDNSLastError(fmt.Sprintf("query %s %s", query, err.Error()[i:]))
			return err
		}
		if len(resp.Answer) == 0 {
			r.state.SetDNSLastError("No DNS results for " + query)
			return err
		}
		var dnsResults []string
		for _, ans := range resp.Answer {
			dnsResults = append(dnsResults, strings.Replace(ans.String(), "\t", " ", -1))
		}
		glog.V(2).Infof("DNS query: %s", strings.Join(dnsResults, " "))
	}
	return nil
}
