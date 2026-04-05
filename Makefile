BINARY=ports

.PHONY: build test fmt install check

build:
	go build ./...

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

install:
	go install .

check: fmt build test
