## pupernetes daemon

Use this command to clean setup and run a Kubernetes local environment

### Synopsis

Use this command to clean setup and run a Kubernetes local environment

### Options

```
  -c, --clean string                 clean options before setup: binaries,etcd,iptables,kubectl,kubelet,manifests,mounts,network,secrets,systemd,all,none (default "etcd,kubelet,mounts,iptables")
      --cni-version string           container network interface (cni) version (default "0.7.0")
      --etcd-version string          etcd version (default "3.1.11")
  -h, --help                         help for daemon
      --hyperkube-version string     hyperkube version (default "1.10.3")
      --kubeconfig-path string       path to the kubeconfig file
      --kubectl-link string          path to create a kubectl link
      --kubelet-root-dir string      directory path for managing kubelet files (default "/var/lib/p8s-kubelet")
      --systemd-unit-prefix string   prefix for systemd unit name (default "p8s-")
      --vault-version string         vault version (default "0.9.5")
```

### Options inherited from parent commands

```
  -v, --verbose int   verbose level (default 2)
```

### SEE ALSO

* [pupernetes](pupernetes.md)	 - Use this command to manage a Kubernetes local environment
* [pupernetes daemon clean](pupernetes_daemon_clean.md)	 - Clean the environment created by setup and altered by a run
* [pupernetes daemon run](pupernetes_daemon_run.md)	 - setup and run the environment
* [pupernetes daemon setup](pupernetes_daemon_setup.md)	 - Setup the environment

