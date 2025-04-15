# Makefile for Go project

# Project variables
BINARY_NAME := gogo-files
BIN_DIR := bin
BINARY := $(BIN_DIR)/$(BINARY_NAME)

GOFLAGS := -trimpath

.PHONY: all build run dev test test-race test-specific clean fmt vet lint setup help

.DEFAULT_GOAL := help

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build $(GOFLAGS) -o $(BINARY) .
	@echo "Build complete!"

run: build
	@echo "Running $(BINARY_NAME)..."
	$(BINARY)

dev:
	@echo "Running in development mode..."
	go run main.go

test:
	@echo "Running tests..."
	go test ./...

test-race:
	@echo "Running tests with race detector..."
	go test -race ./...

test-specific:
	@echo "Usage: make test-specific TEST=TestName PKG=./path"
	go test -run $(TEST) $(if $(PKG),$(PKG),./...)

clean:
	@echo "Cleaning up..."
	@rm -rf $(BIN_DIR)
	@echo "Clean complete!"

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Vetting code..."
	go vet ./...

lint: fmt vet

setup:
	@echo "Setting up local environment..."
	@./setup.sh

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all            - Default target, builds the application"
	@echo "  build          - Build the application to bin directory"
	@echo "  run            - Build and run the application"
	@echo "  dev            - Run without building (development mode)"
	@echo "  test           - Run all tests"
	@echo "  test-race      - Run tests with race detector"
	@echo "  test-specific  - Run a specific test (TEST=TestName PKG=./path)"
	@echo "  fmt            - Format code"
	@echo "  vet            - Vet code"
	@echo "  lint           - Format and vet code"
	@echo "  clean          - Remove build artifacts"
	@echo "  setup          - Setup local environment"
	@echo "  help           - Show this help message"