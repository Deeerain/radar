APP_NAME := radar
BUILD_DIR := build

LDFLAGS := -ldflags="-s -w"

.PHONY: all windows linux macos-intel macos-arm clean help

all: linux windows macos-intel macos-arm

run-cli:
	@echo "Run cli version..."
	@go run cmd/cli/main.go

windows:
	@echo "Building for Windows (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/...

linux:
	@echo "Building for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/...

macos-intel:
	@echo "Building for macOS (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/...

macos-arm:
	@echo "Building for macOS (arm64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/...

clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Done."

help:
	@echo "Available commands:"
	@sed -n 's/^##//p' Makefile | column -t -s ':' | sed -e 's/^/ /'
