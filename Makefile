.PHONY: all build clean run test build-linux build-windows build-all

# Basic Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Binary names
BINARY_NAME=mCHAT
BINARY_UNIX=$(BINARY_NAME)_linux
BINARY_WINDOWS=$(BINARY_NAME).exe
BINARY_MAC=$(BINARY_NAME)_darwin

# Source files
MAIN=./main.go

all: build

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) -v $(MAIN)

clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f bin/$(BINARY_UNIX)_*
	rm -f bin/$(BINARY_WINDOWS)_*
	rm -f bin/$(BINARY_MAC)_*

run: build
	./bin/$(BINARY_NAME)

test:
	$(GOTEST) -v ./...

# Cross compilation
build-all: build-linux build-windows build-mac

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_UNIX)_amd64 -v $(MAIN)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -o bin/$(BINARY_UNIX)_arm64 -v $(MAIN)

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_WINDOWS)_amd64 -v $(MAIN)
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(GOBUILD) -o bin/$(BINARY_WINDOWS)_arm64 -v $(MAIN)

build-mac:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_MAC)_amd64 -v $(MAIN)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -o bin/$(BINARY_MAC)_arm64 -v $(MAIN)