# Makefile for atc_freq project

# Environment variables for cgo generation
# We use g++ because SimConnect.h uses C++ features
export CC = g++
export CGO_CFLAGS = -x c++ -I$(CURDIR)
export TMP := $(or $(TMP),$(TEMP),$(TMPDIR),/tmp)

# Tools
GO = go
WAILS = wails

.PHONY: all generate test build build-cli clean

all: generate test build build-cli

# Run godefs to update defs_generated.go
# Running from root so -I. points to SimConnect.h in the root
generate:
	@echo "Generating definitions..."
	$(GO) tool cgo -godefs internal/sim/defs.go > internal/sim/defs_generated.go

# Compile and run tests
test: generate
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
	rm -f internal/sim/defs_generated.go
	rm -rf _obj
	rm -rf internal/sim/_obj
	rm -rf wails/frontend/wailsjs
