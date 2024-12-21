# PPA Control

A Go application for managing DSP (Digital Signal Processing) boards by PPA. This tool provides functionality for device discovery, preset management, and volume control through both CLI and GUI interfaces.

## Features

- Device Discovery: Automatic detection of PPA DSP boards using UDP broadcast
- Preset Management: Change and manage DSP presets
- Volume Control: Adjust master volume levels
- Multiple Interfaces:
  - Command Line Interface (CLI) for automation and scripting
  - Graphical User Interface (GUI) for user-friendly control
- Network Packet Capture: Monitor and analyze PPA protocol communications
- Device Simulation: Test functionality by simulating a PPA speaker

## Prerequisites

### Linux Dependencies
```shell
# Install required development libraries
apt-get install libpcap-dev     # For packet capture functionality
apt-get install libgl1-mesa-dev # For GUI rendering
apt-get install xorg-dev        # For X11 support
```

### Other Requirements
- Go 1.x or higher
- [Fyne](https://developer.fyne.io/started/) (for GUI development)

## Installation

1. Clone the repository:
```shell
git clone https://github.com/yourusername/ppa-control.git
cd ppa-control
```

2. Build the project:
```shell
make
```

## Usage

### GUI Application

Run the graphical interface:
```shell
go run ./cmd/ui-test
```

### CLI Commands

#### Device Discovery
Find PPA devices on your network:
```shell
go run ./cmd/ppa-cli ping --log-level info --discover
```

#### Device Simulation
Create a virtual PPA speaker for testing:
```shell
go run ./cmd/ppa-cli simulate --address 0.0.0.0 --log-level info
```

#### Packet Capture
Monitor PPA protocol communications:
```shell
go run ./cmd/pcap [options]
```

### Packet Capture Tool

The packet capture utility (`pcap-dump`) allows monitoring and analysis of PPA protocol communications. This is useful for debugging, development, and understanding device behavior.

#### Features
- Real-time packet capture and analysis
- Filtering by message types
- Detailed packet information display
- Hexdump output option
- PCAP file parsing
- Configurable capture timeout
- Network interface selection

#### Usage
```shell
pcap-dump [flags] [pcap file]
```

Flags:
- `--print-packets <types>`: Comma-separated list of packet types to display
  - Available types: deviceData, ping, liveCmd, presetRecall, unknown
  - Default: "deviceData,liveCmd,unknown,ping,presetRecall"
  - Example: `--print-packets deviceData,ping`
- `--print-hexdump`: Print hexadecimal dump of packet payloads
- `--interface <name>`: Network interface to capture packets from
- `--timeout <seconds>`: Capture timeout in seconds (0 for unlimited)
- `-h, --help`: Display help information

Examples:
```shell
# Capture live traffic from specific interface
pcap-dump --interface eth0 --print-packets deviceData,ping

# Analyze a capture file with hexdump output
pcap-dump --print-hexdump capture.pcap

# Monitor specific message types with 60-second timeout
pcap-dump --print-packets liveCmd,presetRecall --timeout 60

# Display all packet types from specific interface
pcap-dump --interface wlan0
```

#### Message Types
- `deviceData`: Device information and status
- `ping`: Keep-alive messages
- `liveCmd`: Real-time control commands
- `presetRecall`: Preset configuration changes
- `unknown`: Unrecognized message types

## Network Protocol

PPA Control uses a custom UDP-based protocol for device communication:
- Device Discovery: Broadcast packets for finding devices
- Command Messages: Control device settings and presets
- Status Updates: Monitor device state and configuration

## Development

### Project Structure
- `cmd/`: Command-line tools and entry points
  - `ppa-cli/`: Main CLI application
  - `pcap/`: Packet capture utility
  - `ui-test/`: GUI application
- `lib/`: Core libraries and utilities
  - `protocol/`: PPA protocol implementation
- `doc/`: Documentation and logs
