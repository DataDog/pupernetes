## pupernetes daemon clean

Clean the environment created by setup and altered by a run

### Synopsis

Clean the environment created by setup and altered by a run

```
pupernetes daemon clean [directory] [flags]
```

### Examples

```

# Clean the environment default:
pupernetes daemon clean /opt/state/

# Clean everything:
pupernetes daemon clean /opt/state/ -c all

# Clean the etcd data-dir, the network configuration and the secrets:
pupernetes daemon clean /opt/state/ -c etcd,network,secrets

```

### Options

```
  -h, --help   help for clean
```

### Options inherited from parent commands

```
  -c, --clean string                         clean options before setup: binaries,etcd,iptables,kubectl,kubelet,logs,manifests,mounts,network,secrets,systemd,all,none (default "etcd,kubelet,logs,mounts,iptables")
      --cni-version string                   container network interface (cni) version (default "0.7.0")
      --container-runtime string             container runtime interface to use (experimental: "containerd") (default "docker")
      --download-timeout string              timeout for each downloaded archive (default "30m0s")
      --etcd-version string                  etcd version (default "3.1.19")
      --hyperkube-version string             hyperkube version (default "1.10.7")
  -k, --keep string                          clean everything but the given options before setup: binaries,etcd,iptables,kubectl,kubelet,logs,manifests,mounts,network,secrets,systemd,all,none, this flag overrides any clean options
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

