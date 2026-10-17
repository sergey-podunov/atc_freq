# ATC Frequency Finder

A Windows app for Microsoft Flight Simulator that shows airport frequencies, weather and cloud density, read from the sim through SimConnect. It has a GUI (Wails) and a CLI.

## Requirements

- Windows
- Go 1.25+
- Microsoft Flight Simulator 2020 or 2024
- `SimConnect.dll` from the MSFS SDK (see below)
- Wails CLI, for the GUI build

## Getting SimConnect.dll

The app talks to the simulator through Microsoft's `SimConnect.dll`. It's part of the MSFS SDK and isn't included in this repository.

1. In the simulator, turn on Developer Mode (Options → General → Developers).
2. In the Developer Mode toolbar, open **Help** and download the SDK installer (`.msi`), then run it.
3. Find `SimConnect.dll` in the SDK's `SimConnect SDK\lib` folder. Default locations:
   - MSFS 2024: `C:\MSFS 2024 SDK\SimConnect SDK\lib\SimConnect.dll`
   - MSFS 2020: `C:\MSFS SDK\SimConnect SDK\lib\SimConnect.dll`
4. Copy it next to `cli.exe` / `freq_app.exe` (or into the project root when running from source), or add that `lib` folder to your `PATH`.

## Build

```bash
make test        # run tests
make build-cli   # builds cli.exe
make build       # builds the Wails GUI, freq_app.exe
```

CLI usage:

```bash
cli.exe freq EDDB
cli.exe weather EDDB,UUMI,KJFK
cli.exe clouds EDDB
```
