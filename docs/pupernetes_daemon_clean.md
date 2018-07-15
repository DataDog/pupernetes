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
  -c, --clean string                 clean options before setup: binaries,etcd,iptables,kubectl,kubelet,logs,manifests,mounts,network,secrets,systemd,all,none (default "etcd,kubelet,logs,mounts,iptables")
      --cni-version string           container network interface (cni) version (default "0.7.0")
      --download-timeout string      timeout for each downloaded archive (default "30m0s")
      --etcd-version string          etcd version (default "3.1.11")
      --hyperkube-version string     hyperkube version (default "1.10.3")
      --kubeconfig-path string       path to the kubeconfig file
      --kubectl-link string          path to create a kubectl link
      --kubelet-root-dir string      directory path for managing kubelet files (default "/var/lib/p8s-kubelet")
      --systemd-unit-prefix string   prefix for systemd unit name (default "p8s-")
      --vault-version string         vault version (default "0.9.5")
  -v, --verbose int                  verbose level (default 2)
```

### SEE ALSO

* [pupernetes daemon](pupernetes_daemon.md)	 - Use this command to clean setup and run a Kubernetes local environment

