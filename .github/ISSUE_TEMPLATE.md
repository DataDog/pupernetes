**Describe what happened:**


**Describe what you expected:**


**Steps to reproduce the issue:**


**Additional environment details:**

```bash
systemctl --version
```

```bash
docker search datadog/agent
```

```bash
ip a
```

```bash
df -h
```

```bash
docker ps -a
```

```bash
systemctl cat e2e-kubelet.service
```

```bash
systemctl cat e2e-etcd.service
```

```bash
tree ${state-dir}
```

```bash
ls -lR ${state-dir}/
cat ${state-dir}/manifest-api/*
cat ${state-dir}/manifest-config/*
cat ${state-dir}/manifest-static-pod/*
cat ${state-dir}/net.d/*
```


```bash
journalctl -u e2e-kubelet.service --no-pager -o cat
```

```bash
journalctl -u e2e-etcd.service --no-pager -o cat
```