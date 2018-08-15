## pupernetes daemon setup

Setup the environment

### Synopsis

Setup the environment

```
pupernetes daemon setup [directory] [flags]
```

### Examples

```
pupernetes daemon setup state/
```

### Options

```
  -h, --help   help for setup
```

### Options inherited from parent commands

```
  -c, --clean string                         clean options before setup: binaries,etcd,iptables,kubectl,kubelet,logs,manifests,mounts,network,secrets,systemd,all,none (default "etcd,kubelet,logs,mounts,iptables")
      --cni-version string                   container network interface (cni) version (default "0.7.0")
      --container-runtime string             container runtime interface to use (experimental: "containerd") (default "docker")
      --download-timeout string              timeout for each downloaded archive (default "30m0s")
      --etcd-version string                  etcd version (default "3.1.11")
      --hyperkube-version string             hyperkube version (default "1.10.3")
  -k, --keep string                          clean everything but the given options before setup: binaries,etcd,iptables,kubectl,kubelet,logs,manifests,mounts,network,secrets,systemd,all,none
      --kubeconfig-path string               path to the kubeconfig file
      --kubectl-link string                  path to create a kubectl link
      --kubelet-root-dir string              directory path for managing kubelet files (default "/var/lib/p8s-kubelet")
      --kubernetes-cluster-ip-range string   kubernetes cluster CIDR (default "192.168.254.0/24")
      --pod-ip-range string                  pod common network interface CIDR (default "192.168.253.0/24")
      --systemd-unit-prefix string           prefix for systemd unit name (default "p8s-")
      --vault-version string                 vault version (default "0.9.5")
  -v, --verbose int                          verbose level (default 2)
```

### SEE ALSO

* [pupernetes daemon](pupernetes_daemon.md)	 - Use this command to clean setup and run a Kubernetes local environment

