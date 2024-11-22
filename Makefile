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

# Source files
MAIN=./main.go

all: build

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) -v $(MAIN)

clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f bin/$(BINARY_UNIX)
	rm -f bin/$(BINARY_WINDOWS)

run: build
	./bin/$(BINARY_NAME)

test:
	$(GOTEST) -v ./...

# Cross compilation
build-all: build-linux build-windows

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_UNIX) -v $(MAIN)

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_WINDOWS) -v $(MAIN)