// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/golang/glog"
)

const (
	cniFileName       = "cni.json"
	defaultBridgeName = "cni-p8s"
)

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func pickInCIDR(cidr string, index int) (*net.IP, error) {
	IP, n, err := net.ParseCIDR(cidr)
	if err != nil {
		glog.Errorf("Cannot parse CIDR: %v", err)
		return nil, err
	}
	for i := 0; i < index; i++ {
		incIP(IP)
	}
	if !n.Contains(IP) {
		err := fmt.Errorf("ip %s not in cidr range %s", IP.String(), cidr)
		glog.Errorf("Unexpected error: %s", err)
		return nil, err
	}
	return &IP, nil
}

func (e *Environment) writeCNIConfig(c *cniConfig) error {
	cniConfPath := path.Join(e.networkConfigABSPath, cniFileName)
	f, err := os.OpenFile(cniConfPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0444)
	if err != nil {
		glog.Errorf("Cannot create %s: %v", cniConfPath, err)
		return err
	}
	b, err := json.Marshal(c)
	if err != nil {
		glog.Errorf("Cannot marshal CNI conf: %v", err)
		return err
	}
	glog.V(5).Infof("CNI config: %s", string(b))
	buf := bytes.NewBuffer(nil)
	err = json.Indent(buf, b, "", "  ")
	if err != nil {
		glog.Errorf("Cannot indent json: %v", err)
		return err
	}
	_, err = f.Write(buf.Bytes())
	if err != nil {
		glog.Errorf("Cannot write CNI conf to %s: %v", cniConfPath, err)
		return err
	}
	return nil
}

type cniConfig struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	Bridge           string `json:"bridge"`
	IsDefaultGateway bool   `json:"isDefaultGateway"`
	ForceAddress     bool   `json:"forceAddress,omitempty"`
	IPMasq           bool   `json:"ipMasq"`
	HairpinMode      bool   `json:"hairpinMode,omitempty"`
	Ipam             *ipam  `json:"ipam"`
}

type ipam struct {
	Type       string  `json:"type"`
	Subnet     string  `json:"subnet"`
	RangeStart string  `json:"rangeStart,omitempty"`
	RangeEnd   string  `json:"rangeEnd,omitempty"`
	Gateway    string  `json:"gateway,omitempty"`
	Routes     []route `json:"routes,omitempty"`
	DataDir    string  `json:"dataDir"`
}

type route struct {
	Destination string `json:"dst"`
	Gateway     string `json:"gw,omitempty"`
}

func getOutboundIP() (*net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		glog.Errorf("Cannot dial: %v", err)
		return nil, err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return &localAddr.IP, nil
}

func (e *Environment) generateResolvConf() error {
	_, err := os.Stat(e.GetResolvConfPath())
	if err == nil {
		glog.V(4).Infof("Already created: %s", e.GetResolvConfPath())
		return nil
	}
	glog.V(3).Infof("Setting %s as first nameserver", e.dnsClusterIP.String())
	nameservers := []string{e.dnsClusterIP.String()}

	discoveredNameservers, err := e.getNameservers()
	if err != nil {
		glog.Errorf("Cannot get nameservers: %v", err)
		return err
	}
	if len(discoveredNameservers) == 0 {
		glog.V(2).Infof("No nameserver discovered, adding default 8.8.8.8, 8.8.4.4") // TODO remove this
		discoveredNameservers = append(discoveredNameservers, "8.8.8.8", "8.8.4.4")
	}
	nameservers = append(nameservers, discoveredNameservers...)

	f, err := os.OpenFile(e.GetResolvConfPath(), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0444)
	if err != nil {
		return err
	}
	for _, ns := range nameservers {
		_, err = f.WriteString(fmt.Sprintf("nameserver %s\n", ns))
		if err != nil {
			glog.Errorf("Cannot write resolv.conf: %v", err)
			return err
		}
	}
	glog.V(4).Infof("Created %s", e.GetResolvConfPath())
	return nil
}

