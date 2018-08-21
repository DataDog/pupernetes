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
pupernetes daemon run /opt/state/

# Clean all the environment, setup and run the environment:
pupernetes daemon run /opt/state/ -c all

# Clean everything but the binaries, setup and run the environment:
pupernetes daemon run /opt/state/ -c etcd,kubectl,kubelet,manifests,network,secrets,systemd,mounts

# Setup and run the environment with a 5 minutes timeout: 
pupernetes daemon run /opt/state/ --run-timeout 5m

# Setup and run the environment, then guarantee a kubelet garbage collection during the drain phase: 
pupernetes daemon run /opt/state/ --gc 1m

# Setup and run the environment as a systemd service:
# Get logs with "journalctl -o cat -efu pupernetes" 
# Get status with "systemctl status pupernetes --no-pager" 
pupernetes daemon run /opt/state/ --job-type systemd

# Setup and run the environment with a readiness on dns:
pupernetes daemon run /opt/state/ --dns-check --dns-queries quay.io.,coredns.kube-system.svc.cluster.local.

```

### Options

```
      --bind-address string       bind address for pupernetes API ip:port (default "127.0.0.1:8989")
      --dns-check                 needed dns queries to notify readiness
      --dns-queries stringSlice   dns queries for readiness, coma-separated values (default [coredns.kube-system.svc.cluster.local.])
  -d, --drain string              drain options after run: iptables,kubeletgc,pods,all,none (default "all")
      --gc duration               grace period for the kubelet GC trigger when draining run, no-op if not draining (default 1m0s)
  -h, --help                      help for run
      --job-type string           type of job: fg or systemd (default "fg")
      --run-timeout duration      maximum time to run pupernetes for until self shutdown
      --skip-probes               skip probing systemd units and kubelet healthz
      --systemd-job-name string   unit name used when running as systemd service (default "pupernetes")
```

### Options inherited from parent commands

```
  -c, --clean string                         clean options before setup: binaries,etcd,iptables,kubectl,kubelet,logs,manifests,mounts,network,secrets,systemd,all,none (default "etcd,kubelet,logs,mounts,iptables")
      --cni-version string                   container network interface (cni) version (default "0.7.0")
      --container-runtime string             container runtime interface to use (experimental: "containerd") (default "docker")
      --download-timeout string              timeout for each downloaded archive (default "30m0s")
      --etcd-version string                  etcd version (default "3.1.11")
      --hyperkube-version string             hyperkube version (default "1.10.3")
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

