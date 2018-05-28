// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package setup

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/unit"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/DataDog/pupernetes/pkg/config"
	"github.com/DataDog/pupernetes/pkg/options"
	defaultTemplates "github.com/DataDog/pupernetes/pkg/setup/templates"
	"github.com/DataDog/pupernetes/pkg/util"
)

const (
	KubeletCRILogPath = "/var/log/pods/"

	defaultBinaryDirName          = "bin"
	defaultSourceTemplatesDirName = "source-templates"
	defaultEtcdDataDirName        = "etcd-data"
	defaultSecretDirName          = "secrets"
	defaultNetworkDirName         = "net.d"

	defaultKubectlClusterName = "p8s"
	defaultKubectlUserName    = "p8s"
	defaultKubectlContextName = "p8s"
)

type Environment struct {
	// host
	rootABSPath string

	binABSPath string

	manifestTemplatesABSPath string
	manifestAPIABSPath       string
	manifestSystemdUnit      string
	manifestStaticPodABSPath string
	manifestConfigABSPath    string
	secretsABSPath           string
	networkABSPath           string

	kubeletRootDir string

	kubeConfigAuthPath     string
	kubeConfigInsecurePath string
	etcdDataABSPath        string

	cleanOptions *options.Clean
	drainOptions *options.Drain

	hostname   string
	dbusClient *dbus.Conn

	// executable dependencies
	binaryHyperkube *exeBinary
	binaryVault     *exeBinary
	binaryEtcd      *exeBinary

	// dependencies
	binaryCNI *depBinary

	templateMetadata *templateMetadata

	// Kubernetes Major.Minor
	templateVersion string

	systemdEnd2EndSection []*unit.UnitOption

	// Kubernetes apiserver
	restConfig *rest.Config
	clientSet  *kubernetes.Clientset

	// Kubernetes kubelet
	kubeletClient  *http.Client
	podListRequest *http.Request

	// Network
	outboundIP          *net.IP
	kubernetesClusterIP *net.IP
	dnsClusterIP        *net.IP
	isDockerBridge      bool

	// Vault token
	vaultRootToken string

	kubectlLink string
}

type templateMetadata struct {
	HyperkubeImageURL     string  `json:"hyperkube-image-url"`
	Hostname              *string `json:"hostname"`
	RootABSPath           *string `json:"root"`
	ServiceClusterIPRange string  `json:"service-cluster-ip-range"`
	KubernetesClusterIP   string  `json:"kubernetes-cluster-ip"`
	DNSClusterIP          string  `json:"dns-cluster-ip"`
	NodeIP                string  `json:"node-ip"`
	KubeletRootDirABSPath string  `json:"kubelet-root-dir"`
}

