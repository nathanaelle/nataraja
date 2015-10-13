# basic Makefile

.PHONY: build

GOOS=$(shell go env GOOS)

all:	build

update:
	go get -u golang.org/x/net/http2
	go get -u golang.org/x/crypto/ocsp
	go get -u golang.org/x/net/idna
	go get -u gopkg.in/fsnotify.v1
	go get -u github.com/naoina/toml
	go get -u github.com/nathanaelle/syslog5424
	go get -u github.com/nathanaelle/useful.types
	go get -u github.com/nathanaelle/pasnet

build:
	go build -o nataraja.${GOOS} src/*.go
