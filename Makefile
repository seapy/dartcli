BINARY   := dartcli
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE     := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
PKG      := github.com/seapy/dartcli/pkg/dartcli

LDFLAGS := -ldflags "-X $(PKG).Version=$(VERSION) -X $(PKG).Commit=$(COMMIT) -X $(PKG).BuildDate=$(DATE)"

.PHONY: build clean install test lint

build:
	go build $(LDFLAGS) -o $(BINARY) .

install:
	go install $(LDFLAGS) .

clean:
	rm -f $(BINARY)

test:
	go test ./...

lint:
	golangci-lint run ./...

.DEFAULT_GOAL := build
