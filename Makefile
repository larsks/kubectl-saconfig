PKG=github.com/larsks/kubectl-saconfig
EXE=build/kubectl-saconfig-$(shell go env GOOS)-$(shell go env GOARCH)

# Generate a list of all the source files in our package
GOSRC = $(shell go list -f '{{ $$dir := .Dir }}{{ range .GoFiles }}{{ $$dir }}/{{ . }}{{ end }}' ./...)

VERSION = $(shell git describe --tags 2> /dev/null || echo unknown)
COMMIT = $(shell git rev-parse --short=10 HEAD || echo unknown)
DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%S")

GOLDFLAGS = \
	    -w -s \
	    -X '$(PKG)/version.BuildVersion=$(VERSION)' \
	    -X '$(PKG)/version.BuildRef=$(COMMIT)' \
	    -X '$(PKG)/version.BuildDate=$(DATE)'

prefix=/usr/local
bindir=$(prefix)/bin

all: $(EXE)

install: all
	install -d -m 755 $(DESTDIR)$(bindir)
	install -m 755 $(EXE) $(DESTDIR)$(bindir)/kubectl-saconfig

$(EXE): $(dir $(EXE)) $(GOSRC)
	go build -o $@ -ldflags "$(GOLDFLAGS)" ./cmd/saconfig/

$(dir $(EXE)):
	mkdir $@

clean:
	rm -rf build

krew:
	krew-release "$(VERSION)" krew/krew.yaml.tmpl

.PHONY: clean realclean install all krew
