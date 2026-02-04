.PHONY: help run build update

# Default port
PORT ?= 8080

# Binary name
BINARY_NAME=local-clipboard

# Build output directory
BUILD_DIR=build

# Version (can be overridden: make build VERSION=1.2.3)
VERSION ?= dev

# Linker flags to inject version
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

run: ## Run the server
	@echo "Starting $(BINARY_NAME) on port $(PORT)..."
	@-go run main.go -port $(PORT)

build: ## Build for macOS, Linux and Windows
	@mkdir -p $(BUILD_DIR)
	@echo "Building v$(VERSION)..."
	@echo "  macOS (Intel)..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-mac-intel main.go
	@echo "  macOS (Silicon)..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-mac-silicon main.go
	@echo "  Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64 main.go
	@echo "  Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64.exe main.go
	@echo ""
	@echo "Binaries created in ./$(BUILD_DIR)"

update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u
	@go mod tidy
	@echo "Dependencies updated."

.DEFAULT_GOAL := help
