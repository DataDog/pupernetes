# pupernetes - p8s

Run a managed Kubernetes setup.

This project purpose is to provide a simple Kubernetes setup to validate any software on top of it.

You can use it to validate a software dependence on Kubernetes itself or just to run some classic app workflows with [argo](https://github.com/argoproj/argo)

Our main use-case is the end to end testing pipeline of the [datadog-agent](https://github.com/DataDog/datadog-agent)

[![asciicast](https://asciinema.org/a/K1Q2eBlCpYfpqXBjOmXGmx4Jy.png)](https://asciinema.org/a/K1Q2eBlCpYfpqXBjOmXGmx4Jy)

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

* valid x509 certificates provided by an embedded vault PKI
    * Able to use the Kubernetes CSR and the service account root-ca
* HTTPS webhook to provide token lookups for the kubelet API
* RBAC

## Table of Contents
- [Requirements](#requirements)
  * [Runtime](#runtime)
  * [Development](#development)
    + [Install VMware Fusion](#install-vmware-fusion)
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

`pupernetes` must be run on linux. To run a linux VM install [VMware Fusion](https://www.vmware.com/products/fusion/fusion-evaluation.html) or your preferred virtualization software. You can use the VMware Fusion Pro 30-day trial.

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
sudo ./pupernetes daemon run sandbox/
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
kube-system   kube-controller-manager    1/1       Running   0          3m
kube-system   kube-proxy-wggdn           1/1       Running   0          3m
kube-system   kube-scheduler-92zrj       1/1       Running   0          3m
```

## Hyperkube version:

Example: `--hyperkube-version=1.9.3`

- [x] 1.11
- [x] 1.10
- [x] 1.9
- [x] 1.8
- [x] 1.7
- [x] 1.6
- [x] 1.5
- [] 1.4
- [] 1.3

### Command line

The full documentation is available [here](docs).

### Quick run

```bash
make
sudo ./pupernetes daemon run sandbox/
```

Graceful stop it with:

* SIGINT
* SIGTERM
* `--timeout`
* `curl -XPOST 127.0.0.1:8989/stop`

### Quick systemd-run

```bash
sudo systemd-run ./pupernetes daemon run ${PWD}/sandbox
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
* Versions
  * You just can minimally change the version of the downloaded binaries in the state directory during the `run` phase but the compatibility isn't granted
