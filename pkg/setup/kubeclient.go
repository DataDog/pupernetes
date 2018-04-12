// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func (e *Environment) setupKubectl() error {
	// cluster
	b, err := exec.Command(e.GetHyperkubePath(),
		"kubectl",
		"config",
		"set-cluster",
		defaultKubectlClusterName,
		"--server=https://127.0.0.1:6443",
		"--certificate-authority="+path.Join(e.secretsABSPath, "apiserver.issuing_ca"),
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
		"set-credentials",
		defaultKubectlUserName,
		"--username="+defaultKubectlUserName,
		"--client-certificate="+path.Join(e.secretsABSPath, "apiserver.certificate"),
		"--client-key="+path.Join(e.secretsABSPath, "apiserver.private_key"),
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
	b, err = exec.Command(e.GetHyperkubePath(), "kubectl", "config", "use-context", "e2e").CombinedOutput()
	output = string(b)
	if err != nil {
		glog.Errorf("Cannot use kubectl context: %v, %s", err, output)
		return err
	}
	glog.V(4).Infof("kubectl use-context: %s", output)

	sudoUser := os.Getenv("SUDO_USER")
	// Atoi returns 0 on error so we can deal with it
	sudoUID, _ := strconv.Atoi(os.Getenv("SUDO_UID"))
	sudoGID, _ := strconv.Atoi(os.Getenv("SUDO_GID"))
	if sudoUser != "" && sudoUID != 0 && sudoGID != 0 {
		kubeConfig := fmt.Sprintf("/home/%s/.kube/config", sudoUser)
		glog.V(5).Infof("chown %s: %s", sudoUser, kubeConfig)
		os.Chown(kubeConfig, sudoUID, sudoGID)
	}
	return nil
}

func (e *Environment) setupKubeletClient() error {
	cert, err := tls.LoadX509KeyPair(path.Join(e.secretsABSPath, "apiserver.certificate"), path.Join(e.secretsABSPath, "apiserver.private_key"))
	if err != nil {
		glog.Errorf("Cannot load x509 key pair: %v", err)
		return err
	}

	caCertBytes, err := ioutil.ReadFile(path.Join(e.secretsABSPath, "apiserver.issuing_ca"))
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
	glog.V(4).Infof("Building restConfig from %s", e.GetKubeconfigAuthPath())
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
