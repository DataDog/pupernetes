# Release process

Pupernetes follows Semantic Versionning ([_SemVer_](https://semver.org/)):
```text
# Patch update
0.3.0 -> 0.3.1

# Minor update
0.3.0 -> 0.4.0
0.9.0 -> 0.10.0

# Major update
0.3.0 -> 1.0.0
1.15.4 -> 2.0.0
...
```

## Local check

Clean:
```bash
git clean . -fdx -e .idea
make clean
```

Update master branch with *origin* as remote:
```bash
git fetch origin master
git checkout -B master origin/master
```

```bash
git checkout -b v0.3.0
```
> note: v0.3.0 is a example and need to be adapted

Compile statically the binary and generate the sha512sum.
```bash
CGO_ENABLED=0 make sha512sum

# or using docker
docker run --rm -v "$GOPATH":/go -w /go/src/github.com/DataDog/pupernetes golang:1.10 make sha512sum
```
> note: you need go 1.10

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

## Submit a PR

Create a PR named as example: `release-v0.3.0`

Updates:
- [releasenotes](./releasenotes.md) accordingly.
    * Features
    * Bugfix
    * Other
        Docs, CI, ...
- [ignition example](environments/container-linux/ignition.yaml) on the storage/files section:
    * /opt/bin/setup-pupernetes
    * /opt/bin/pupernetes.sha512sum
- [README](README.md)
    * [download section](README.md#download) :`${VERSION}`.

Commit and push the changes and open the PR.

## Push tags

After validation, merge your PR and checkout the latest master branch:
```bash
git checkout master
git pull
```

Verify:
```bash
cp -v kube-csr.sha512sum pr-kube-csr.sha512sum
make clean
CGO_ENABLED=0 make sha512sum
diff pupernetes.sha512sum pr-pupernetes.sha512sum
sha512sum -c pupernetes.sha512sum 
./pupernetes: OK
```

Tag the release with the new version (e.g. v0.3.0):
```bash
git tag v0.3.0
git push --tags
```

After validation, merge your PR and checkout the latest master branch:
```bash
git checkout master
git pull
```

Tag the release with the new version (e.g. v0.3.0):
```bash
git tag v0.3.0
git push --tags
```

Then upload `pupernetes` + `pupernetes.sha512sum` in the release page and select your tag.

The release must be marked as pre-release if lower than `v0.x.x`

Create the following field in the github release page:

> ### Binary
> ```
> curl -fLO https://github.com/DataDog/pupernetes/releases/download/v0.3.0/pupernetes
> curl -fLO https://github.com/DataDog/pupernetes/releases/download/v0.3.0/pupernetes.sha512sum
> sha512sum -c ./pupernetes.sha512sum
> chmod +x ./pupernetes
> ```

Append the associated fields from the [releasenotes](./releasenotes.md).
- Enhancement
- Bugfixes
- Other