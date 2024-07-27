Here's a README for the pcap dumping tool based on the provided `main.go` file:

# PPA Protocol PCAP Dumper

This tool is designed to analyze PCAP (Packet Capture) files containing PPA (used to control a DSP amplifier) protocol data. It parses and displays information about specific types of packets in a human-readable format.

## Features

- Parses PCAP files containing PPA protocol data
- Filters packets based on specified message types
- Displays detailed information about each packet, including:
  - Source and destination IP addresses and ports
  - Message type and status
  - Device unique ID and sequence number
  - Component ID
  - Specific payload data based on message type

## Usage

```
pcap-dump [-print-packets <types>] [-print-hexdump] <filename>
```

### Options

- `-print-packets`: Comma-separated list of packet types to display. Default is "deviceData,liveCmd,unknown,ping,presetRecall".
- `-print-hexdump`: If set, prints the hexdump of the packet payload.

### Packet Types

- `deviceData`: Device data messages
- `ping`: Ping messages
- `liveCmd`: Live command messages
- `presetRecall`: Preset recall messages
- `unknown`: Unknown message types

## Example

```
pcap-dump -print-packets deviceData,ping -print-hexdump capture.pcap
```

This command will analyze `capture.pcap`, displaying only device data and ping messages, and including hexdump output for each packet.

## Output

For each matching packet, the tool displays:

1. Source and destination IP addresses and ports
2. Message type and protocol details
3. Device unique ID and sequence number
4. Component ID
5. Specific payload data based on the message type (e.g., device data, live command, preset recall)

If the `-print-hexdump` flag is set, it will also display a hexdump of the packet payload.

## PPA Protocol Overview

The PPA (Protocol for Powersoft Amplifiers) is used to control DSP amplifiers. Here's a brief overview of its structure:

### Message Structure

Each PPA message consists of a header followed by a payload:

1. **Header** (12 bytes):
   - Message Type (1 byte)
   - Protocol ID (1 byte, always 1)
   - Status (2 bytes)
   - Device Unique ID (4 bytes)
   - Sequence Number (2 bytes)
   - Component ID (1 byte)
   - Reserved (1 byte)

2. **Payload** (variable length, depends on message type)

### Message Types

- `Ping` (0): Used for keep-alive messages
- `LiveCmd` (1): Used for real-time control commands
- `DeviceData` (2): Used for device information requests/responses
- `PresetRecall` (4): Used to recall preset configurations
- `PresetSave` (5): Used to save preset configurations

### LiveCmd Structure

LiveCmd messages use a path-based structure to control various aspects of the amplifier:

- Path: 5 pairs of (position, LevelType)
- LevelTypes include: Input, Output, EQ, Gain, EQ Type, Quality, Active, Mute, Delay, Phase Inversion

### Data Encoding

- Most numeric values use little-endian encoding
- Gain values are encoded as: `value = 10 * gain_in_dB + 800`

For more detailed information about the protocol, refer to the `ppa-protocol.go` file in the `lib/protocol` directory.