# Release process

Clean:
```bash
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

Submit a PR, merge it after validation. 
Then upload `pupernetes` + `pupernetes.sha512sum` in the release page.
