.PHONY: build install clean test

BINARY_NAME=broom

build:
	go build -o $(BINARY_NAME) ./cmd/broom

install: build
	mv $(BINARY_NAME) /usr/local/bin/

clean:
	go clean
	rm -f $(BINARY_NAME)

test:
	cd $(CURDIR) && go test -v ./...

.DEFAULT_GOAL := build
