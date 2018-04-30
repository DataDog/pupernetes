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

fmt:
	./scripts/update-gofmt.sh

gen-docs:
	$(CC) run ./scripts/docs.go

PKG=job options setup util
$(PKG): %:
	$(CC) test -v ./pkg/$@
check: $(PKG)

verify:
	./scripts/verify-gofmt.sh
	./scripts/verify-generated-docs.sh

sha512sum: pupernetes
	$@ ./$^ > $^.$@
