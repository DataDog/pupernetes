// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
)

func getHome() string {
	user := os.Getenv("SUDO_USER")
	if user == "" {
		return os.Getenv("HOME")
	}
	return path.Join("/home", user)
}

func (e *Environment) setupKubectl() error {
	kubeDirPath := path.Join(getHome(), ".kube")
	kubeConfigPath := path.Join(kubeDirPath, "config")
	glog.V(4).Infof("Building kubectl configuration in %s ...", kubeConfigPath)

	// cluster
	b, err := exec.Command(e.GetHyperkubePath(),
		"kubectl",
		"config",
		"--kubeconfig="+kubeConfigPath,
		"set-cluster",
		defaultKubectlClusterName,
		"--server=https://127.0.0.1:6443",
		"--certificate-authority="+path.Join(e.secretsABSPath, "kubernetes.issuing_ca"),
	).CombinedOutput()
	output := string(b)
	if err != nil {
		glog.Errorf("Cannot set kubectl config/set-cluster: %v, %s", err, output)
		return err
	}
	glog.V(4).Infof("kubectl config/set-cluster: %s", output)

	// user
	b, err = exec.Command(e.GetHyperkubePath(),
		"kubectl",
		"config",
		"--kubeconfig="+kubeConfigPath,
		"set-credentials",
		defaultKubectlUserName,
		"--username="+defaultKubectlUserName,
		"--client-certificate="+path.Join(e.secretsABSPath, "kubernetes.certificate"),
		"--client-key="+path.Join(e.secretsABSPath, "kubernetes.private_key"),
	).CombinedOutput()
	output = string(b)
	if err != nil {
		glog.Errorf("Cannot set kubectl set-credentials: %v, %s", err, output)
		return err
	}
	glog.V(4).Infof("kubectl config/set-credentials: %s", output)

	// context
	b, err = exec.Command(e.GetHyperkubePath(),
		"kubectl",
		"config",
		"--kubeconfig="+kubeConfigPath,
		"set-context",
		defaultKubectlContextName,
		"--user="+defaultKubectlUserName,
		"--cluster="+defaultKubectlClusterName,
		"--namespace=default",
	).CombinedOutput()
	output = string(b)
	if err != nil {
		glog.Errorf("Cannot set kubectl set-context: %v, %s", err, output)
		return err
	}
	glog.V(4).Infof("kubectl config/set-context: %s", output)

	// use context
	b, err = exec.Command(e.GetHyperkubePath(),
		"kubectl",
		"config",
		"--kubeconfig="+kubeConfigPath,
		"use-context",
		defaultKubectlContextName,
	).CombinedOutput()
	output = string(b)
	if err != nil {
		glog.Errorf("Cannot use kubectl context: %v, %s", err, output)
		return err
	}
	glog.V(4).Infof("kubectl use-context: %s", output)

	/*
		If the kubeconfig file doesn't exists, the kubectl binary will create it
		To let the ${SUDO_USER} continues to use kubectl without privileges,
		the kubeDirPath is chown recursively
	*/
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		cmdLine := []string{"chown", "-R", fmt.Sprintf("%s:", sudoUser), kubeDirPath}
		glog.V(5).Infof("%s", strings.Join(cmdLine, " "))
		b, err := exec.Command(cmdLine[0], cmdLine[1:]...).CombinedOutput()
		output = string(b)
		if err != nil {
			glog.Warningf("Cannot execute %s: %v, %s", strings.Join(cmdLine, " "), err, output)
		}
		glog.V(4).Infof("%s output: %q", strings.Join(cmdLine, " "), output)
	}
	return e.createKubectlLink()
}

func (e *Environment) createKubectlLink() error {
	if e.kubectlLink == "" {
		return nil
	}
	glog.V(4).Infof("Creating a kubectl link %s -> %s", e.kubectlLink, e.GetHyperkubePath())
	_, err := os.Stat(e.kubectlLink)
	if err == nil {
		target, err := os.Readlink(e.kubectlLink)
		if err != nil {
			glog.Errorf("Unexpected error: %v", err)
			return err
		}
		glog.V(4).Infof("Already existent link: %s -> %s", e.kubectlLink, target)
		if target != e.GetHyperkubePath() {
			err = fmt.Errorf("link %s already created and pointing to %s", e.kubectlLink, target)
			glog.Errorf("Unexpected error: %v, please remove %s", err, e.kubectlLink)
			return err
		}
		glog.V(3).Infof("Already valid link: %s -> %s", e.kubectlLink, target)
		return nil
	}

	err = os.Symlink(e.GetHyperkubePath(), e.kubectlLink)
	if err != nil {
		glog.Errorf("Cannot create kubectl link: %v", err)
		return err
	}
	glog.V(3).Infof("Successfully created kubectl link %s", e.kubectlLink)
	return nil
}

func (e *Environment) setupKubeletClient() error {
	glog.V(4).Infof("Building kubelet client ...")
	cert, err := tls.LoadX509KeyPair(path.Join(e.secretsABSPath, "kubernetes.certificate"), path.Join(e.secretsABSPath, "kubernetes.private_key"))
	if err != nil {
		glog.Errorf("Cannot load x509 key pair: %v", err)
		return err
	}

	caCertBytes, err := ioutil.ReadFile(path.Join(e.secretsABSPath, "kubernetes.issuing_ca"))
	if err != nil {
		glog.Errorf("Cannot read CA: %v", err)
		return err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertBytes)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()

	e.kubeletClient = &http.Client{}
	e.kubeletClient.Transport = &http.Transport{TLSClientConfig: tlsConfig}

	podListURL, err := url.Parse("https://127.0.0.1:10250/pods")
	if err != nil {
		glog.Errorf("Cannot parse Kubelet URL: %s", err)
		return err
	}

	e.podListRequest = &http.Request{
		Method: http.MethodGet,
		URL:    podListURL,
	}
	return nil
}

func (e *Environment) setupAPIServerClient() error {
	var err error
	glog.V(4).Infof("Building restConfig from %s", e.GetKubeconfigInsecurePath())
	e.restConfig, err = clientcmd.BuildConfigFromFlags("", e.GetKubeconfigInsecurePath())
	if err != nil {
		glog.Errorf("Cannot build restConfig: %v", err)
		return err
	}

	e.clientSet, err = kubernetes.NewForConfig(e.restConfig)
	if err != nil {
		glog.Errorf("Cannot build clientSet: %v", err)
		return err
	}
	return nil
}

func (e *Environment) setupKubeClients() error {
	glog.V(4).Infof("Creating kubeclients configuration ...")
	err := e.setupKubectl()
	if err != nil {
		return err
	}

	err = e.setupKubeletClient()
	if err != nil {
		return err
	}

	return e.setupAPIServerClient()
}
