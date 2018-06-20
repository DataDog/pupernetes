CC=go
CFLAGS?=-i
GOOS=linux
CGO_ENABLED?=0

pupernetes:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) $(CC) build $(CFLAGS) -o $@ cmd/main.go

clean:
	$(RM) pupernetes
	$(RM) pupernetes.sha512sum

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

verify-license:
	./scripts/verify/license.sh

verify: verify-misc verify-gofmt verify-docs verify-license

sha512sum: pupernetes
	$@ ./$^ > $^.$@

# Everything but the pupernetes target
.PHONY: clean re gofmt docs license check verify-gofmt verify-docs verify-license verify sha512sum
