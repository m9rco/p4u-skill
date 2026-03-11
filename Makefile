BINARY_NAME := p4u
VERSION     ?= dev
LDFLAGS     := -ldflags "-X main.version=$(VERSION)"

.PHONY: build build-all test clean install

# Default build for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Cross-platform builds
build-all: build-linux build-darwin-amd64 build-darwin-arm64 build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -rf dist/

install:
	go install $(LDFLAGS) .

dist:
	mkdir -p dist
