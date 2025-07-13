# PPA Protocol PCAP Analysis Scripts

This directory contains scripts for recording and analyzing PPA protocol traffic to aid in reverse engineering and protocol validation.

## Scripts Overview

### 📹 Recording Scripts

- **`quick-capture.sh`** - Quick packet capture for immediate analysis
- **`record-ppa-session.sh`** - Record a complete PPA session
- **`record-specific-operations.sh`** - Record individual PPA operations

### 🔍 Analysis Scripts

- **`analyze-ppa-captures.sh`** - Comprehensive analysis of captured packets
- **`test-protocol-implementation.sh`** - Test our protocol understanding
- **`protocol-fuzzer.sh`** - Edge case and robustness testing

## Quick Start

### 1. Quick Capture (30 seconds)
```bash
./quick-capture.sh 192.168.0.200 30
```

### 2. Record Specific Operations
```bash
./record-specific-operations.sh 192.168.0.200
```

### 3. Analyze Captures
```bash
./analyze-ppa-captures.sh ppa-ops-20250713
```

## Protocol Details

Based on the implementation in `lib/protocol/ppa-protocol.go`:

### Message Types
- **Ping (0)**: Device discovery and keepalive
- **LiveCmd (1)**: Real-time control commands
- **DeviceData (2)**: Device information and status
- **PresetRecall (4)**: Load saved presets
- **PresetSave (5)**: Save current settings

### Status Codes
**Client → Server:**
- `0x0102` (258) - Command
- `0x0106` (262) - Request
- `0x0101` (257) - Response
- `0x0109` (265) - Error
- `0x0141` (321) - Wait

**Server → Client:**
- `0x0002` (2) - Command
- `0x0006` (6) - Request
- `0x0001` (1) - Response
- `0x0009` (9) - Error
- `0x0041` (65) - Wait

### Control Hierarchy (LiveCmd)
- **Level Types**: Input(1), Output(2), EQ(3), Gain(4), Mute(9), Delay(10), etc.
- **Path Structure**: 5 pairs of (position, levelType)
- **Value Encoding**: Gain = `dB * 10 + 800`

## Usage Examples

### Basic Session Recording
```bash
# Record for 5 minutes while using PPA software
./record-ppa-session.sh my-session 300 192.168.0.200

# Analyze the session
./analyze-ppa-captures.sh my-session
```

### Specific Operation Analysis
```bash
# Record individual operations with prompts
./record-specific-operations.sh 192.168.0.200

# Test our protocol implementation
./test-protocol-implementation.sh 192.168.0.200
```

### Protocol Fuzzing
```bash
# Test edge cases and robustness
./protocol-fuzzer.sh 192.168.0.200
```

## File Organization

```
pcap/
├── captures/           # Raw PCAP files
├── analysis/          # Analysis results and reports
├── scripts/           # The analysis scripts
└── README.md         # This file
```

### Capture Files
- `*-full.pcap` - Complete session captures
- `*-01-connection.pcap` - Connection handshake
- `*-04-volume-control.pcap` - Volume control operations
- `*-ping.pcap` - Ping test results

### Analysis Files
- `*-message-types.txt` - Message type distribution
- `*-status-codes.txt` - Status code analysis
- `*-unknown.txt` - Unknown message investigation
- `*-summary-report.md` - Comprehensive analysis
- `*.json` - Full packet data in JSON format

## Protocol Analysis Workflow

1. **Record Operations** using the recording scripts
2. **Analyze Patterns** with the analysis scripts
3. **Validate Implementation** against our protocol code
4. **Test Edge Cases** using the fuzzer
5. **Update Documentation** based on findings

## Advanced Analysis

### Manual PCAP Analysis
```bash
# Build the pcap tool first
cd ../../..
go build ./cmd/pcap

# View all packets
./pcap --output-format text capture.pcap

# Filter specific message types
./pcap --print-packets liveCmd --print-hexdump capture.pcap

# Export to JSON for processing
./pcap --output-format json capture.pcap > analysis.json

# Message type summary
./pcap --output-format jsonl capture.pcap | jq -r '.header.message_type' | sort | uniq -c
```

### Custom Analysis
```bash
# Extract timing patterns
./pcap --output-format jsonl capture.pcap | \
  jq -r '[.timestamp, .header.message_type, .direction] | @csv' > timing.csv

# Find sequence number gaps
./pcap --output-format jsonl capture.pcap | \
  jq -r '.header.sequence_number' | sort -n | \
  awk 'NR>1 && $1!=prev+1 {print "Gap: " prev " -> " $1} {prev=$1}'

# Analyze payload sizes
./pcap --output-format jsonl capture.pcap | \
  jq -r '.payload | if . == null then 0 else (. | tostring | length) end' | \
  sort -n | uniq -c
```

## Prerequisites

- **Root/sudo access** for packet capture
- **tcpdump** for packet recording
- **jq** for JSON processing
- **PPA device** on network (default: 192.168.0.200:5001)

## Configuration

Default settings can be modified in each script:
- **Device IP**: 192.168.0.200
- **Port**: 5001
- **Capture duration**: varies by script
- **Output directories**: `captures/` and `analysis/`

## Troubleshooting

### No packets captured
- Check device IP and port
- Verify network connectivity
- Ensure device is powered on
- Check firewall settings

### Permission denied
- Run capture scripts with sudo
- Check tcpdump permissions

### Tool not found
- Scripts will build the pcap tool automatically
- Ensure Go is installed and working

## Contributing

When adding new analysis capabilities:
1. Follow the existing script naming convention
2. Update this README with new features
3. Add usage examples
4. Document any new protocol findings

## Related Files

- `../../../cmd/pcap/` - PCAP analysis tool source
- `../../../lib/protocol/` - Protocol implementation
- `../../../tutorial-protocol.md` - Protocol documentation
- `../../protocol-api-analysis.md` - Protocol analysis report
