// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	"github.com/golang/glog"
	vault "github.com/hashicorp/vault/api"
)

const (
	rootCertificateAuthorityName = "pupernetes"
)

var pemParts = []string{"certificate", "issuing_ca", "private_key"}

func tearDownCommand(cmd *exec.Cmd, originalErr error) error {
	glog.V(4).Infof("Stopping vault")
	cmd.Process.Signal(syscall.SIGTERM)

	err := cmd.Wait()
	if err == nil {
		return originalErr
	}
	glog.Errorf("Unexpected error during vault shutdown: %v", err)
	cmd.Process.Signal(syscall.SIGKILL)
	err = cmd.Wait()
	if err == nil {
		return fmt.Errorf("error during the vault shutdown: SIGKILL")
	}
	glog.Errorf("Unexpected error during vault timeout shutdown: %v", err)
	return err
}

func waitForUnseal(vRaw *vault.Client) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.NewTimer(time.Second * 5)
	defer timeout.Stop()

	for {
		select {
		case <-ticker.C:
			_, err := vRaw.Sys().SealStatus()
			if err == nil {
				return nil
			}
			glog.Warningf("Vault still not ready: %v", err)
		case <-timeout.C:
			err := fmt.Errorf("timeout awaiting vault")
			glog.Errorf("%v", err)
			return err
		}
	}
}

func (e *Environment) generateVaultPKI() error {
	vaultCFG := vault.DefaultConfig()
	vaultCFG.Address = "http://127.0.0.1:8200"
	vRaw, err := vault.NewClient(vaultCFG)
	if err != nil {
		glog.Errorf("Cannot use vault client: %v", err)
		return err
	}
	vRaw.SetToken(e.vaultRootToken)
	vClient := vRaw.Logical()
	err = waitForUnseal(vRaw)
	if err != nil {
		glog.Errorf("Unexpected state of vault: %v", err)
		return err
	}

	/*
		The ROOT CA is pupernetes.
		The certificates for etcd and kubernetes are issued against the pupernetes ROOT CA.
	*/

	// ROOT CA - pupernetes
	glog.V(3).Infof("Mounting pupernetes root CA")
	err = vRaw.Sys().Mount(rootCertificateAuthorityName, &vault.MountInput{
		Config:     vault.MountConfigInput{MaxLeaseTTL: "87600h"},
		Type:       "pki",
		PluginName: "pki",
	})
	if err != nil {
		glog.Errorf("Cannot mount pki: %v", err)
		return err
	}
	glog.V(3).Infof("Creating %s root CA", rootCertificateAuthorityName)
	rootCAConf := make(map[string]interface{})
	rootCAConf["common_name"] = "p8s"
	rootCAConf["ttl"] = "87600h"
	sec, err := vClient.Write(rootCertificateAuthorityName+"/root/generate/exported", rootCAConf)
	if err != nil {
		glog.Errorf("Cannot write: %v", err)
		return err
	}

	privateKey := []byte(sec.Data["private_key"].(string))
	certificate := []byte(sec.Data["certificate"].(string))
	issuingCA := []byte(sec.Data["issuing_ca"].(string))

	err = ioutil.WriteFile(path.Join(e.secretsABSPath, rootCertificateAuthorityName+".private_key"), privateKey, 0444)
	if err != nil {
		glog.Errorf("Cannot write secret file: %v", err)
		return err
	}
	err = ioutil.WriteFile(path.Join(e.secretsABSPath, rootCertificateAuthorityName+".certificate"), certificate, 0444)
	if err != nil {
		glog.Errorf("Cannot write secret file: %v", err)
		return err
	}
	err = ioutil.WriteFile(path.Join(e.secretsABSPath, rootCertificateAuthorityName+".issuing_ca"), issuingCA, 0444)
	if err != nil {
		glog.Errorf("Cannot write secret file: %v", err)
		return err
	}

	// Prepare the role / issue configuration
	roleConf := make(map[string]interface{})
	roleConf["allow_any_name"] = "true"
	roleConf["max_ttl"] = "43800h"
	issueConf := make(map[string]interface{})
	issueConf["common_name"] = "p8s"
	issueConf["alt_names"] = fmt.Sprintf("localhost,%s", e.hostname)
	issueConf["ip_sans"] = fmt.Sprintf("127.0.0.1,%s,%s", e.outboundIP.String(), e.kubernetesClusterIP.String())

	// Generate secrets - certificates for each component:
	for _, component := range []string{"kubernetes", "etcd"} {
		err = e.generateSecretFor(vRaw, vClient, roleConf, issueConf, component)
		if err != nil {
			glog.Errorf("Unexpected error during the secret generation of %s: %v", component, err)
			return err
		}
	}
	return nil
}

