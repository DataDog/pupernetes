## pupernetes run

setup and run the environment

### Synopsis

setup and run the environment

```
pupernetes run [directory] [flags]
```

### Examples

```

# Setup and run the environment with the default options: 
pupernetes run state/

# Clean all the environment, setup and run the environment:
pupernetes run state/ -c all

# Clean everything but the binaries, setup and run the environment:
pupernetes run state/ -c etcd,kubectl,kubelet,manifests,network,secrets,systemd,mounts

# Setup and run the environment with a 5 minutes timeout: 
pupernetes run state/ --timeout 5m

# Setup and run the environment, then garantee a kubelet garbage collection during the drain phase: 
pupernetes run state/ --gc 1m

```

### Options

```
      --bind-address string   bind address for the API ip:port (default "127.0.0.1:8989")
  -d, --drain string          drain options after run: iptables,kubeletgc,pods,all,none (default "all")
      --gc duration           grace period for the kubelet GC trigger when draining run, no-op if not draining (default 1m0s)
  -h, --help                  help for run
      --timeout duration      timeout for run (default 6h0m0s)
```

### Options inherited from parent commands

```
  -c, --clean string                 clean options before setup: binaries,etcd,iptables,kubectl,kubelet,manifests,mounts,network,secrets,systemd,all,none (default "etcd,mounts,iptables")
      --cni-version string           container network interface (cni) version (default "0.7.0")
      --etcd-version string          etcd version (default "3.1.11")
      --hyperkube-version string     hyperkube version (default "1.10.1")
      --kubelet-cadvisor-port int    enable kubelet cAdvisor on port
      --kubelet-root-dir string      directory path for managing kubelet files (default "/var/lib/e2e-kubelet")
      --systemd-unit-prefix string   prefix for systemd unit name (default "e2e-")
      --vault-version string         vault version (default "0.9.5")
  -v, --verbose int                  verbose level (default 2)
```

### SEE ALSO

* [pupernetes](pupernetes.md)	 - Use this command to manage a Kubernetes testing environment

