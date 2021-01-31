VERSION = $(shell GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/build-version.go)
GOVARS = -X main.Version=$(VERSION)

build:
	go build -trimpath -ldflags "-s -w $(GOVARS)" ./cmd/sre

install:
	go install -trimpath -ldflags "-s -w $(GOVARS)" ./cmd/sre

sre.1: man/sre.md
	pandoc man/sre.md -s -t man -o sre.1

package: build sre.1
	mkdir sre-$(VERSION)
	cp README.md sre-$(VERSION)
	cp LICENSE sre-$(VERSION)
	cp sre.1 sre-$(VERSION)
	cp sre sre-$(VERSION)
	tar -czf sre-$(VERSION).tar.gz sre-$(VERSION)

clean:
	rm -f sre sre.1 sre-*.tar.gz
	rm -rf sre-*/

.PHONY: build clean install package
