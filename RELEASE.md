# Release process

### Submit a PR

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

# or using docker
docker run --rm -v "$GOPATH":/go -w /go/src/github.com/DataDog/pupernetes golang:1.10 make sha512sum
```

Check the shared object dependencies:
```bash
ldd pupernetes
	not a dynamic executable
echo $?
1

# or using docker
docker run --rm -v "$GOPATH":/go -w /go/src/github.com/DataDog/pupernetes golang:1.10 sh -c 'ldd pupernetes ; echo $?'
	not a dynamic executable
1
```

Check the sha512sum:
```bash
sha512sum -c pupernetes.sha512sum 
./pupernetes: OK
```

Update the [ignition example](environments/container-linux/ignition.yaml) on the storage/files section:
* /opt/bin/setup-pupernetes
* /opt/bin/pupernetes.sha512sum


### Push tags

After validation, merge your PR and checkout the latest master branch:
```bash
git checkout master
git pull
```

Tag the release with the new version (e.g. v0.3):
```bash
git tag v0.3
git push --tags
```

The minor version should be incremented by one like:
```text
0.1 -> 0.2
0.9 -> 0.10
...
```

Then upload `pupernetes` + `pupernetes.sha512sum` in the release page.

The release must be marked as pre-release.
