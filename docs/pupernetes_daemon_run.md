## pupernetes daemon run

setup and run the environment

### Synopsis

setup and run the environment

```
pupernetes daemon run [directory] [flags]
```

### Examples

```

# Setup and run the environment with the default options: 
pupernetes daemon run state/

# Clean all the environment, setup and run the environment:
pupernetes daemon run state/ -c all

# Clean everything but the binaries, setup and run the environment:
pupernetes daemon run state/ -c etcd,kubectl,kubelet,manifests,network,secrets,systemd,mounts

# Setup and run the environment with a 5 minutes timeout: 
pupernetes daemon run state/ --timeout 5m

# Setup and run the environment, then garantee a kubelet garbage collection during the drain phase: 
pupernetes daemon run state/ --gc 1m

# Setup and run the environment as a systemd service:
# Get logs with "journalctl -o cat -efu pupernetes" 
# Get status with "systemctl status pupernetes --no-pager" 
pupernetes daemon run state/ --job-type systemd

```

### Options

```
      --bind-address string       Bind address for pupernetes API ip:port (default "127.0.0.1:8989")
  -d, --drain string              drain options after run: iptables,kubeletgc,pods,all,none (default "all")
      --gc duration               grace period for the kubelet GC trigger when draining run, no-op if not draining (default 1m0s)
  -h, --help                      help for run
      --job-type string           type of job: fg or systemd (default "fg")
      --systemd-job-name string   unit name used when running as systemd service (default "pupernetes")
      --timeout duration          timeout for run (default 6h0m0s)
```

### Options inherited from parent commands

```
  -c, --clean string                 clean options before setup: binaries,etcd,iptables,kubectl,kubelet,manifests,mounts,network,secrets,systemd,all,none (default "etcd,kubelet,mounts,iptables")
      --cni-version string           container network interface (cni) version (default "0.7.0")
      --etcd-version string          etcd version (default "3.1.11")
      --hyperkube-version string     hyperkube version (default "1.10.3")
      --kubectl-link string          Path to create a kubectl link
      --kubelet-cadvisor-port int    enable kubelet cAdvisor on port
      --kubelet-root-dir string      directory path for managing kubelet files (default "/var/lib/p8s-kubelet")
      --systemd-unit-prefix string   prefix for systemd unit name (default "p8s-")
      --vault-version string         vault version (default "0.9.5")
  -v, --verbose int                  verbose level (default 2)
```

### SEE ALSO

* [pupernetes daemon](pupernetes_daemon.md)	 - Use this command to clean setup and run a Kubernetes local environment

