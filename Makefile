.PHONY: test

.ONESHELL:
SHELL = /bin/bash

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

OS := linux
ifeq ($(UNAME_S),Linux)
	OS := linux
endif
ifeq ($(UNAME_S),Darwin)
	OS := darwin
endif

ARCH := amd64
ifeq ($(UNAME_M), arm64)
	ARCH := arm64
endif

versionLabel := $(shell git rev-parse --short HEAD)

format:
	gofmt -s -w .
	goimports -w -l .

gotests:
	go test -v -failfast -cover

test: gotests

clean:
	rm -rf bin/*

my_binary: clean
	env GOOS=${OS} GOARCH=${ARCH} CGO_ENABLED=0 go build \
		-o bin/metric-explorer -ldflags="-X 'main.VersionLabel=${versionLabel}'"

darwin_arm64:
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build \
		-o bin/metric-explorer_darwin_arm64 -ldflags="-X 'main.VersionLabel=${versionLabel}'"

darwin_amd64:
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build \
		-o bin/metric-explorer_darwin_amd64 -ldflags="-X 'main.VersionLabel=${versionLabel}'"

linux_amd64:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
		-o bin/metric-explorer_linux_amd64 -ldflags="-X 'main.VersionLabel=${versionLabel}'"

build_all: clean darwin_amd64 darwin_arm64 linux_amd64
