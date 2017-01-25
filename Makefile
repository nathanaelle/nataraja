#!/usr/bin/make -f
# -*- makefile -*-

# basic Makefile
OS=linux
ARCH=amd64

export	GOOS=$(shell [ "x${OS}" != "x" ] && echo ${OS} || (go env GOOS) )
export	GOARCH=$(shell [ "x${ARCH}" != "x" ] && echo ${ARCH} || (go env GOARCH) )

#export	GO15VENDOREXPERIMENT=1

.PHONY: build

all: update build

update:
	go get -u golang.org/x/crypto/ocsp
	go get -u gopkg.in/vmihailenco/msgpack.v2
	go get -u github.com/boltdb/bolt
	go get -u golang.org/x/net/http2
	go get -u github.com/go-fsnotify/fsnotify
	go get -u github.com/nathanaelle/syslog5424
	go get -u github.com/nathanaelle/useful.types
	go get -u github.com/nathanaelle/pasnet
	go get -u github.com/naoina/toml


build:
	@echo building for ${GOOS}/${GOARCH}${GOARM}
	go build -o nataraja src/*.go
