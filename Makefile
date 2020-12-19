CURRENT_DIRECTORY := $(shell pwd)
PROXY_BIN_NAME := proxy-bin
PROXY_MAIN := $(CURRENT_DIRECTORY)/cmd/proxy

# build-proxy:  Build proxy binary
build-proxy:
	@echo "Building proxy binary..."
	go build -o $(PROXY_BIN_NAME) $(PROXY_MAIN)

# test: Run all unit tests.
test:
	@echo "Running unit tests"
	CURRENT_DIRECTORY=$(CURRENT_DIRECTORY) go test -cover -race -coverprofile=coverage.txt -covermode=atomic -v ./...


