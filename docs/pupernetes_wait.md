## pupernetes wait

Wait for systemd unit name to be Running

### Synopsis

Wait for systemd unit name to be Running

```
pupernetes wait a systemd unit [flags]
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

