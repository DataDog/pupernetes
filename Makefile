CC=go
CFLAGS?=-i
GOOS=linux
CGO_ENABLED?=1

.PHONY: pupernetes clean re fmt

pupernetes:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) $(CC) build $(CFLAGS) -o $@ cmd/main.go

clean:
	$(RM) pupernetes

re: clean pupernetes

fmt:
	hack/update-gofmt.sh

gen-doc:
	$(CC) build -o $@ hack/docs.go
	./$@
	$(RM) $@

check:
	$(CC) test -v ./pkg/options/
	$(CC) test -v ./pkg/setup/
