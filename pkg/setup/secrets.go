// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	"github.com/golang/glog"
	vault "github.com/hashicorp/vault/api"
)

const vaultRootToken = "1234567890"

var pemParts = []string{"certificate", "issuing_ca", "private_key"}

func tearDownCommand(cmd *exec.Cmd, originalErr error) error {
	glog.V(4).Infof("Stopping vault")
	cmd.Process.Signal(syscall.SIGTERM)
	select {
	case <-time.After(0):
		err := cmd.Wait()
		if err == nil {
			return originalErr
		}
		glog.Errorf("Unexpected error during vault shutdown: %v", err)
		return err
	case <-time.After(time.Second * 10):
		glog.Warningf("Timeout, SIGKILL vault")
		cmd.Process.Signal(syscall.SIGKILL)
		err := cmd.Wait()
		if err == nil {
			return fmt.Errorf("timeout during the vault shutdown")
		}
		glog.Errorf("Unexpected error during vault timeout shutdown: %v", err)
		return err
	}
}

func waitForUnseal(vRaw *vault.Client) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_, err := vRaw.Sys().SealStatus()
			if err == nil {
				return nil
			}
			glog.Warningf("Vault still not ready: %v", err)
		case <-time.After(time.Second * 5):
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
	vRaw.SetToken(vaultRootToken)
	err = waitForUnseal(vRaw)
	if err != nil {
		glog.Errorf("Unexpected state of vault: %v", err)
		return err
	}
	err = vRaw.Sys().Mount("pki/pupernetes", &vault.MountInput{
		Config:     vault.MountConfigInput{MaxLeaseTTL: "87600h"},
		Type:       "pki",
		PluginName: "pki",
	})
	if err != nil {
		glog.Errorf("Cannot mount pki: %v", err)
		return err
	}

	vClient := vRaw.Logical()
	caConf := make(map[string]interface{})
	caConf["common_name"] = "p8s"
	caConf["ttl"] = "87600h"
	_, err = vClient.Write("pki/pupernetes/root/generate/internal", caConf)
	if err != nil {
		glog.Errorf("Cannot write: %v", err)
		return err
	}

	roleConf := make(map[string]interface{})
	roleConf["allow_any_name"] = "true"
	roleConf["max_ttl"] = "43800h"
	issueConf := make(map[string]interface{})
	issueConf["common_name"] = "p8s"
	issueConf["ip_sans"] = fmt.Sprintf("127.0.0.1,%s,%s", e.outboundIP.String(), e.kubernetesClusterIP.String())
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
	_, err := vClient.Write(fmt.Sprintf("pki/pupernetes/roles/%s", component), roleConf)
	if err != nil {
		glog.Errorf("Cannot write role: %v", err)
		return err
	}
	err = vRaw.Sys().PutPolicy(fmt.Sprintf("pupernetes/%s", component),
		fmt.Sprintf(`path "pki/pupernetes/issue/%s" { policy = "write" }`, component))
	if err != nil {
		glog.Errorf("Cannot write policy: %v", err)
		return err
	}
	sec, err := vClient.Write(fmt.Sprintf("pki/pupernetes/issue/%s", component), issueConf)
	if err != nil {
		glog.Errorf("Cannot generateSecretFor %s: %v", component, err)
		return err
	}
	for _, part := range pemParts {
		certFile, err := os.OpenFile(path.Join(e.secretsABSPath, fmt.Sprintf("%s.%s", component, part)), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0444)
		if err != nil {
			return err
		}
		_, err = certFile.WriteString(sec.Data[part].(string))
		if err != nil {
			return err
		}
		glog.V(4).Infof("Successfully created %s", certFile.Name())
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

	b, err := exec.Command("openssl", "genrsa", "2048").CombinedOutput()
	if err != nil {
		glog.Errorf("Cannot generate rsa: %s %v", string(b), err)
		return err
	}
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
	cmd := exec.Command(e.binaryVault.binaryABSPath, "server", "-dev", "-dev-root-token-id="+vaultRootToken)
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
