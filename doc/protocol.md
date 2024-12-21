# PPA Protocol Documentation

This document describes the protocol used for communication with PPA DSP devices. The protocol operates over UDP and uses a binary message format with a consistent header structure followed by message-specific payloads.

## Basic Header Structure

Every message in the PPA protocol starts with a basic header containing the following fields:

| Field           | Type    | Size (bytes) | Description |
|----------------|---------|--------------|-------------|
| MessageType    | byte    | 1            | Type of message being sent |
| ProtocolId     | byte    | 1            | Always set to 1 |
| Status         | uint16  | 2            | Status/direction of the message |
| DeviceUniqueId | [4]byte | 4            | Unique identifier of the device |
| SequenceNumber | uint16  | 2            | Sequence number for message ordering |
| ComponentId    | byte    | 1            | Target component identifier |
| Reserved       | byte    | 1            | Reserved for future use (set to 0) |

Total header size: 12 bytes

## Message Types

The protocol defines several message types for different purposes:

| Type | Value | Description |
|------|-------|-------------|
| Ping | 0     | Keep-alive and discovery messages |
| LiveCmd | 1   | Real-time control commands |
| DeviceData | 2 | Device information and status |
| PresetRecall | 4 | Load/recall preset configurations |
| PresetSave | 5  | Save current settings as preset |
| Unknown | 255  | Reserved for unknown message types |

## Status Types

Messages can have different status types indicating their direction and purpose:

### Client Status Types
- CommandClient (0x0102): Command from client to device
- RequestClient (0x0106): Request for information from client
- ResponseClient (0x0101): Response to server from client
- ErrorClient (0x0109): Error response from client
- WaitClient (0x0141): Wait/pending status from client

### Server Status Types
- CommandServer (0x0002): Command from server/device
- RequestServer (0x0006): Request for information from server
- ResponseServer (0x0001): Response to client from server
- ErrorServer (0x0009): Error response from server
- WaitServer (0x0041): Wait/pending status from server

## Message-Specific Payloads

### Preset Recall Message
Used to load specific presets on the device. Payload structure:

| Field         | Type  | Size (bytes) | Description |
|--------------|-------|--------------|-------------|
| CrtFlags     | uint8 | 1            | Control flags |
| OptFlags     | uint8 | 1            | Optional flags |
| IndexPosition| uint8 | 1            | Preset index/position |
| Reserved     | uint8 | 1            | Reserved (set to 0) |

Recall Methods:
- By Preset Index (0): Select preset using index number
- By Preset Position (2): Select preset using position value

### Device Data Messages

#### Device Data Request
Used to request device information. Payload structure:

| Field    | Type  | Size (bytes) | Description |
|----------|-------|--------------|-------------|
| CrtFlags | uint8 | 1            | Control flags |
| OptFlags | uint8 | 1            | Optional flags |

#### Device Data Response
Contains detailed device information. Payload structure:

| Field               | Type     | Size (bytes) | Description |
|--------------------|----------|--------------|-------------|
| CrtFlags           | uint8    | 1            | Control flags |
| OptFlags           | uint8    | 1            | Optional flags |
| DeviceTypeId       | uint16   | 2            | Device type identifier |
| SubnetPrefixLength | uint8    | 1            | Network subnet prefix length |
| DiagnosticState    | uint8    | 1            | Device diagnostic state |
| FirmwareVersion    | uint32   | 4            | Firmware version number |
| SerialNumber       | uint16   | 2            | Device serial number |
| Reserved           | uint32   | 4            | Reserved field |
| GatewayIP          | [4]byte  | 4            | Gateway IP address |
| StaticIP           | [4]byte  | 4            | Static IP address |
| HardwareFeatures   | uint32   | 4            | Hardware feature flags |
| StartPresetId      | uint8    | 1            | Starting preset identifier |
| Reserved2          | [6]byte  | 6            | Reserved field |
| DeviceName         | [32]byte | 32           | Device name |
| VendorID           | uint8    | 1            | Vendor identifier |

### Live Command Message
Used for real-time control of device parameters. Payload structure:

| Field      | Type     | Size (bytes) | Description |
|------------|----------|--------------|-------------|
| CrtFlags   | uint8    | 1            | Control flags (0x01 for string values) |
| OptFlags   | uint8    | 1            | Optional flags |
| Path       | [10]byte | 10           | Control path (5 pairs of position/type) |
| Value      | uint32   | 4            | Command value or string length |
| ValueString| string   | variable     | Optional string value (when CrtFlags=0x01) |

#### Live Command Path Structure
The Path field consists of 5 pairs of (position, LevelType) bytes, where:
- Position: 0-based index (first position is always 0)
- LevelType: Type of control (see Level Types below)

#### Level Types and Hierarchy
Level types can be used in specific hierarchical relationships:

1. Top Level:
   - Input (1)
   - Output (2)

2. Under Input/Output:
   - Output (2) - under Input only
   - Eq (3)
   - Gain (4)
   - Mute (9)
   - Delay (10)
   - Phase Inversion (11)

3. Under Eq:
   - Gain (4)
   - EqType (5)
   - Quality (7)
   - Active (8)

#### Equalizer Types
| Type    | Value | Description |
|---------|-------|-------------|
| LP6     | 0     | Low Pass 6dB/oct |
| LP12    | 1     | Low Pass 12dB/oct |
| HP6     | 2     | High Pass 6dB/oct |
| HP12    | 3     | High Pass 12dB/oct |
| Bell    | 4     | Bell/Parametric |
| LS6     | 5     | Low Shelf 6dB/oct |
| LS12    | 6     | Low Shelf 12dB/oct |
| HS6     | 7     | High Shelf 6dB/oct |
| HS12    | 8     | High Shelf 12dB/oct |
| AP6     | 9     | All Pass 6dB/oct |
| AP12    | 10    | All Pass 12dB/oct |

#### Value Encoding
- Boolean values: 0 = false, 1 = true
- Gain values: value = 10 * gain(dB) + 800
- String values: CrtFlags = 0x01, Value = string length

## Communication Flow

1. Device Discovery:
   - Clients send ping messages with StatusRequestServer
   - Devices respond with ping messages with StatusResponseServer

2. Command Sequence:
   - Client sends command with appropriate MessageType and StatusCommandClient
   - Device processes command and responds with StatusResponseServer
   - If processing takes time, device may respond with StatusWaitServer
   - Errors are indicated with StatusErrorServer/StatusErrorClient

3. Volume Control:
   - Uses DeviceData message type
   - Gain values range from 0 (minimum) to 0x3e8 (maximum)
   - 0dB corresponds to gain value 1.0
   - -72dB corresponds to gain value 0.0

## Implementation Notes

- All multi-byte values are encoded in little-endian format
- Maximum packet size is typically 1024 bytes
- Communication timeout is typically set to 10 seconds
- Sequence numbers should be incremented for each new command
- Component IDs are used to target specific parts of the device