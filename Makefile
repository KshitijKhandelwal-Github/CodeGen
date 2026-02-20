.PHONY: build test clean install lint fmt help

# Build variables
BINARY_NAME=codegen
VERSION?=0.1.0
BUILD_DIR=bin
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

# Default target
.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the CLI binary
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/cli

## build-all: Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/cli
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/cli
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/cli
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/cli
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/cli

## test: Run all tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

## lint: Run linters
lint:
	@echo "Running linters..."
	golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	gofmt -s -w .

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) ./cmd/cli

## run: Run the CLI (pass args with ARGS="...")
run:
	@$(GO) run ./cmd/cli $(ARGS)

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

## update-deps: Update dependencies
update-deps:
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy

.PHONY: release
release:
	@echo "Building release binaries..."
	@./scripts/build-release.sh $(VERSION)

.PHONY: install
install: build
	@echo "Installing codegen..."
	@mkdir -p $(HOME)/.local/bin
	@cp bin/codegen $(HOME)/.local/bin/codegen
	@echo "✓ Installed to $(HOME)/.local/bin/codegen"
	@echo ""
	@echo "Add to PATH (if not already):"
	@echo '  export PATH="$$PATH:$$HOME/.local/bin"'

.PHONY: uninstall
uninstall:
	@rm -f $(HOME)/.local/bin/codegen
	@rm -f /usr/local/bin/codegen
	@echo "✓ Uninstalled codegen"