# PPA Control UI

A graphical user interface for controlling PPA DSP devices, built using the [Fyne](https://fyne.io/) toolkit.

## Features

- Device Discovery: Automatically detect PPA devices on the network
- Preset Management: 16 preset buttons for quick preset recall
- Volume Control: Master volume slider with fine-grained control
- Device Monitoring: Real-time device status and connection monitoring
- Settings Management: Configure application settings through a graphical interface
- Log Management: Built-in logging with automatic log rotation and upload capability

## Installation

### Prerequisites

1. Go 1.x or higher
2. Fyne toolkit dependencies:
   ```bash
   # For Debian/Ubuntu
   sudo apt-get install gcc libgl1-mesa-dev xorg-dev
   
   # For Fedora
   sudo dnf install gcc libXcursor-devel libXrandr-devel mesa-libGL-devel libXi-devel libXinerama-devel libXxf86vm-devel
   ```

### Building

```bash
go build ./cmd/ui-test
```

## Usage

### Basic Usage

```bash
ui-test [flags]
```

### Global Flags

- `--log-level string`: Set log level (debug, info, warn, error, fatal) (default "debug")
- `--log-format string`: Log format (json, text) (default "json")
- `--with-caller`: Log caller information
- `--dump-mem-profile string`: Dump memory profile to file
- `--track-leaks`: Track memory and goroutine leaks

### Network Configuration

- `--addresses string`: Comma-separated list of device addresses
- `--discover`: Enable device discovery (default true)
- `--port uint`: Network port (default 5001)
- `--component-id uint`: Component ID for device communication (default 0xFF)

### Log Upload Configuration

- `--log-upload-api string`: API endpoint for log upload
- `--log-upload-region string`: S3 region for log upload
- `--log-upload-bucket string`: S3 bucket for log upload

## Interface Overview

### Main Window

The main window consists of several sections:

1. Preset Controls
   - 16 preset buttons arranged in a 4x4 grid
   - Each button recalls a specific device preset

2. Volume Control
   - Master volume slider (0-100%)
   - Fine-grained control with 0.01 step increments
   - Debounced updates to prevent flooding the network

3. Settings
   - Access to application configuration
   - Network settings
   - Log management

4. Status Console
   - Real-time device status updates
   - Connection events
   - Error messages

## Subcommands

### upload

Upload application logs to a configured S3 bucket.

```bash
ui-test upload [flags]
```

The upload command provides progress feedback and handles:
- Log file collection
- Configuration file backup
- Secure credential management
- Progress monitoring

## Examples

### Start with Device Discovery

```bash
ui-test --discover
```

### Connect to Specific Devices

```bash
ui-test --addresses 192.168.1.100,192.168.1.101
```

### Debug Mode with Text Logs

```bash
ui-test --log-level debug --log-format text
```

### Upload Logs

```bash
ui-test upload
```

## Log Management

Logs are automatically stored in:
- Linux: `~/.cache/Hoffmann Audio/ppa-control/logs`
- macOS: `~/Library/Caches/Hoffmann Audio/ppa-control/logs`
- Windows: `%LOCALAPPDATA%\Hoffmann Audio\ppa-control\logs`

Log files are automatically rotated with:
- Maximum size: 10MB
- Maximum backups: 3
- Maximum age: 28 days 