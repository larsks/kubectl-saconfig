PKG=github.com/larsks/kubectl-saconfig
EXE=kubectl-saconfig-$(shell go env GOOS)-$(shell go env GOARCH)

GOSRC =  cmd/saconfig/main.go \
	 version/version.go

VERSION = $(shell git describe --tags 2> /dev/null || echo unknown)
COMMIT = $(shell git rev-parse --short=10 HEAD)
DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%S")

GOLDFLAGS = \
	    -w -s \
	    -X '$(PKG)/version.BuildVersion=$(VERSION)' \
	    -X '$(PKG)/version.BuildRef=$(COMMIT)' \
	    -X '$(PKG)/version.BuildDate=$(DATE)'

all: build/$(EXE)

build/$(EXE): build $(GOSRC)
	go build -o $@ -ldflags "$(GOLDFLAGS)" ./cmd/saconfig/

build:
	mkdir build

clean:
	rm -rf build

realclean: clean
	rm -f krew/krew.yaml

krew/krew.yaml:
	krew-release "$(VERSION)" krew/krew.yaml.tmpl -o $@
