## pupernetes wait

Wait for a systemd unit to be "running"

### Synopsis

Wait for a systemd unit to be "running"

```
pupernetes wait a systemd unit [flags]
```

### Examples

```

# Wait until the pupernetes.service systemd unit is running:
pupernetes wait

# Wait until the p8s-kubelet.service systemd unit is running:
pupernetes wait -u p8s-kubelet

```

### Options

```
  -h, --help                     help for wait
      --logging-since duration   Display the logs of the unit since (default 5m0s)
      --timeout duration         Timeout for wait (default 15m0s)
  -u, --unit-to-watch string     Systemd unit name to watch (default "pupernetes.service")
```

### Options inherited from parent commands

```
  -v, --verbose int   verbose level (default 2)
```

### SEE ALSO

* [pupernetes](pupernetes.md)	 - Use this command to manage a Kubernetes local environment

