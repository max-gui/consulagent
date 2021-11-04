SHELL := /bin/bash

PREFIX = consulagent

PACKAGES = $(shell go list ./... | grep -v /vendor/)
TESTARGS ?= -race
#os = linux or darwin
os=linux

CURRENTDIR = $(shell pwd)
SOURCEDIR = $(CURRENTDIR)
APP_SOURCES := $(shell find $(SOURCEDIR) -name '*.go' -not -path '$(SOURCEDIR)/vendor/*')

PATH := $(CURRENTDIR)/bin:$(PATH)

VERSION?=$(shell git describe --tags)

LD_FLAGS = -ldflags "-X main.VERSION=$(VERSION) -s -w"

all: build

.PHONY: clean build docker check
default: build
build: dist/config-resolver

test:
# go test -v  github.com/max-gui/consulagent/internal/confgen -test.run makeconfiglist
	go test -covermode=count -coverprofile=coverage.out -coverpkg ./... ./...
	@#workaround:https://github.com/golang/go/issues/22430
	@sed -i "s/_${PWDSLASH}/./g" coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out -o coverage.txt
	@tail -n 1 coverage.txt | awk '{print $$1,$$3}'

clean:
	rm -rf dist vendor

dist/config-resolver:
	mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=${os} go build $(LD_FLAGS) -v -o dist/${PREFIX} cmd/${PREFIX}/main.go

docker:
	docker build -t $(PREFIX):$(VERSION) .
	docker save -o dist/$(PREFIX):$(VERSION).tar $(PREFIX):$(VERSION)

$(PACKAGES): check-deps format
	go test $(TESTARGS) $@
	cd $(GOPATH)/src/$@; gometalinter --deadline  720s --vendor -D gotype -D dupl -D gocyclo -D gas -D errcheck

check-deps:
	@which gometalinter > /dev/null || \
	(go get github.com/alecthomas/gometalinter && gometalinter --install)

check: $(PACKAGES)

vendor:
	glide install --strip-vendor

format:
	goimports -w -l $(APP_SOURCES)