func (e *Environment) newCNIBridgeConfig(bridgeName string) *cniConfig {
	return &cniConfig{
		Name:             "p8s",
		Type:             "bridge",
		Bridge:           bridgeName,
		IsDefaultGateway: true,
		IPMasq:           true,
		Ipam: &ipam{
			Type:    "host-local",
			Subnet:  e.podCIDR.String(),
			DataDir: e.networkStateABSPath,
			Routes:  []route{{Destination: "0.0.0.0/0", Gateway: e.podBridgeGatewayIP.String()}},
		},
	}
}

func getNameserverFromSystemdOutput(b []byte) []string {
	const prefix = "DNS Servers: "
	var nameServerIPs []string

	scan := bufio.NewScanner(bytes.NewReader(b))
	for scan.Scan() {
		line := scan.Text()
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		ip := strings.TrimSpace(line[len(prefix):])
		if net.ParseIP(ip) == nil {
			glog.V(4).Infof("Invalid nameserver in systemd-resolve: %s", line)
			continue
		}
		glog.V(4).Infof("Found nameserver in systemd-resolve: %s", ip)
		nameServerIPs = append(nameServerIPs, ip)
	}
	return nameServerIPs
}

func getNameserverFromResolvConf(b []byte) []string {
	const prefix = "nameserver "
	var nameServerIPs []string

	scan := bufio.NewScanner(bytes.NewReader(b))
	for scan.Scan() {
		line := scan.Text()
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		ipString := strings.TrimSpace(line[len(prefix):])
		ip := net.ParseIP(ipString)
		if ip == nil {
			glog.V(4).Infof("Invalid nameserver in /etc/resolv.conf: %s", line)
			continue
		}
		if ip.IsLoopback() {
			glog.V(2).Infof("Skipping nameserver %s, loopback", ip.String())
			continue
		}
		glog.V(4).Infof("Found nameserver in /etc/resolv.conf: %s", ipString)
		nameServerIPs = append(nameServerIPs, ipString)
	}
	return nameServerIPs

}

func (e *Environment) getNameservers() ([]string, error) {
	b, err := exec.Command("systemd-resolve", "--status", "--no-pager").CombinedOutput()
	output := string(b)
	if err != nil {
		glog.Warningf("Cannot run systemd-resolve --status: %s, %v", output, err)
		const resolvConf = "/etc/resolv.conf"
		glog.Infof("Failling back to %s", resolvConf)
		b, err = ioutil.ReadFile(resolvConf)
		if err != nil {
			glog.Errorf("Cannot read %s: %v", resolvConf, err)
			return nil, err
		}
		return getNameserverFromResolvConf(b), err
	}
	return getNameserverFromSystemdOutput(b), nil
}

func (e *Environment) generateCNIConf(bridgeName string) error {
	c := e.newCNIBridgeConfig(bridgeName)
	err := e.writeCNIConfig(c)
	if err != nil {
		glog.Errorf("Cannot write CNI config: %v", err)
		return err
	}
	return nil
}

func (e *Environment) setupNetwork() error {
	var err error

	e.outboundIP, err = getOutboundIP()
	if err != nil {
		glog.Errorf("Cannot get outboundIP: %v", err)
		return err
	}
	glog.V(4).Infof("Outbound IP is: %v", e.outboundIP.String())

	err = e.generateResolvConf()
	if err != nil {
		return err
	}

	err = e.generateCNIConf(defaultBridgeName)
	if err != nil {
		return err
	}

	// docker set a drop by default
	iptablesRules := [][]string{
		{"iptables", "-A", "FORWARD", "--in-interface", defaultBridgeName, "-j", "ACCEPT"},
		{"iptables", "-A", "FORWARD", "--out-interface", defaultBridgeName, "-j", "ACCEPT"},
	}
	for _, rule := range iptablesRules {
		glog.V(4).Infof("Adding iptables rule: %s", strings.Join(rule, " "))
		b, err := exec.Command(rule[0], rule[1:]...).CombinedOutput()
		if err != nil {
			glog.Errorf("Cannot run %s: %s", strings.Join(rule, " "), string(b))
			return err
		}
	}
	return nil
}
