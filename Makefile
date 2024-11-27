PKG=kubectl-saconfig
EXE=kubectl-saconfig-$(shell go env GOOS)-$(shell go env GOARCH)

GOSRC =  main.go \
	 $(wildcard version/*.go)

VERSION = $(shell git describe --tags --exact-match 2> /dev/null || echo unknown)
COMMIT = $(shell git rev-parse --short=10 HEAD)
DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%S")

GOLDFLAGS = \
	    -X '$(PKG)/version.BuildVersion=$(VERSION)' \
	    -X '$(PKG)/version.BuildRef=$(COMMIT)' \
	    -X '$(PKG)/version.BuildDate=$(DATE)'

all: build/$(EXE)

build/$(EXE): build $(GOSRC)
	go build -o $@ -ldflags "$(GOLDFLAGS)"

build:
	mkdir build

clean:
	rm -rf build