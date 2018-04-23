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
	scripts/update-gofmt.sh

gen-doc:
	$(CC) build -o $@ scripts/docs.go
	./$@
	$(RM) $@

check:
	$(CC) test -v ./pkg/options/
	$(CC) test -v ./pkg/setup/

sha512sum: pupernetes
	$@ ./$^ > $^.$@
