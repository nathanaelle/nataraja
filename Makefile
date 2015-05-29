# basic Makefile

.PHONY: build

GOOS=$(shell go env GOOS)

all:	build

update:
	go get -u github.com/bradfitz/http2
	go get -u github.com/naoina/toml
	go get -u gopkg.in/fsnotify.v1
	go get -u golang.org/x/crypto/ocsp

build:
	go build -o nataraja.${GOOS} src/*.go
