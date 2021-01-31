VERSION = $(shell GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) \
	go run tools/build-version.go)
GOVARS = -X main.Version=$(VERSION)

GOOS ?= $(shell go env GOHOSTOS)
GOARCH ?= $(shell go env GOHOSTARCH)

PKGEXT = $(VERSION)-$(GOOS)-$(GOARCH)

build:
	go build -trimpath -ldflags "-s -w $(GOVARS)" ./cmd/sregx

install:
	go install -trimpath -ldflags "-s -w $(GOVARS)" ./cmd/sregx

sregx.1: man/sregx.md
	pandoc man/sregx.md -s -t man -o sregx.1

package: build sregx.1
	mkdir sregx-$(PKGEXT)
	cp README.md sregx-$(PKGEXT)
	cp LICENSE sregx-$(PKGEXT)
	cp sregx.1 sregx-$(PKGEXT)
	cp sregx sregx-$(PKGEXT)
	tar -czf sregx-$(PKGEXT).tar.gz sregx-$(PKGEXT)

clean:
	rm -f sregx sregx.1 sregx-*.tar.gz
	rm -rf sregx-*/

.PHONY: build clean install package
