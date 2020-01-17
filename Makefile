.PHONY: build
build:
	$(MAKE) -C gcast build
	$(MAKE) -C upnp build
	go build -o bin/omnicastd cmd/omnicastd/main.go

.PHONY: clean
clean:
	$(MAKE) -C gcast clean
	$(MAKE) -C upnp clean
	rm -f bin/omnicastd
