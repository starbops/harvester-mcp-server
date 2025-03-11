.PHONY: bin build run clean package test

BINARY_NAME=harvester-mcp-server
IMAGE_NAME=starbops/harvester-mcp-server

# Create the bin directory for artifacts
bin:
	mkdir -p bin

# Build the application
build: bin
	go build -o bin/$(BINARY_NAME) ./cmd/harvester-mcp-server

# Run the application
run: build
	go run ./cmd/harvester-mcp-server

# Clean the binary
clean:
	go clean
	rm -f bin/$(BINARY_NAME)

# Build Docker image
package:
	docker build -t $(IMAGE_NAME):latest .

# Run tests
test:
	go test -v ./...

# Install dependencies
deps:
	go mod tidy

# Default target
all: deps test build package