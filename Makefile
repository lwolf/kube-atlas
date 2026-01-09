.PHONY: build test e2e clean all install

BINARY=atlas-bin
PKG=./cmd/atlas

all: build test

build:
	go build -o $(BINARY) $(PKG)

test:
	go test -v ./...

e2e: build
	bash test_e2e.sh

install:
	go install $(PKG)

clean:
	rm -f $(BINARY)
	rm -rf test_env
	rm -f kubectl
