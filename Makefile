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

img = quay.io/ericyan/omnicast

.PHONY: images
images:
	docker build --build-arg ARCH=amd64 -t $(img):amd64 .
	docker build --build-arg ARCH=arm64 -t $(img):arm64 .

.PHONY: manifest
manifest: images
	docker manifest create $(img) $(img):amd64 $(img):arm64
	docker manifest annotate $(img) $(img):amd64 --os linux --arch amd64
	docker manifest annotate $(img) $(img):arm64 --os linux --arch arm64 --variant v8
	docker manifest push $(img)

.PHONY: clean
clean:
	rm -f bin/omnicastd-*

.PHONY: clean-all
clean-all: clean
	$(MAKE) -C gcast clean
	$(MAKE) -C upnp clean
