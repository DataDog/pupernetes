CC=go
CFLAGS?=-i
GOOS=linux
CGO_ENABLED?=1

.PHONY: pupernetes clean re fmt

pupernetes:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) $(CC) build $(CFLAGS) -o $@ cmd/main.go

clean:
	$(RM) pupernetes
	$(RM) pupernetes.sha512sum

re: clean pupernetes

gofmt:
	./scripts/update/gofmt.sh

gen-docs:
	$(CC) run ./scripts/update/docs.go

gen-license:
	./scripts/update/license.sh

PKG=job options setup util
$(PKG): %:
	$(CC) test -v ./pkg/$@
check: $(PKG)

verify-gofmt:
	./scripts/verify/gofmt.sh

verify-docs:
	./scripts/verify/docs.sh

verify-license:
	./scripts/verify/license.sh

verify: verify-gofmt verify-docs verify-license

sha512sum: pupernetes
	$@ ./$^ > $^.$@
