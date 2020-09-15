.PHONY: build
build: bin/omnicastd-amd64 bin/omnicastd-arm64

.PHONY: deps
deps:
	$(MAKE) -C gcast build
	$(MAKE) -C upnp build

bin/omnicastd-amd64: deps
	GOARCH=amd64 go build -o bin/omnicastd-amd64 cmd/omnicastd/main.go

bin/omnicastd-arm64: deps
	GOARCH=arm64 go build -o bin/omnicastd-arm64 cmd/omnicastd/main.go

.PHONY: clean
clean:
	rm -f bin/omnicastd-*

.PHONY: clean-all
clean-all: clean
	$(MAKE) -C gcast clean
	$(MAKE) -C upnp clean
