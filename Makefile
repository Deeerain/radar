APP_NAME := radar
BUILD_DIR := build

LDFLAGS := -ldflags="-s -w"

.PHONY: all windows linux macos-intel macos-arm clean help

all: linux windows macos-intel macos-arm

windows:
	@echo "Building for Windows (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe main.go

linux:
	@echo "Building for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 main.go

macos-intel:
	@echo "Building for macOS (amd64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 main.go

macos-arm:
	@echo "Building for macOS (arm64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 main.go

clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Done."

help:
	@echo "Available commands:"
	@sed -n 's/^##//p' Makefile | column -t -s ':' | sed -e 's/^/ /'
