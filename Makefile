CC=go
CFLAGS?=-i
GOOS=linux
CGO_ENABLED?=0

pupernetes: go-constraint
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) $(CC) build $(CFLAGS) -o $@ cmd/main.go

go-constraint:
	go version | grep "go version go1.10"

clean:
	$(RM) pupernetes pupernetes.sha512sum

re: clean pupernetes

gofmt:
	./scripts/update/gofmt.sh

docs:
	$(CC) run ./scripts/update/docs.go

license:
	./scripts/update/license.sh

goget:
	@which ineffassign || go get github.com/gordonklaus/ineffassign
	@which golint || go get golang.org/x/lint/golint
	@which misspell || go get github.com/client9/misspell/cmd/misspell
	@which wwhrd || go get github.com/frapposelli/wwhrd

# Private targets
PKG=.cmd .pkg .docs
$(PKG): %:
	@# remove the leading '.'
	ineffassign $(subst .,,$@)
	golint -set_exit_status $(subst .,,$@)/...
	misspell -error $(subst .,,$@)

check:
	$(CC) test -v ./pkg/...

verify-misc: goget $(PKG)

verify-gofmt:
	./scripts/verify/gofmt.sh

verify-docs:
	./scripts/verify/docs.sh

verify-license: goget
	./scripts/verify/license.sh

verify: verify-misc verify-gofmt verify-docs verify-license

sha512sum: pupernetes
	$@ ./$^ > $^.$@

pupernetes-docker:
	docker run --rm --net=host -v $(PWD):/go/src/github.com/DataDog/pupernetes -w /go/src/github.com/DataDog/pupernetes golang:1.10 make

ci-validation:
	./.ci/pupernetes-validation.sh

ci-sonobuoy:
	./.ci/sonobuoy.sh

.PHONY: clean re gofmt docs license check verify-gofmt verify-docs verify-license verify sha512sum ci-validation ci-sonobuoy go-constraint
