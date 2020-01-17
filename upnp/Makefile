.PHONY: build
build: scpd/statik.go

.PHONY: clean
clean:
	rm -f scpd/statik.go

scpd/statik.go:
	statik -src scpd/ -dest . -p scpd -c "Package scpd contains service definitions."
