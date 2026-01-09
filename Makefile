# Makefile for atc_freq project

# Environment variables for cgo generation
# We use g++ because SimConnect.h uses C++ features
export CC = g++
export CGO_CFLAGS = -x c++ -I$(CURDIR)

# Tools
GO = go

.PHONY: all generate build clean

all: generate build

# Run godefs to update defs_generated.go
# Running from root so -I. points to SimConnect.h in the root
generate:
	@echo "Generating definitions..."
	$(GO) tool cgo -godefs internal/sim/defs.go > internal/sim/defs_generated.go

# Build the main application
build:
	@echo "Building application..."
	$(GO) build -o main.exe ./cmd/app/main.go

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -f main.exe
	rm -f internal/sim/defs_generated.go
	rm -rf _obj
	rm -rf internal/sim/_obj