func NewConfigSetup(givenRootPath string) (*Environment, error) {
	if givenRootPath == "" {
		err := fmt.Errorf("must provide a path")
		glog.Errorf("%v", err)
		return nil, err
	}
	rootABSPath, err := filepath.Abs(givenRootPath)
	if err != nil {
		glog.Errorf("Unexpected error during abspath: %v", err)
		return nil, err
	}

	e := &Environment{
		rootABSPath: rootABSPath,
		binABSPath:  path.Join(rootABSPath, defaultBinaryDirName),

		manifestTemplatesABSPath: path.Join(rootABSPath, defaultSourceTemplatesDirName),
		manifestStaticPodABSPath: path.Join(rootABSPath, defaultTemplates.ManifestStaticPod),
		manifestAPIABSPath:       path.Join(rootABSPath, defaultTemplates.ManifestAPI),
		manifestConfigABSPath:    path.Join(rootABSPath, defaultTemplates.ManifestConfig),
		manifestSystemdUnit:      path.Join(rootABSPath, defaultTemplates.ManifestSystemdUnit),
		kubeletRootDir:           config.ViperConfig.GetString("kubelet-root-dir"),
		secretsABSPath:           path.Join(rootABSPath, defaultSecretDirName),
		networkABSPath:           path.Join(rootABSPath, defaultNetworkDirName),
		templateVersion:          getMajorMinorVersion(config.ViperConfig.GetString("hyperkube-version")),

		kubeConfigAuthPath:     path.Join(rootABSPath, defaultTemplates.ManifestConfig, "kubeconfig-auth.yaml"),
		kubeConfigInsecurePath: path.Join(rootABSPath, defaultTemplates.ManifestConfig, "kubeconfig-insecure.yaml"),
		etcdDataABSPath:        path.Join(rootABSPath, defaultEtcdDataDirName),
		cleanOptions:           options.NewCleanOptions(config.ViperConfig.GetString("clean")),
		drainOptions:           options.NewDrainOptions(config.ViperConfig.GetString("drain")),
		kubectlLink:            config.ViperConfig.GetString("kubectl-link"),
	}

	// Kubernetes
	e.binaryHyperkube = &exeBinary{
		depBinary: depBinary{
			archivePath:   path.Join(e.binABSPath, fmt.Sprintf("hyperkube-v%s.tar.gz", config.ViperConfig.GetString("hyperkube-version"))),
			binaryABSPath: path.Join(e.binABSPath, "hyperkube"),
			archiveURL:    fmt.Sprintf("https://dl.k8s.io/v%s/kubernetes-server-linux-amd64.tar.gz", config.ViperConfig.GetString("hyperkube-version")),
			version:       config.ViperConfig.GetString("hyperkube-version"),
		},
		commandVersion: []string{"kubelet", "--version"},
	}

	// Vault
	e.binaryVault = &exeBinary{
		depBinary: depBinary{
			archivePath:   path.Join(e.binABSPath, fmt.Sprintf("vault-v%s.zip", config.ViperConfig.GetString("vault-version"))),
			binaryABSPath: path.Join(e.binABSPath, "vault"),
			archiveURL:    fmt.Sprintf("https://releases.hashicorp.com/vault/%s/vault_%s_linux_amd64.zip", config.ViperConfig.GetString("vault-version"), config.ViperConfig.GetString("vault-version")),
			version:       config.ViperConfig.GetString("vault-version"),
		},
		commandVersion: []string{"--version"},
	}

	// Etcd
	e.binaryEtcd = &exeBinary{
		depBinary: depBinary{
			archivePath:   path.Join(e.binABSPath, fmt.Sprintf("etcd-v%s.tar.gz", config.ViperConfig.GetString("etcd-version"))),
			binaryABSPath: path.Join(e.binABSPath, "etcd"),
			archiveURL:    fmt.Sprintf("https://github.com/coreos/etcd/releases/download/v%s/etcd-v%s-linux-amd64.tar.gz", config.ViperConfig.GetString("etcd-version"), config.ViperConfig.GetString("etcd-version")),
			version:       config.ViperConfig.GetString("etcd-version"),
		},
		commandVersion: []string{"--version"},
	}

	// CNI
	e.binaryCNI = &depBinary{
		archivePath:   path.Join(e.binABSPath, fmt.Sprintf("cni-v%s.tar.gz", config.ViperConfig.GetString("cni-version"))),
		binaryABSPath: path.Join(e.binABSPath, "bridge"),
		archiveURL:    fmt.Sprintf("https://github.com/containernetworking/plugins/releases/download/v%s/cni-plugins-amd64-v%s.tgz", config.ViperConfig.GetString("cni-version"), config.ViperConfig.GetString("cni-version")),
		version:       config.ViperConfig.GetString("cni-version"),
	}

	// SystemdUnits X-Section
	e.systemdEnd2EndSection = e.createEnd2EndSection()

	e.kubernetesClusterIP, err = getKubernetesClusterIP()
	if err != nil {
		glog.Errorf("Unexpected error: %v", err)
		return nil, err
	}

	e.dnsClusterIP, err = getDNSClusterIP()
	if err != nil {
		glog.Errorf("Unexpected error: %v", err)
		return nil, err
	}

	// Template for manifests
	e.templateMetadata = &templateMetadata{
		// TODO conf this
		HyperkubeImageURL:     fmt.Sprintf("gcr.io/google_containers/hyperkube:v%s", e.binaryHyperkube.version),
		Hostname:              &e.hostname,
		RootABSPath:           &e.rootABSPath,
		ServiceClusterIPRange: config.ViperConfig.GetString("kubernetes-cluster-ip-range"),
		KubernetesClusterIP:   e.kubernetesClusterIP.String(),
		DNSClusterIP:          e.dnsClusterIP.String(),
		KubeletRootDirABSPath: e.kubeletRootDir,
		// NodeIP: during network phase
	}

	// Vault root token
	e.vaultRootToken = config.ViperConfig.GetString("vault-root-token")
	if e.vaultRootToken == "" {
		e.vaultRootToken = util.RandStringBytesMaskImprSrc(20)
		glog.V(4).Infof("Generated the vault root-token of length: %d", len(e.vaultRootToken))
	}
	return e, nil
}

func (e *Environment) setupDirectories() error {
	for _, dir := range []string{
		e.binABSPath,
		e.manifestTemplatesABSPath,
		e.manifestStaticPodABSPath,
		e.manifestConfigABSPath,
		e.manifestSystemdUnit,
		path.Join(e.manifestTemplatesABSPath, defaultTemplates.ManifestSystemdUnit),
		path.Join(e.manifestTemplatesABSPath, defaultTemplates.ManifestStaticPod),
		e.manifestAPIABSPath,
		path.Join(e.manifestTemplatesABSPath, defaultTemplates.ManifestAPI),
		path.Join(e.manifestTemplatesABSPath, defaultTemplates.ManifestConfig),
		e.etcdDataABSPath,
		e.secretsABSPath,
		e.networkABSPath,
		e.kubeletRootDir,
	} {
		glog.V(4).Infof("Creating directory: %s", dir)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			glog.Errorf("Cannot create %s: %v", dir, err)
			return err
		}
		glog.V(4).Infof("Directory exists: %s", dir)
	}
	return nil
}

func (e *Environment) Setup() error {
	var err error
	glog.V(3).Infof("Setup starting %s", e.rootABSPath)
	for _, f := range []func() error{
		checkRequirements,
		e.setupHostname,
		e.setupDirectories,
		e.setupBinaryCNI,
		e.setupBinaryEtcd,
		e.setupBinaryVault,
		e.setupBinaryHyperkube,
		e.setupNetwork,
		e.setupManifests,
		e.setupSystemd,
		e.setupSecrets,
		e.setupKubeClients,
	} {
		err = f()
		if err != nil {
			return err
		}
	}
	glog.V(2).Infof("Setup ready %s", e.rootABSPath)
	return nil
}
