# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ATC Frequency Finder is a Windows desktop application for Microsoft Flight Simulator that retrieves airport frequencies, weather information, and cloud density data via the SimConnect SDK. It provides both a GUI (Wails) and CLI interface.

## Build Commands

```bash
# Run tests
make test

# Build Wails GUI application (outputs freq_app.exe)
make build

# Build CLI application (outputs cli.exe)
make build-cli

# Full build (test, build both apps)
make all

# Clean build artifacts
make clean
```

### Running a Single Test

```bash
go test -v -run TestName ./internal/...
```

### Development Mode (Wails hot reload)

```bash
wails dev
```

## Architecture

### Layer Structure

```
App (internal/app)         - Application facade, Wails bindings
    ↓
Service (internal/sim)     - Business logic, timeout handling, input validation
    ↓
Client (internal/sim)      - SimConnect protocol implementation, data parsing
    ↓
Connection (msfs-connect)  - SimConnect.dll wrapper, separate module
```

### Key Components

- **Connection interface** (`simconnect.Connection` from `github.com/sergey-podunov/msfs-connect/simconnect`, lives in `../msfs-connect`): Abstraction over SimConnect DLL calls. Has two implementations:
  - `DllConnection` - Real SimConnect.dll calls via windows syscalls
  - `MockConnection` - For unit testing with testify/mock

- **Client** (`internal/sim/client.go`): Implements SimConnect protocols for:
  - Airport frequency retrieval via facility data requests
  - Weather observation via METAR parsing
  - Cloud density via coordinate-based grid sampling

- **Service** (`internal/sim/service.go`): Wraps Client with input validation and consistent timeout handling (10s default)

- **App** (`internal/app/app.go`): Thin facade exposing methods to Wails frontend. Methods are bound via `wails.Run()` in `wails_root.go`

### SimConnect Integration

- The SimConnect layer (`Connection`, `DllConnection`, `MockConnection`, generated `SIMCONNECT_*` types and consts, `ExceptionName`, `CString`/`TrimCString`) lives in the separate module `msfs-connect` (`../msfs-connect`)
- Local development uses the Go workspace `../go.work`, which points at both `atc_freq` and `msfs-connect`
- Struct definitions are generated in `msfs-connect` (`make generate` there) from SimConnect.h
- SimConnect.dll is loaded dynamically at runtime via `windows.LoadDLL`
- DLL must be available in PATH or local directory (from the MSFS SDK; not tracked in git, see README)

### Frontend

- Wails v2 with TypeScript frontend in `wails/frontend/`
- Vite build system
- Go bindings auto-generated to `wails/frontend/wailsjs/`

## Testing

Tests use a mock connection (`MockConnection`) with testify/mock. Test helpers in `internal/testutil/` create SimConnect response structures for testing.

## Environment Requirements

- Windows only (uses windows syscalls and SimConnect.dll)
- Go 1.25+
- Microsoft Flight Simulator with SimConnect.dll
- Wails CLI for GUI builds

## Misc
Don't edit defs_generated.go directly. When some new consts from SimConnect.h are needed — add them to `msfs-connect/simconnect/defs.go` and run `make generate` in `msfs-connect`.
