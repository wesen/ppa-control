# PPA Control Web Interface

A web-based interface for controlling PPA DSP devices, built using Go, htmx, and Bootstrap.

## Features

- Device Connection: Set and manage destination IP address
- Command Interface:
  - Ping device
  - Recall presets (16 buttons)
  - Volume control slider
- Real-time Log Window: View command responses and device communication

## Prerequisites

1. Go 1.x or higher
2. templ - Install with:
   ```bash
   go install github.com/a-h/templ/cmd/templ@latest
   ```

## Building

1. Generate template files:
   ```bash
   templ generate
   ```

2. Build the binary:
   ```bash
   go build
   ```

## Usage

1. Start the server:
   ```bash
   ./ppa-web
   ```

2. Open a web browser and navigate to:
   ```
   http://localhost:8080
   ```

3. Enter the IP address of your PPA device and click "Set IP"

4. Use the interface to:
   - Ping the device
   - Recall presets
   - Control volume
   - Monitor device communication in the log window

## Development

The project structure:
- `main.go` - Main server and route handlers
- `templates/` - templ template files
  - `base.templ` - Base layout template
  - `index.templ` - Main page components
- `static/` - Static assets (if needed)

## Environment Variables

- `PORT` - Server port (default: 8080) 