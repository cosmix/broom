.PHONY: build install clean test test-unit test-e2e

BINARY_NAME=broom

build:
	go build -o $(BINARY_NAME) ./cmd/broom

install: build
	mv $(BINARY_NAME) /usr/local/bin/

clean:
	go clean
	rm -f $(BINARY_NAME)
	docker rmi broom-test -f 2>/dev/null || true

test: test-unit test-e2e

test-unit:
	cd $(CURDIR) && go test -v ./...

test-e2e:
	./scripts/run-e2e-tests.sh

.DEFAULT_GOAL := build
