PROJECT         := signy
ORG             := cnabio
BINDIR          := $(CURDIR)/bin
GOFLAGS         :=
GOBUILDTAGS     := osusergo
LDFLAGS         :=

ifeq ($(OS),Windows_NT)
	TARGET = $(PROJECT).exe
	SHELL  = cmd.exe
	CHECK  = where.exe
else
	TARGET = $(PROJECT)
	SHELL  ?= bash
	CHECK  ?= which
endif

.PHONY: build
build:
	go build $(GOFLAGS) -tags '$(GOBUILDTAGS)' -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(TARGET) github.com/$(ORG)/$(PROJECT)/cmd/...

.PHONY: install
install: build
	mv $(BINDIR)/$(TARGET) $(GOPATH)/bin

.PHONY: test
test:
	go test $(TESTFLAGS) ./...

.PHONY: lint
lint:
	golangci-lint run --config ./golangci.yml

HAS_GOLANGCI     := $(shell $(CHECK) golangci-lint)
HAS_GOIMPORTS    := $(shell $(CHECK) goimports)
GOLANGCI_VERSION := v1.23.6


.PHONY: bootstrap
bootstrap:
ifndef HAS_GOLANGCI
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOPATH)/bin $(GOLANGCI_VERSION)
endif
ifndef HAS_GOIMPORTS
	go get -u golang.org/x/tools/cmd/goimports
endif

.PHONY: e2e
e2e:
	make e2e-list

.PHONY: e2e-list
e2e-list:
	./bin/signy list docker.io/library/alpine
