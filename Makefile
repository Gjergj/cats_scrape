
export GO111MODULE=on

SHELL=/bin/bash -o pipefail

VERSION=0.7.0
TARGETS=linux/amd64 windows/amd64 darwin/amd64 linux/s390x
SERVICE_EXE=${SERVICE_NAME}-v${VERSION}-windows-amd64.exe

PWD = $(shell pwd)
GO ?= go

.PHONY: all
all: cats_scrape

.PHONY: cats_scrape
cats_scrape:
	$(GO) build -v ./

.PHONY: test
test:
	$(GO) vet ./...
	$(GO) test -v -failfast $(go list ./... | grep -v test) --race ./...