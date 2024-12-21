# PPA Protocol Tutorial

The PPA protocol is a binary protocol designed for device communication and control. This tutorial explains its structure and usage.

## Message Structure

Every PPA message consists of two parts:
1. A Basic Header (mandatory)
2. A Message-specific payload (optional)

### Basic Header

The Basic Header is 12 bytes long and structured as follows:

```
Byte  0     : Message Type    (1 byte)
Byte  1     : Protocol ID     (1 byte, always 1)
Bytes 2-3   : Status         (2 bytes)
Bytes 4-7   : Device ID      (4 bytes)
Bytes 8-9   : Sequence Number (2 bytes)
Byte  10    : Component ID    (1 byte)
Byte  11    : Reserved       (1 byte)
```

#### Message Types

The protocol supports several message types:
- `Ping` (0): Used for device discovery and keepalive
- `LiveCmd` (1): Real-time control commands
- `DeviceData` (2): Device information and status
- `PresetRecall` (4): Load saved presets
- `PresetSave` (5): Save current settings as preset

#### Status Types

Status types indicate the direction and nature of the message:

Client → Server:
- `StatusCommandClient` (0x0102): Command from client
- `StatusRequestClient` (0x0106): Request for information
- `StatusResponseClient` (0x0101): Response to server
- `StatusErrorClient` (0x0109): Error condition
- `StatusWaitClient` (0x0141): Processing state

Server → Client:
- `StatusCommandServer` (0x0002): Command from server
- `StatusRequestServer` (0x0006): Request for information
- `StatusResponseServer` (0x0001): Response to client
- `StatusErrorServer` (0x0009): Error condition
- `StatusWaitServer` (0x0041): Processing state

## Message Types in Detail

### LiveCmd

LiveCmd messages control device parameters in real-time. They use a path-based addressing system:

```
Bytes 0-1   : Control Flags
Bytes 2-11  : Path (5 pairs of position/level type)
Bytes 12-15 : Value
Optional    : String value (if CrtFlags indicates string)
```

The path system uses a hierarchical structure:
1. Top Level: Input/Output selection
2. Channel Level: Specific input/output channel
3. Processing Level: EQ, Gain, etc.

Example path components:
- Input (1): Top-level input selection
- Output (2): Top-level output selection
- EQ (3): Equalizer settings
- Gain (4): Gain control
- EQ Type (5): Equalizer type selection
- Quality (7): Q-factor for EQ
- Active (8): Enable/disable state
- Mute (9): Mute control
- Delay (10): Delay settings
- Phase Inversion (11): Phase control

### DeviceData

DeviceData messages carry device information:

```
Bytes 0-1   : Control Flags
Bytes 2-3   : Device Type ID
Byte  4     : Subnet Prefix Length
Byte  5     : Diagnostic State
Bytes 6-9   : Firmware Version
Bytes 10-11 : Serial Number
Bytes 12-15 : Reserved
Bytes 16-19 : Gateway IP
Bytes 20-23 : Static IP
Bytes 24-27 : Hardware Features
Byte  28    : Start Preset ID
Bytes 29-34 : Reserved
Bytes 35-66 : Device Name (32 bytes)
Byte  67    : Vendor ID
```

### PresetRecall

PresetRecall messages load saved device configurations:

```
Byte 0: Control Flags
Byte 1: Option Flags
Byte 2: Index/Position
Byte 3: Reserved
```

Recall modes:
- By Index (0): Select preset by number
- By Position (2): Select preset by position

## Example Usage

Here's a practical example of setting an input channel's gain:

```go
// Create a LiveCmd to set Input 1's gain to -3dB
cmd := NewLiveCmd(
    WithPath(
        NewLiveCmdTuple(0, LevelTypeInput),
        NewLiveCmdTuple(1, LevelTypeGain),
    ),
    WithGain(-3.0),
)
```

The gain value is encoded as: `value = dB * 10 + 800`
So -3dB becomes: (-3 * 10) + 800 = 770

## Best Practices

1. Always check message types and status codes
2. Maintain sequence numbers for message tracking
3. Handle unknown message types gracefully
4. Validate paths before sending LiveCmd messages
5. Keep track of device timeouts (recommended: 30 seconds) 