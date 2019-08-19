# pupernetes - p8s

[![CircleCI](https://circleci.com/gh/DataDog/pupernetes.svg?style=svg)](https://circleci.com/gh/DataDog/pupernetes) [![Build Status](https://travis-ci.org/DataDog/pupernetes.svg?branch=master)](https://travis-ci.org/DataDog/pupernetes) [![Go Report Card](https://goreportcard.com/badge/github.com/DataDog/pupernetes)](https://goreportcard.com/report/github.com/DataDog/pupernetes)

pupernetes (a play on “Kubernetes” and “puppy”) is a tool written at [Datadog](https://www.datadoghq.com/) for spinning up a full-fledged Kubernetes environment for local development and CI environments similar to other tools like [minikube](https://github.com/kubernetes/minikube) but with a few more [features](#features). pupernetes was originally designed to perform e2e testing of the [Datadog Agent](https://github.com/DataDog/datadog-agent).

## Table of Contents
- [Features](#features)
- [Requirements](#requirements)
  * [Runtime](#runtime)
    * [Executables](#executables)
    * [Systemd](#systemd)
    * [Resources](#resources)
    * [DNS](#dns)
  * [Development](#development)
    * [Build](#build)
- [Getting started](#getting-started)
  * [Download](#download)
  * [Run](#run)
  * [Stop](#stop)
  * [Hyperkube versions](#hyperkube-versions)
  * [Container runtimes](#container-runtimes)
  * [Systemd as job type](#systemd-as-job-type)
  * [Command line docs](#command-line-docs)
- [Metrics](#metrics)
- [Current limitations](#current-limitations)

## Features

The goal of pupernetes is to be a smarter "Makefile" to setup, run, and clean up a full-fledged Kubernetes environment using any combination of the supported versions of Kubernetes, etcd, container runtime, and CNI plugin to validate any software project on top of it. Additionally, pupernetes provides user-friendly features like:

- Probing the control plane components (including `coredns`) during startup so you can use `kubectl` immediately after pupernetes has started.
- Complete clean up of the Kubernetes environment to leave your laptop in the same state it was in before running pupernetes.

**Provides**:
* etcd v3
* kubectl
* kubelet
* kube-apiserver
* kube-scheduler
* kube-controller-manager
* kube-proxy
* coredns
* containerd (if specified with `--container-runtime=containerd`)

**The default setup is secured with:**
* Valid x509 certificates provided by an embedded vault PKI
    * Able to use the Kubernetes CSR and the service account root-ca
* HTTPS webhook to provide token lookups for the kubelet API
* RBAC

You can use pupernetes to validate a software dependency on Kubernetes itself or just to run some app workflows with [argo](https://github.com/argoproj/argo).

As pupernetes runs in [travis](./.travis.yml) and [circle-ci](./.circleci/config.yml), it becomes very easy to integrate this tool in any Kubernetes project.

![img](docs/pupernetes.jpg)

## Requirements

### Runtime

#### Executables

* `tar`
* `unzip`
* `systemctl`
* `systemd-resolve` (or a non-systemd managed `/etc/resolv.conf`)
* `mount`
* `iptables`
* `nsenter`
* `libseccomp2` (if using containerd)

Additionally any implicit requirements needed by the **kubelet**, like the container runtime and [more](https://github.com/kubernetes/kubernetes/issues/26093).
Currently only reporting `docker`, please see the [current limitations](#current-limitations).

#### Docker

If you're using `Docker` as the container runtime, you must already have Docker installed.

#### Systemd

A recent systemd version is better to gain:
* `systemd-resolve`
* `journalctl --since`
* more convenient dbus API

#### Resources

* 4GB of memory is required
* 5GB of free disk space for the binaries and the container images

#### DNS

Ensure your hostname is discoverable:
```bash
dig $(hostname) +short
```

### Development

pupernetes must be run on linux (or linux VM).

Please see our [ubuntu 18.04](./environments/ubuntu/README.md) notes about it.

To compile pupernetes, you need the following binaries:

* `go` **1.10**
* `make`

#### Build

```bash
go get -u github.com/DataDog/pupernetes
cd ${GOPATH}/src/github.com/DataDog/pupernetes
make
```

## Getting started

### Download

You need to download the last version:
```bash
VERSION=0.10.0
curl -LOf https://github.com/DataDog/pupernetes/releases/download/v${VERSION}/pupernetes
chmod +x ./pupernetes
./pupernetes --help
```

### Run

```bash
sudo ./pupernetes daemon run /opt/sandbox/
```

>Note:
>
>`kubectl` can be automatically installed by pupernetes.
>
>You need to run the following command to add `kubectl` to the `$PATH`:
>
>```bash
> sudo ./pupernetes daemon run /opt/sandbox/ --kubectl-link /usr/local/bin/kubectl
>```

```bash
$ kubectl get svc,ds,deploy,job,po --all-namespaces

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

### Stop

Gracefully stop it with:
* SIGINT
* SIGTERM
* `--timeout`
* `curl -XPOST 127.0.0.1:8989/stop`

### Hyperkube versions

pupernetes can start a specific Kubernetes version with the flag `--hyperkube-version=1.9.3`.

These are the current supported versions:
- [x] 1.15
- [x] 1.14
- [x] 1.13
- [x] 1.12
- [x] 1.11
- [x] 1.10
- [x] 1.9
- [x] 1.8
- [x] 1.7
- [x] 1.6 (experimental)
- [x] 1.5 (experimental)
- [ ] 1.4
- [ ] 1.3

### Container runtimes

pupernetes can start a specific container runime with the flag `--container-runtime=docker`. The default is `docker`.

These are the current supported container runtimes:
- [Docker](https://github.com/docker/for-linux)
- [containerd](https://github.com/containerd/containerd) (experimental)

### Systemd as job type

It's possible to run pupernetes as a systemd service directly with the command line.
In this case, pupernetes asks to systemd-dbus to be daemonised with the given arguments.
See more info about it in the [run command](./docs/pupernetes_run.md).

This command line is very convenient to run pupernetes in SaaS CI:
* [travis](./examples/travis.yaml)
* [circle-ci](./examples/circleci.yaml)

### Command line docs

The full documentation is available [here](./docs).

## Metrics

pupernetes exposes prometheus metrics to improve the observability.

You can have a look at which metrics are available [here](./docs/metrics.csv).

## Current limitations

* Systemd
  * Currently working with systemd only
  * Could be containerized with extensive mounts
    * binaries
    * dbus
* Support for Custom Metrics
  * You can register an API Service for an External Metrics Provider.
  This is only supported for 1.10.x and 1.11.x.