func (e *Environment) generateSecretFor(vRaw *vault.Client, vClient *vault.Logical, roleConf, issueConf map[string]interface{}, component string) error {
	_, err := vClient.Write(fmt.Sprintf("%s/roles/%s", rootCertificateAuthorityName, component), roleConf)
	if err != nil {
		glog.Errorf("Cannot write role: %v", err)
		return err
	}
	err = vRaw.Sys().PutPolicy(fmt.Sprintf("%s/%s", rootCertificateAuthorityName, component), fmt.Sprintf(`path "%s/issue/%s" { policy = "write" }`, rootCertificateAuthorityName, component))
	if err != nil {
		glog.Errorf("Cannot write policy: %v", err)
		return err
	}
	sec, err := vClient.Write(fmt.Sprintf("%s/issue/%s", rootCertificateAuthorityName, component), issueConf)
	if err != nil {
		glog.Errorf("Cannot generateSecretFor %s: %v", component, err)
		return err
	}
	for _, part := range pemParts {
		content := []byte(sec.Data[part].(string))
		certABSPath := path.Join(e.secretsABSPath, fmt.Sprintf("%s.%s", component, part))
		err = ioutil.WriteFile(certABSPath, content, 0444)
		if err != nil {
			return err
		}
		glog.V(4).Infof("Successfully created %s", certABSPath)
	}

	// Creating the pem_bundle
	certificate := []byte(sec.Data["certificate"].(string))
	issuingCA := []byte(sec.Data["issuing_ca"].(string))
	err = ioutil.WriteFile(path.Join(e.secretsABSPath, component+".bundle"), append(certificate, issuingCA...), 0444)
	if err != nil {
		glog.Errorf("Cannot write secret file: %v", err)
		return err
	}
	return nil
}

func (e *Environment) generateServiceAccountRSA() error {
	rsaPath := path.Join(e.secretsABSPath, "service-accounts.rsa")
	_, err := os.Stat(rsaPath)
	if err == nil {
		glog.V(4).Infof("Service Account RSA key already here: %s", rsaPath)
		return nil
	}

	rsaFile, err := os.OpenFile(rsaPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0444)
	if err != nil {
		glog.Errorf("Cannot open file for rsa: %v", err)
		return err
	}
	defer rsaFile.Close()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		glog.Errorf("Unexpected error during the RSA Key generation: %v", err)
		return err
	}
	b := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	_, err = rsaFile.Write(b)
	if err != nil {
		glog.Errorf("Cannot write rsa key: %v", err)
		return err
	}
	glog.V(4).Infof("Successfully generated Service Account RSA: %s", rsaPath)
	return nil
}

func (e *Environment) isVaultSecrets() bool {
	for _, part := range pemParts {
		certFile := path.Join(e.secretsABSPath, "apiserver."+part)
		_, err := os.Stat(certFile)
		if err != nil {
			return false
		}
	}
	glog.V(4).Infof("Already created vault secrets in %s", e.secretsABSPath)
	return true
}

func (e *Environment) generateVaultSecrets() error {
	if e.isVaultSecrets() {
		return nil
	}
	glog.V(4).Infof("Starting vault %s", e.binaryVault.binaryABSPath)
	cmd := exec.Command(e.binaryVault.binaryABSPath, "server", "-dev", "-dev-root-token-id="+e.vaultRootToken)
	var std bytes.Buffer
	cmd.Stderr = &std
	cmd.Stdout = &std
	err := cmd.Start()
	if err != nil {
		glog.Errorf("Cannot start vault: %v", err)
		return err
	}
	glog.V(4).Infof("Vault successfully started")

	err = e.generateVaultPKI()
	level := glog.V(5)
	if err != nil {
		level = glog.V(0)
		glog.Errorf("Unexpected error during vault commands: %v", err)
	}
	err = tearDownCommand(cmd, err)
	if err != nil {
		level = glog.V(0)
	}
	level.Infof("Vault logs:\n%s", std.String())
	return err
}

func (e *Environment) setupSecrets() error {
	err := e.generateServiceAccountRSA()
	if err != nil {
		return err
	}
	err = e.generateVaultSecrets()
	if err != nil {
		return err
	}
	return nil
}
