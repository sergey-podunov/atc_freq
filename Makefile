# Makefile for atc_freq project

export TMP := $(or $(TMP),$(TEMP),$(TMPDIR),/tmp)

# Tools
GO = go
WAILS = wails

.PHONY: all test build build-cli clean

all: test build build-cli

# Compile and run tests
test:
	@echo "Compiling..."
	$(GO) build ./internal/...
	@echo "Running tests..."
	$(GO) test -v ./internal/...

# Build the Wails application
build: test
	@echo "Building Wails application..."
	$(WAILS) build -o freq_app.exe
	@if [ -f wails/build/bin/freq_app.exe ]; then mv wails/build/bin/freq_app.exe .; fi

# Build the CLI application
build-cli: test
	@echo "Building CLI application..."
	$(GO) build -o cli.exe ./cmd/cli

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -f freq_app.exe
	rm -f cli.exe
	rm -rf _obj
	rm -rf wails/frontend/wailsjs
