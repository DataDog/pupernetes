# pupernetes - p8s

Run a Kubernetes setup in 45 seconds.

![img](docs/pupernetes.jpg)

**Provides**:

* etcd v3
* kubectl
* kubelet
* kube-apiserver
* kube-scheduler
* kube-controller-manager
* kube-proxy
* coredns

The default setup is secured with:

* x509 certificates provided by an embedded vault PKI
* HTTPS webhook to provide token lookups for the kubelet API
* RBAC

## Table of Contents
- [Requirements](#requirements)
  * [Runtime](#runtime)
  * [Development](#development)
    + [Install VMWare Fusion](#install-vmware-fusion)
    + [Create Ubuntu VM](#create-ubuntu-vm)
    + [Install Docker](#install-docker)
- [Build it](#build-it)
- [Run it](#run-it)
- [Use it](#use-it)
  * [Command line](#command-line)
  * [Quick run](#quick-run)
  * [Quick systemd-run](#quick-systemd-run)
  * [Systemd as job type](#systemd-as-job-type)
- [Current limitations](#current-limitations)

## Requirements

### Runtime

Executables in PATH:

* tar
* unzip
* systemctl
* systemd-resolve (or a non-systemd managed `/etc/resolv.conf`)
* openssl
* mount

Any implicit requirements for the **kubelet** like the container runtime and [more](https://github.com/kubernetes/kubernetes/issues/26093)

A systemd environment.

### Development

Setup a linux environment for running `pupernetes`. **This is only a suggested environment for running pupernetes. You could also create a VM using Vagrant (not yet documented here).**

#### Install VMWare Fusion

`pupernetes` must be run on linux. To run a linux VM install [VMWare Fusion](https://www.vmware.com/products/fusion/fusion-evaluation.html) or your preferred virtualization software. You can use the VMWare Fusion Pro 30-day trial.

#### Create Ubuntu VM

Download the latest version of [Ubuntu Desktop](https://www.ubuntu.com/download/desktop) and create the Ubuntu VM with VMWare Fusion or whichever virtualization software you prefer.

#### Install Docker

Follow the instructions [here](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-docker-ce) to install docker.

>Note:
>
>If you are seeing the following error after running `sudo apt-get install docker-ce` to install `docker-ce`.
>
>```
>E: Invalid operation docker-ce
>```
>
>Try running the following command to setup the **stable** repository that instead specifies an older Ubuntu distribution like `xenial` instead of using `lsb_release -cs` (using `bionic` doesn't seem to always works).
>
>```
>$ sudo add-apt-repository \
>   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
>   xenial \
>  stable"
>```
>
>Now try running `$ sudo apt-get install docker-ce` again.

To manage docker as a non-root user (so you don't have to keep using `sudo`) follow the instructions [here](https://docs.docker.com/install/linux/linux-postinstall/). **You must log out and log back in (or just restart your VM) so that your group membership is re-evaluated**

## Build it

* go
* make

## Run it

```bash
sudo ./pupernetes run sandbox/
```

```text
I0412 19:24:01.349686   38841 clean.go:30] Removed /home/jb/go/src/github.com/DataDog/pupernetes/sandbox/etcd-data
I0412 19:24:01.350733   38841 clean.go:136] Cleanup successfully finished
I0412 19:24:03.788224   38841 systemd.go:31] Already created systemd unit: /run/systemd/system/p8s-kubelet.service
I0412 19:24:03.788277   38841 systemd.go:31] Already created systemd unit: /run/systemd/system/p8s-etcd.service
I0412 19:24:05.277634   38841 setup.go:249] Setup ready /home/jb/go/src/github.com/DataDog/pupernetes/sandbox
I0412 19:24:05.278049   38841 run.go:95] Timeout for this current run is 6h0m0s
I0412 19:24:05.278124   38841 systemd.go:40] Starting systemd unit: p8s-etcd.service ...
I0412 19:24:06.024161   38841 systemd.go:40] Starting systemd unit: p8s-kubelet.service ...
W0412 19:24:07.034545   38841 run.go:192] Kubenertes apiserver not ready yet: Get http://127.0.0.1:8080/healthz: dial tcp 127.0.0.1:8080: connect: connection refused
W0412 19:24:21.031852   38841 run.go:192] Kubenertes apiserver not ready yet: bad status code for http://127.0.0.1:8080/healthz: 500
I0412 19:24:24.031872   38841 kubectl.go:14] Calling kubectl apply -f /home/jb/go/src/github.com/DataDog/pupernetes/sandbox/manifest-api ...
I0412 19:24:25.541348   38841 kubectl.go:21] Successfully applied manifests:
serviceaccount "coredns" created
clusterrole.rbac.authorization.k8s.io "system:coredns" created
clusterrolebinding.rbac.authorization.k8s.io "system:coredns" created
configmap "coredns" created
deployment.extensions "coredns" created
service "coredns" created
clusterrolebinding.rbac.authorization.k8s.io "p8s-admin" created
serviceaccount "kube-controller-manager" created
pod "kube-controller-manager" created
daemonset.extensions "kube-proxy" created
daemonset.extensions "kube-scheduler" created
I0412 19:24:25.541409   38841 run.go:156] Kubernetes apiserver hooks done
I0412 19:24:26.091092   38841 run.go:176] Kubelet is running 1 pods
I0412 19:24:36.035496   38841 run.go:176] Kubelet is running 4 pods
I0412 19:24:46.044459   38841 run.go:176] Kubelet is running 5 pods
```

## Use it

>Note:
>
>`kubectl` is automatically installed by `pupernetes`.
>
>You may need to run the following command to add `kubectl` to the `$PATH`:
>
>```bash
> sudo ./pupernetes run sandbox/ --kubectl-link /usr/local/bin/kubectl
>```

```bash
kubectl get svc,ds,deploy,job,po --all-namespaces
```

```text
NAMESPACE     NAME         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)         AGE
default       kubernetes   ClusterIP   192.168.254.1   <none>        443/TCP         3m
kube-system   coredns      ClusterIP   192.168.254.2   <none>        53/UDP,53/TCP   3m

NAMESPACE     NAME             DESIRED   CURRENT   READY     UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
kube-system   kube-proxy       1         1         1         1            1           <none>          3m
kube-system   kube-scheduler   1         1         1         1            1           <none>          3m

NAMESPACE     NAME      DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
kube-system   coredns   1         1         1            1           3m

NAMESPACE     NAME                       READY     STATUS    RESTARTS   AGE
kube-system   coredns-747dbcf5df-p2lhq   1/1       Running   0          3m
kube-system   kube-apiserver-v1704       1/1       Running   0          2m
kube-system   kube-controller-manager    1/1       Running   0          3m
kube-system   kube-proxy-wggdn           1/1       Running   0          3m
kube-system   kube-scheduler-92zrj       1/1       Running   0          3m
```

### Command line

The full documentation is available [here](docs).

### Quick run

```bash
make
sudo ./pupernetes run sandbox/
```

Graceful stop it with:

* SIGINT
* SIGTERM
* `--timeout`
* `curl -XPOST 127.0.0.1:8989/stop`

### Quick systemd-run

```bash
sudo systemd-run ./pupernetes run ${PWD}/sandbox
```

Graceful stop it with:

* `systemctl stop run-r${UNIT_ID}.service`
* `--timeout`
* `curl -XPOST 127.0.0.1:8989/stop`

Find any systemd-run unit with:

```bash
sudo systemctl list-units run-r*.service
```

The `DESCRIPTION` field should match the initial `{COMMAND} [ARGS...]`

### Systemd as job type

It's possible to run pupernetes as a systemd service directly with the command line.
In this case, pupernetes asks to be started with the given arguments.
See more info about it in the [run command](docs/pupernetes_run.md).

Graceful stop it with:

* `systemctl stop pupernetes.service`
* `--timeout`
* `curl -XPOST 127.0.0.1:8989/stop`

## Current limitations

* Container runtime
  * You need docker already up and running
  * You cannot use cri-containerd / crio without changing manually the systemd unit `/run/systemd/system/p8s-kubelet.service`
* Systemd
  * Currently working with systemd only
  * Could be containerized with extensive mounts
* Networking
  * The CNI bridge cannot be used yet
  * Kubernetes cluster IP range is statically set
* Secrets
  * IP SAN
    * Statically configured with the given Kubernetes cluster IP range
* Draining
  * The garbage collection can't be done without adding some delay during the drain phase
* Versions
  * You just can minimally change the version of the downloaded binaries in the state directory during the `run` phase but the compatibility isn't granted
