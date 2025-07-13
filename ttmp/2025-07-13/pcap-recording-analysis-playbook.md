# PCAP Recording and Analysis Playbook for PPA Protocol Reverse Engineering

This playbook provides step-by-step instructions for recording and analyzing network packets to continue reverse engineering and validating the PPA (Powersoft Amplifier) protocol.

## Prerequisites

- PPA control software (official client)
- DSP amplifier device on the network
- Network capture tools (Wireshark or tcpdump)
- Built pcap-dump tool from `ppa-control/cmd/pcap`

## Part 1: Setting Up the Environment

### 1.1 Build the pcap-dump tool
```bash
cd ppa-control
go build ./cmd/pcap
```

### 1.2 Identify the target device
- Find the amplifier's IP address on your network
- Note the device's default port (typically UDP port ranges)
- Identify your control PC's IP address

### 1.3 Prepare capture environment
- Ensure the amplifier and PC are on the same network segment
- Close other network-intensive applications to reduce noise
- Have the official PPA control software ready

## Part 2: Recording PCAP Files

### 2.1 Start packet capture

**Using tcpdump (Linux/macOS):**
```bash
# Capture all traffic to/from amplifier IP
sudo tcpdump -i any -w ppa-session-$(date +%Y%m%d-%H%M%S).pcap host <AMPLIFIER_IP>

# Or capture specific port range if known
sudo tcpdump -i any -w ppa-session-$(date +%Y%m%d-%H%M%S).pcap host <AMPLIFIER_IP> and portrange 12000-13000
```

**Using Wireshark:**
1. Start Wireshark with admin privileges
2. Select the network interface
3. Apply filter: `ip.addr == <AMPLIFIER_IP>`
4. Start capture
5. Save as .pcap when done

### 2.2 Perform controlled actions

Record specific interactions to understand protocol behavior:

#### Basic Connection Flow
1. **Start capture**
2. **Launch official PPA software**
3. **Connect to device** - observe handshake
4. **Wait 30 seconds** - capture idle/ping traffic
5. **Stop capture** and save as `01-connection-handshake.pcap`

#### Device Discovery
1. **Start capture**
2. **Device scan/discovery** in official software
3. **Stop capture** and save as `02-device-discovery.pcap`

#### Control Operations
For each operation, create separate captures:

1. **Gain adjustment** (`03-gain-adjustment.pcap`)
   - Start capture
   - Adjust input/output gain values
   - Stop capture

2. **EQ adjustments** (`04-eq-adjustment.pcap`)
   - Start capture
   - Modify EQ settings (frequency, gain, Q)
   - Stop capture

3. **Preset operations** (`05-preset-operations.pcap`)
   - Start capture
   - Save preset
   - Load preset
   - Stop capture

4. **Mute operations** (`06-mute-operations.pcap`)
   - Start capture
   - Mute/unmute channels
   - Stop capture

5. **Real-time monitoring** (`07-realtime-monitoring.pcap`)
   - Start capture
   - Enable real-time parameter monitoring
   - Let run for 2 minutes
   - Stop capture

## Part 3: Initial Analysis

### 3.1 Quick packet overview
```bash
# Get packet count and basic info
./pcap --output-format text --print-packets deviceData,ping,liveCmd,presetRecall,unknown 01-connection-handshake.pcap | head -50

# See all packet types present
./pcap --output-format jsonl 01-connection-handshake.pcap | jq -r '.header.message_type' | sort | uniq -c
```

### 3.2 Analyze connection sequence
```bash
# Focus on initial packets
./pcap --print-hexdump --output-format text 01-connection-handshake.pcap | head -100

# Export to JSON for detailed analysis
./pcap --output-format json 01-connection-handshake.pcap > 01-connection-handshake.json
```

### 3.3 Identify message patterns
```bash
# Count message types
./pcap --output-format jsonl *.pcap | jq -r '.header.message_type' | sort | uniq -c

# Look at sequence numbers
./pcap --output-format jsonl *.pcap | jq -r '.header.sequence_number' | sort -n
```

## Part 4: Deep Analysis Workflow

### 4.1 Message type analysis

For each message type found:

1. **Extract all instances:**
```bash
./pcap --output-format jsonl --print-packets <message_type> *.pcap > analysis-<message_type>.jsonl
```

