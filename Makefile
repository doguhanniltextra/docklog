# Makefile for Docklog

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/doguhanniltextra/docklog/pkg/config.Version=$(VERSION)"

.PHONY: audit build test clean help

# Default target
all: build

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## [-a-zA-Z0-9_]+:' Makefile | sed 's/## //; s/:/	/'

## build: Build the docklog binary for the current platform
build:
	go build $(LDFLAGS) -o docklog .

## test: Run all Go tests
test:
	go test ./...

## audit: Run the Stability & Portability Audit (Cross-platform)
audit:
ifeq ($(OS),Windows_NT)
	@powershell -ExecutionPolicy Bypass -File .\scripts\verify-release.ps1
else
	@chmod +x ./scripts/verify-release.sh
	@./scripts/verify-release.sh
endif

## clean: Remove build artifacts and temporary files
clean:
ifeq ($(OS),Windows_NT)
	@if exist dist rmdir /s /q dist
	@if exist docklog.exe del /f /q docklog.exe
else
	@rm -rf dist/
	@rm -f docklog
endif
	@echo "Cleaned build artifacts."
