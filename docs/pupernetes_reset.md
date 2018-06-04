## pupernetes reset

Reset the Kubernetes resources in the given namespace

### Synopsis

Reset the Kubernetes resources in the given namespace

```
pupernetes reset [namespaces ...] [flags]
```

### Examples

```

# Reset the namespace default:
pupernetes reset default

# Reset the namespace kube-system and redeploy the initial setup:
pupernetes reset kube-system --apply

# Reset the namespaces default and kube-system then redeploy the initial setup:
pupernetes reset default kube-system --apply

# Reset all namespaces and redeploy the initial setup:
pupernetes reset default $(kubectl get ns -o name) --apply

```

### Options

```
      --api-address string   Address for the API ip:port (default "127.0.0.1:8989")
  -a, --apply                Apply manifests-api after reset, useful when resetting kube-system namespace
  -h, --help                 help for reset
```

### Options inherited from parent commands

```
  -v, --verbose int   verbose level (default 2)
```

### SEE ALSO

* [pupernetes](pupernetes.md)	 - Use this command to manage a Kubernetes local environment