2. **Analyze payload patterns:**
```bash
# Look for payload size patterns
jq -r '.payload | length' analysis-<message_type>.jsonl | sort | uniq -c

# Extract unique payload structures
jq -r '.payload' analysis-<message_type>.jsonl | sort | uniq
```

### 4.2 Protocol validation

Compare with existing protocol implementation:

1. **Check message structure consistency:**
```bash
# Verify header structure matches lib/protocol/basic-header.go
./pcap --output-format jsonl sample.pcap | jq '.header'
```

2. **Validate known message types:**
```bash
# Compare with protocol.MessageType constants
grep -n "MessageType" ppa-control/lib/protocol/*.go
```

### 4.3 Unknown message investigation

For unknown/undocumented messages:

1. **Capture hex dumps:**
```bash
./pcap --print-hexdump --print-packets unknown mystery.pcap > unknown-analysis.txt
```

2. **Look for patterns:**
```bash
# Extract payload bytes for pattern analysis
./pcap --output-format jsonl --print-packets unknown mystery.pcap | jq -r '.hex_dump' | sort | uniq -c
```

## Part 5: Validation Testing

### 5.1 Replay validation

Test our protocol understanding by:

1. **Using ppa-cli to replicate captured commands:**
```bash
# If we see a gain change in capture
./ppa-cli connect <IP> gain set input 1 -6.5

# Compare new capture with original
```

2. **Verify response patterns match expectations**

### 5.2 Edge case testing

Create specific test scenarios:

1. **Rapid commands** - test sequence number handling
2. **Invalid values** - test error responses  
3. **Connection interruption** - test reconnection protocol
4. **Multiple clients** - test concurrent access

## Part 6: Documentation Updates

### 6.1 Update protocol documentation

Based on findings, update:
- `ppa-control/lib/protocol/` - Add new message types
- `ppa-control/tutorial-protocol.md` - Document new patterns
- `ppa-control/cmd/pcap/README.md` - Update analysis examples

### 6.2 Create test cases

```bash
# Add packet samples to test suite
cp important-captures/*.pcap ppa-control/test/fixtures/

# Update packet-handler tests
# Add validation for new message types
```

## Part 7: Advanced Analysis Techniques

### 7.1 Timing analysis
```bash
# Analyze timing patterns between request/response
./pcap --output-format jsonl session.pcap | jq -r '[.timestamp, .header.message_type, .direction] | @csv'
```

### 7.2 State machine analysis
- Map device state changes based on command sequences
- Identify required initialization sequences
- Document state dependencies

### 7.3 Error condition analysis
- Capture error scenarios deliberately
- Analyze error response formats
- Update error handling in client code

## File Organization

Store all captures and analysis in timestamped directories:

```
ppa-control/ttmp/2025-07-13/pcap-analysis/
├── captures/
│   ├── 01-connection-handshake.pcap
│   ├── 02-device-discovery.pcap
│   ├── 03-gain-adjustment.pcap
│   └── ...
├── analysis/
│   ├── message-type-analysis.json
│   ├── timing-analysis.csv
│   └── unknown-messages.txt
└── findings/
    ├── new-message-types.md
    ├── protocol-updates.md
    └── validation-results.md
```

## Automation Script Example

```bash
#!/bin/bash
# auto-analyze.sh - Automated analysis of PCAP files

for pcap in *.pcap; do
    echo "Analyzing $pcap..."
    
    # Basic analysis
    ./pcap --output-format json "$pcap" > "${pcap%.pcap}.json"
    
    # Extract message type summary
    ./pcap --output-format jsonl "$pcap" | jq -r '.header.message_type' | sort | uniq -c > "${pcap%.pcap}-summary.txt"
    
    # Look for unknown messages
    ./pcap --print-packets unknown "$pcap" > "${pcap%.pcap}-unknown.txt"
done
```

## Next Steps

1. **Execute this playbook systematically**
2. **Document all findings in ttmp/2025-07-13/pcap-analysis/**
3. **Update protocol implementation based on discoveries**
4. **Add comprehensive test coverage for new findings**
5. **Validate with real hardware**

This playbook should be updated as new insights are discovered during the reverse engineering process.
