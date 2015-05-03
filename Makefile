

OS=linux
ARCH=amd64


.PHONY: build


build:
	go get -u github.com/bradfitz/http2
	go get -u github.com/naoina/toml
	go get -u github.com/spacemonkeygo/openssl
	go get -u gopkg.in/fsnotify.v1
	go get -u golang.org/x/crypto/ocsp

	GOOS=$(OS) GOARCH=$(ARCH) go build -o nataraja.linux src/*.go
