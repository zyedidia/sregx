VERSION = $(shell GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/build-version.go)
GOVARS = -X main.Version=$(VERSION)

GOOS ?= $(shell go env GOHOSTOS)
GOARCH ?= $(shell go env GOHOSTARCH)

PKGEXT = $(VERSION)-$(GOOS)-$(GOARCH)

build:
	go build -trimpath -ldflags "-s -w $(GOVARS)" ./cmd/sre

install:
	go install -trimpath -ldflags "-s -w $(GOVARS)" ./cmd/sre

sre.1: man/sre.md
	pandoc man/sre.md -s -t man -o sre.1

package: build sre.1
	mkdir sre-$(PKGEXT)
	cp README.md sre-$(PKGEXT)
	cp LICENSE sre-$(PKGEXT)
	cp sre.1 sre-$(PKGEXT)
	cp sre sre-$(PKGEXT)
	tar -czf sre-$(PKGEXT).tar.gz sre-$(PKGEXT)

clean:
	rm -f sre sre.1 sre-*.tar.gz
	rm -rf sre-*/

.PHONY: build clean install package
