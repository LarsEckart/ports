BINARY=ports
GO_DIR=.

.PHONY: fmt lint test build check install

fmt:
	cd $(GO_DIR) && gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

lint:
	cd $(GO_DIR) && golangci-lint run ./...

test:
	cd $(GO_DIR) && go test ./...

build: fmt
	cd $(GO_DIR) && go build -o $(BINARY) .

check: lint test build

install:
	cd $(GO_DIR) && go install .
