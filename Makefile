.PHONY: help run build install vendor

VERSION=1.0

# Default port
PORT ?= 8080

# Binary name
BINARY_NAME=local-clipboard

# Build output directory
BUILD_DIR=build

# Linker flags to inject version
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

run: ## Run the server
	@echo "Starting $(BINARY_NAME) v$(VERSION) on port $(PORT)..."
	@-go run -mod=vendor $(LDFLAGS) main.go -port $(PORT)

build: ## Build for macOS, Linux and Windows
	@mkdir -p $(BUILD_DIR)
	@echo "Building v$(VERSION)..."
	@echo "  macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 go build -mod=vendor $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64 main.go
	@echo "  macOS (arm64)..."
	@GOOS=darwin GOARCH=arm64 go build -mod=vendor $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64 main.go
	@echo "  Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 go build -mod=vendor $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64 main.go
	@echo "  Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 go build -mod=vendor $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64.exe main.go
	@echo ""
	@echo "Binaries created in ./$(BUILD_DIR)"

install: ## Install dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed."

vendor: ## Vendorize dependencies
	@echo "Vendorizing dependencies..."
	@go mod tidy
	@go mod vendor
	@echo "Dependencies vendored."

.DEFAULT_GOAL := help
