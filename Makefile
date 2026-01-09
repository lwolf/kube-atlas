.PHONY: build test e2e clean all install

BINARY=bin/atlas
PKG=./cmd/atlas

all: build test

build:
	mkdir -p bin
	go build -o $(BINARY) $(PKG)

test:
	go test -v ./...

e2e: build
	bash test_e2e.sh

install:
	go install $(PKG)

clean:
	rm -rf bin
	rm -rf test_env
	rm -f kubectl
