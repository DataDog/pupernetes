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
pupernetes daemon clean state/

# Clean everything:
pupernetes daemon clean state/ -c all

# Clean the etcd data-dir, the network configuration and the secrets:
pupernetes daemon clean state/ -c etcd,network,secrets

```

### Options

```
  -h, --help   help for clean
```

### Options inherited from parent commands

```
  -c, --clean string                 clean options before setup: binaries,etcd,iptables,kubectl,kubelet,manifests,mounts,network,secrets,systemd,all,none (default "etcd,kubelet,mounts,iptables")
      --cni-version string           container network interface (cni) version (default "0.7.0")
      --etcd-version string          etcd version (default "3.1.11")
      --hyperkube-version string     hyperkube version (default "1.10.1")
      --kubectl-link string          Path to create a kubectl link
      --kubelet-cadvisor-port int    enable kubelet cAdvisor on port
      --kubelet-root-dir string      directory path for managing kubelet files (default "/var/lib/p8s-kubelet")
      --systemd-unit-prefix string   prefix for systemd unit name (default "p8s-")
      --vault-version string         vault version (default "0.9.5")
  -v, --verbose int                  verbose level (default 2)
```

### SEE ALSO

* [pupernetes daemon](pupernetes_daemon.md)	 - Use this command to clean setup and run a Kubernetes local environment

