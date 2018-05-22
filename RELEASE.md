# Release process

### Locally

Updated master branch:
```bash
git checkout master
git pull
```

Clean:
```bash
git clean . -fdx -e .idea
make clean
```

Compile statically the binary and generate the sha512sum:
```bash
CGO_ENABLED=0 make sha512sum
```

Check the shared object dependencies:
```bash
ldd pupernetes
	not a dynamic executable
echo $?
1
```

Check the sha512sum:
```bash
sha512sum -c pupernetes.sha512sum 
./pupernetes: OK
```

Update the [ignition example](environments/container-linux/ignition.yaml) especially the storage/files section:
* /opt/bin/setup-pupernetes
* /opt/bin/pupernetes.sha512sum


### Github

Submit a PR, merge it after validation.
 
Then upload `pupernetes` + `pupernetes.sha512sum` in the release page.

The minor version should be incremented by one like:
```text
0.1 -> 0.2
0.9 -> 0.10
...
```

The release must be marked as pre-release.
