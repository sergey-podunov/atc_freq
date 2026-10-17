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
simconnect.Connection      - DLL abstraction for SimConnect.dll (github.com/sergey-podunov/msfs-connect)
```

### Key Components

- **Connection interface** (`simconnect.Connection` from `github.com/sergey-podunov/msfs-connect/simconnect`): Abstraction over SimConnect DLL calls. Has two implementations:
  - `simconnect.DllConnection` (created via `simconnect.NewConnection()`) - Real SimConnect.dll calls via windows syscalls
  - `simconnect.MockConnection` - For unit testing with testify/mock

- **Client** (`internal/sim/client.go`): Implements SimConnect protocols for:
  - Airport frequency retrieval via facility data requests
  - Weather observation via METAR parsing
  - Cloud density via coordinate-based grid sampling

- **Service** (`internal/sim/service.go`): Wraps Client with input validation and consistent timeout handling (10s default)

- **App** (`internal/app/app.go`): Thin facade exposing methods to Wails frontend. Methods are bound via `wails.Run()` in `wails_root.go`

### SimConnect Integration

- SimConnect types (`simconnect.SIMCONNECT_*`), the DLL wrapper and the mock live in the `github.com/sergey-podunov/msfs-connect` module; this project has no cgo/generated code of its own
- SimConnect.dll is loaded dynamically at runtime via `windows.LoadDLL`
- DLL must be available in PATH or local directory (from the MSFS SDK; not tracked in git, see README)

### Frontend

- Wails v2 with TypeScript frontend in `wails/frontend/`
- Vite build system
- Go bindings auto-generated to `wails/frontend/wailsjs/`

## Testing

Tests use a mock connection (`simconnect.MockConnection`) with testify/mock. Test helpers in `internal/testutil/` create SimConnect response structures for testing.

## Environment Requirements

- Windows only (uses windows syscalls and SimConnect.dll)
- Go 1.25+
- Microsoft Flight Simulator with SimConnect.dll
- Wails CLI for GUI builds

## Custom Agents

### Verify Agent (`/verify`)

Use the verify agent to check if your changes are valid before committing:

```
/verify
```

This agent will:
1. Run `go build ./...` to verify compilation
2. Run `go test ./internal/...` to execute tests

The agent reports success/failure for each step and suggests fixes for any issues.

## Misc
When new SimConnect consts, types or calls are needed, add them to `msfs-connect` (see its CLAUDE.md), not to this project.

Always use Context7 MCP when I need library/API documentation, code generation, setup or configuration steps without me having to explicitly ask.
