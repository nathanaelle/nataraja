# basic Makefile

.PHONY: build

GOOS=$(go env GOOS)

build:
	go get -u github.com/bradfitz/http2
	go get -u github.com/naoina/toml
	go get -u gopkg.in/fsnotify.v1
	go get -u golang.org/x/crypto/ocsp

	go build -o nataraja.${GOOS} src/*.go
