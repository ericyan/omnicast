.PHONY: build
build: internal/scpd/statik.go

.PHONY: clean
clean:
	rm -f internal/scpd/statik.go

internal/scpd/statik.go:
	statik -src internal/scpd/ -dest internal/ -p scpd -c "Package scpd contains service definitions."
