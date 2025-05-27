# PPA Control System Developer Guide

## Table of Contents
1. [System Overview](#system-overview)
2. [Architecture Deep Dive](#architecture-deep-dive)
3. [Protocol Implementation](#protocol-implementation)
4. [Building Your First Application](#building-your-first-application)
5. [Web Interface Development](#web-interface-development)
6. [Testing and Simulation](#testing-and-simulation)
7. [Advanced Topics](#advanced-topics)
8. [Troubleshooting](#troubleshooting)

---

## System Overview

Welcome to the PPA Control System! This is a comprehensive Go-based application for managing DSP (Digital Signal Processing) boards manufactured by PPA. Think of it as a universal remote control for professional audio equipment, but instead of infrared signals, we use UDP network packets with a custom binary protocol.

### What Does This System Do?

The PPA Control System allows you to:
- **Discover devices** automatically on your network using UDP broadcast
- **Control volume** in real-time with precise dB adjustments
- **Manage presets** by recalling saved audio configurations
- **Monitor device status** and receive real-time feedback
- **Build custom interfaces** using CLI, web, or mobile applications

### Why This Architecture?

The system is designed around several key principles:

**Real-time Performance**: Audio equipment needs immediate response. UDP provides low-latency communication without the overhead of TCP handshakes.

**Network Discovery**: Professional audio setups often involve multiple devices. The broadcast discovery mechanism automatically finds all PPA devices on the network.

**Modular Design**: The codebase separates protocol handling, client management, and user interfaces, making it easy to build new applications.

**Professional Reliability**: Error handling, timeouts, and graceful degradation ensure the system works reliably in live audio environments.

---

## Architecture Deep Dive

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Interface│    │   Application   │    │   PPA Devices   │
│                 │    │     Layer       │    │                 │
│  ┌─────────────┐│    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│  │   Web UI    ││    │ │ Command     │ │◄──►│ │   Speaker   │ │
│  │   CLI Tool  ││◄──►│ │ Context     │ │    │ │   System    │ │
│  │   Mobile    ││    │ │ Multi-Client│ │    │ │   DSP Board │ │
│  └─────────────┘│    │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                ▲
                                │
                       ┌─────────────────┐
                       │   Protocol      │
                       │   Layer         │
                       │ ┌─────────────┐ │
                       │ │ UDP Sockets │ │
                       │ │ Binary      │ │
                       │ │ Protocol    │ │
                       │ └─────────────┘ │
                       └─────────────────┘
```

### Core Components

#### 1. Protocol Layer (`lib/protocol/`)

This is the foundation of everything. The PPA protocol is a custom binary protocol that runs over UDP.

**Key Files:**
- `ppa-protocol.go` - Core protocol implementation
- `tutorial-protocol.md` - Protocol documentation

**Message Structure:**
Every message starts with a 12-byte header:

```go
type BasicHeader struct {
    MessageType    MessageType  // What kind of message (ping, command, etc.)
    ProtocolId     byte        // Always 1 for PPA protocol
    Status         StatusType  // Direction and purpose of message
    DeviceUniqueId [4]byte     // Identifies the target device
    SequenceNumber uint16      // For tracking request/response pairs
    ComponentId    byte        // Which component on the device
    Reserved       byte        // Future use
}
```

**Message Types:**
```go
const (
    MessageTypePing         MessageType = 0  // Device discovery and keepalive
    MessageTypeLiveCmd      MessageType = 1  // Real-time control commands
    MessageTypeDeviceData   MessageType = 2  // Device information exchange
    MessageTypePresetRecall MessageType = 4  // Load saved presets
    MessageTypePresetSave   MessageType = 5  // Save current settings
)
```

#### 2. Client Layer (`lib/client/`)

This layer manages communication with individual devices and groups of devices.

**Key Components:**

**SingleDevice** (`single-device.go`):
```go
type SingleDevice struct {
    AddrPort    string              // "192.168.1.100:5001"
    Interface   string              // Network interface name
    SendChannel chan *bytes.Buffer  // Outgoing messages
    ComponentId uint                // Device component ID
    seqCmd      uint16             // Sequence counter
}
```

**MultiClient** (`multiclient.go`):
Manages multiple devices simultaneously. This is crucial for professional setups with many speakers.

```go
type MultiClient struct {
    clients map[string]Client           // Address -> Client mapping
    cancels map[string]context.CancelFunc // Cancellation functions
    // ... other fields
}
```

**Discovery** (`discovery/discovery.go`):
Automatically finds devices on the network using UDP broadcast.

#### 3. Command Context (`lib/command-context.go`)

This is the glue that holds everything together. It provides a consistent way to set up and manage the various components.

```go
type CommandContext struct {
    Config      *CommandConfig    // Configuration (ports, addresses, etc.)
    Channels    *CommandChannels  // Communication channels
    ctx         context.Context   // Cancellation context
    multiClient *MultiClient     // Device manager
}
```

### Data Flow Example

Let's trace what happens when you adjust the volume:

1. **User Action**: User moves volume slider in web interface
2. **HTTP Request**: Browser sends POST to `/volume` with new value
3. **Handler Processing**: Web handler receives request, validates it
4. **Protocol Encoding**: Volume value is encoded into LiveCmd message
5. **UDP Transmission**: Message sent to device via UDP socket
6. **Device Response**: Device acknowledges command and updates volume
7. **UI Update**: Web interface shows confirmation

---

## Protocol Implementation

### Understanding the Binary Protocol

The PPA protocol might seem intimidating at first, but it's actually quite logical. Let's break it down with practical examples.

#### Example 1: Sending a Ping

A ping is the simplest message - just a header with no payload:

```go
func (c *SingleDevice) SendPing() {
    buf := new(bytes.Buffer)
    
    // Create the header
    bh := protocol.NewBasicHeader(
        protocol.MessageTypePing,        // This is a ping message
        protocol.StatusRequestServer,    // We're requesting from server
        [4]byte{0, 0, 0, 0},            // Device ID (0 for broadcast)
        c.seqCmd,                       // Sequence number
        byte(c.ComponentId),            // Component ID
    )
    
    c.seqCmd++ // Increment for next message
    
    // Encode header to binary
    err := protocol.EncodeHeader(buf, bh)
    if err != nil {
        log.Warn().Err(err).Msg("Failed to encode header")
        return
    }
    
    // Send it
    c.SendChannel <- buf
}
```

#### Example 2: Volume Control

Volume control is more complex because it uses the LiveCmd message type with a hierarchical path system:

```go
func (c *SingleDevice) SendMasterVolume(volume float32) {
    buf := new(bytes.Buffer)
    
    // Create header for LiveCmd
    bh := protocol.NewBasicHeader(
        protocol.MessageTypeDeviceData,  // Note: Volume uses DeviceData type
        protocol.StatusCommandClient,    // We're sending a command
        [4]byte{0, 0, 0, 0},
        c.seqCmd,
        byte(c.ComponentId),
    )
    
    // Volume encoding: 0.0 = -80dB, 1.0 = +20dB
    twentyDB := 0x3e8      // +20dB encoded value
    minusEightyDB := 0x00  // -80dB encoded value
    gain := uint32(volume * float32(twentyDB-minusEightyDB))
    
    c.seqCmd++
    
    // Encode header
    protocol.EncodeHeader(buf, bh)
    
    // Add volume-specific payload
    binary.Write(buf, binary.LittleEndian, []int8{01, 00, 03, 06})
    binary.Write(buf, binary.LittleEndian, gain)
    
    c.SendChannel <- buf
}
```

#### Example 3: Preset Recall

Recalling a preset is straightforward:

```go
func (c *SingleDevice) SendPresetRecallByPresetIndex(index int) {
    buf := new(bytes.Buffer)
    
    // Header for preset recall
    bh := protocol.NewBasicHeader(
        protocol.MessageTypePresetRecall,
        protocol.StatusCommandClient,
        [4]byte{0, 0, 0, 0},
        c.seqCmd,
        byte(c.ComponentId),
    )
    
    // Preset recall payload
    pr := protocol.NewPresetRecall(
        protocol.RecallByPresetIndex,  // Recall by index (not position)
        0,                            // Option flags
        byte(index),                  // Which preset to recall
    )
    
    c.seqCmd++
    
    // Encode both header and payload
    protocol.EncodeHeader(buf, bh)
    protocol.EncodePresetRecall(buf, pr)
    
    c.SendChannel <- buf
}
```

### Path System for LiveCmd

The LiveCmd message type uses a hierarchical path system to specify exactly what you want to control:

```
Path Structure: [Position, LevelType, Position, LevelType, ...]

Examples:
- Master Volume:    [0, Input, 0, Gain]
- Input 1 Mute:     [0, Input, 1, Mute]  
- Output 2 EQ:      [0, Output, 2, EQ, 1, Gain]
```

**LevelType Constants:**
```go
const (
    LevelTypeInput          LevelType = 1   // Input selection
    LevelTypeOutput         LevelType = 2   // Output selection
    LevelTypeEq             LevelType = 3   // Equalizer
    LevelTypeGain           LevelType = 4   // Gain control
    LevelTypeEqType         LevelType = 5   // EQ type selection
    LevelTypeQuality        LevelType = 7   // Q factor
    LevelTypeActive         LevelType = 8   // Enable/disable
    LevelTypeMute           LevelType = 9   // Mute control
    LevelTypeDelay          LevelType = 10  // Delay settings
    LevelTypePhaseInversion LevelType = 11  // Phase control
)
```

---

## Building Your First Application

Let's build a simple command-line application that discovers devices and controls their volume. This will teach you the fundamental patterns used throughout the system.

### Step 1: Basic Setup

Create a new file `examples/simple-volume-control/main.go`:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "ppa-control/lib"
    "ppa-control/lib/client"
    "time"

    "github.com/rs/zerolog/log"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "simple-volume",
    Short: "Simple volume control example",
    Run:   run,
}

func init() {
    // Add command-line flags
    rootCmd.PersistentFlags().BoolP("discover", "d", true, "Enable device discovery")
    rootCmd.PersistentFlags().Float32P("volume", "v", 0.5, "Volume level (0.0-1.0)")
    rootCmd.PersistentFlags().UintP("port", "p", 5001, "Port to use")
}

func run(cmd *cobra.Command, args []string) {
    // Get flag values
    volume, _ := cmd.PersistentFlags().GetFloat32("volume")
    
    // Setup command context - this handles all the boilerplate
    cmdCtx := lib.SetupCommand(cmd)
    defer cmdCtx.Cancel()
    
    // Setup multiclient for device management
    if err := cmdCtx.SetupMultiClient("simple-volume"); err != nil {
        log.Fatal().Err(err).Msg("Failed to setup multiclient")
        return
    }
    
    // Enable discovery to find devices automatically
    cmdCtx.SetupDiscovery()
    
    // Start the multiclient
    cmdCtx.StartMultiClient()
    
    // Main application loop
    cmdCtx.RunInGroup(func() error {
        return mainLoop(cmdCtx, volume)
    })
    
    // Wait for completion (Ctrl+C or error)
    cmdCtx.Wait()
}

func mainLoop(cmdCtx *lib.CommandContext, volume float32) error {
    // Send initial volume command to all devices
    cmdCtx.GetMultiClient().SendMasterVolume(volume)
    
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-cmdCtx.Context().Done():
            return cmdCtx.Context().Err()
            
        case <-ticker.C:
            // Send periodic ping to keep devices alive
            cmdCtx.GetMultiClient().SendPing()
            
        case msg := <-cmdCtx.Channels.ReceivedCh:
            // Handle responses from devices
            if msg.Header != nil {
                log.Info().
                    Str("from", msg.RemoteAddress.String()).
                    Str("type", msg.Header.MessageType.String()).
                    Msg("Received response")
            }
            
        case discoveryMsg := <-cmdCtx.Channels.DiscoveryCh:
            // Handle device discovery
            log.Info().
                Str("addr", discoveryMsg.GetAddress()).
                Msg("Device discovered")
                
            if newClient, err := cmdCtx.HandleDiscoveryMessage(discoveryMsg); err != nil {
                return err
            } else if newClient != nil {
                // Send volume to newly discovered device
                newClient.SendMasterVolume(volume)
            }
        }
    }
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
```

### Step 2: Running Your Application

```bash
# Build and run with discovery enabled
go run examples/simple-volume-control/main.go --discover --volume 0.7

# Run against specific device
go run examples/simple-volume-control/main.go --addresses "192.168.1.100" --volume 0.3
```

### Step 3: Understanding the Output

You'll see logs like this:

```
INFO Device discovered addr=192.168.1.100:5001
INFO Received response from=192.168.1.100:5001 type=DeviceData
INFO Client started address=192.168.1.100:5001
```

This tells you:
1. A device was discovered at 192.168.1.100
2. The device responded to your volume command
3. A client connection was established

---

## Web Interface Development

The web interface demonstrates how to build a real-time, interactive application on top of the PPA control system. It uses modern web technologies while maintaining the robust UDP communication underneath.

### Technology Stack

**Backend:**
- **Go with Gorilla Mux**: HTTP routing and middleware
- **Templ**: Type-safe HTML templating (compiles to Go code)
- **Server-Sent Events (SSE)**: Real-time updates to the browser
- **Structured logging**: Using zerolog for debugging

**Frontend:**
- **HTMX**: Dynamic HTML updates without complex JavaScript
- **Bootstrap 5**: Responsive UI components
- **Vanilla JavaScript**: For real-time features and interactivity

### Architecture Overview

```
Browser ←→ HTTP/SSE ←→ Web Server ←→ Command Context ←→ UDP ←→ PPA Devices
```

### Key Components

#### 1. Server Structure (`cmd/ppa-web/server/server.go`)

The server maintains application state and coordinates between the web interface and the UDP communication layer:

```go
type Server struct {
    state           types.AppState      // Current application state
    mu              sync.RWMutex       // Protects state access
    cmdCtx          *lib.CommandContext // UDP communication
    receiveCh       chan client.ReceivedMessage
    discoveryCtx    context.Context
    discoveryCancel context.CancelFunc
    updateListeners []chan struct{}    // For real-time updates
}
```

**State Management Pattern:**
```go
// Thread-safe state updates
func (s *Server) SetState(fn func(*types.AppState)) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    fn(&s.state)           // Apply the update
    s.notifyUpdateListeners() // Notify web clients
}

// Example usage
s.SetState(func(state *types.AppState) {
    state.DestIP = "192.168.1.100"
    state.Status = "Connected"
})
```

#### 2. HTTP Handlers (`cmd/ppa-web/handler/handler.go`)

Handlers bridge HTTP requests to UDP commands:

```go
func (h *Handler) HandleVolume(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Check if we're connected to a device
    if !h.srv.IsConnected() {
        http.Error(w, "Not connected to device", http.StatusBadRequest)
        return
    }

    // Get volume from form data
    volume := r.FormValue("volume")
    
    // Log the action
    h.srv.LogPacket("Setting volume to %s", volume)
    
    // TODO: Convert to float and send UDP command
    // volumeFloat, _ := strconv.ParseFloat(volume, 32)
    // h.srv.GetMultiClient().SendMasterVolume(float32(volumeFloat))

    // Return updated log window
    err := templates.LogWindow(h.srv.GetState()).Render(r.Context(), w)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
```

#### 3. Real-time Updates with SSE

Server-Sent Events provide real-time updates to the web interface:

```go
func (h *Handler) HandleDiscoveryEvents(w http.ResponseWriter, r *http.Request) {
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // Create notification channel
    updateCh := make(chan struct{}, 1)
    h.srv.AddUpdateListener(updateCh)
    defer h.srv.RemoveUpdateListener(updateCh)

    // Send periodic updates
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-r.Context().Done():
            return
        case <-updateCh:
            // Send updated discovery section
            var buf bytes.Buffer
            templates.DiscoverySection(h.srv.GetState()).Render(r.Context(), &buf)
            fmt.Fprintf(w, "data: %s\n\n", buf.String())
            w.(http.Flusher).Flush()
        case <-ticker.C:
            // Send keepalive ping
            fmt.Fprintf(w, "event: ping\ndata: \n\n")
            w.(http.Flusher).Flush()
        }
    }
}
```

#### 4. Templating with Templ

Templ provides type-safe HTML generation. Here's the volume control component:

```go
// templates/index.templ
templ VolumeControl(state types.AppState) {
    <h6 class="mt-3">Volume Control</h6>
    <div class="mb-3">
        <input type="range" 
               class="form-range" 
               min="0" max="100" step="1" 
               id="volume"
               hx-post="/volume"
               hx-trigger="change"
               hx-target="#log-window"
               hx-swap="innerHTML"
               name="volume"/>
        <div class="text-center" id="volume-value">50</div>
    </div>
}
```

### Building the Web Interface

#### Step 1: Generate Templates

Templ templates need to be compiled to Go code:

```bash
# Install templ if you haven't already
go install github.com/a-h/templ/cmd/templ@latest

# Generate Go code from templates
cd cmd/ppa-web
templ generate
```

#### Step 2: Run the Web Server

```bash
# Start with device discovery enabled
go run cmd/ppa-web/main.go --discover --log-level debug

# Or specify interfaces for discovery
go run cmd/ppa-web/main.go --discover --interfaces eth0,wlan0
```

#### Step 3: Access the Interface

Open your browser to `http://localhost:8080`. You'll see:

1. **Device Connection**: Enter IP addresses manually or use discovery
2. **Volume Control**: Real-time volume slider
3. **Preset Buttons**: Quick access to presets 1-16
4. **Discovery Panel**: Shows automatically discovered devices
5. **Log Window**: Real-time feedback from device communication

### Extending the Web Interface

#### Adding New Controls

To add a new control (e.g., mute button):

1. **Add to template** (`templates/index.templ`):
```go
templ MuteControl(state types.AppState) {
    <button class="btn btn-warning"
            hx-post="/mute"
            hx-target="#log-window"
            hx-swap="innerHTML">
        if state.IsMuted {
            Unmute
        } else {
            Mute
        }
    </button>
}
```

2. **Add route** (`router/router.go`):
```go
r.HandleFunc("/mute", h.HandleMute).Methods(http.MethodPost)
```

3. **Add handler** (`handler/handler.go`):
```go
func (h *Handler) HandleMute(w http.ResponseWriter, r *http.Request) {
    // Toggle mute state
    h.srv.SetState(func(state *types.AppState) {
        state.IsMuted = !state.IsMuted
    })
    
    // Send mute command to device
    // h.srv.GetMultiClient().SendMute(state.IsMuted)
    
    // Return updated UI
    templates.LogWindow(h.srv.GetState()).Render(r.Context(), w)
}
```

#### Adding Real-time Status Updates

For real-time device status (volume feedback, preset changes):

1. **Extend AppState** (`types/types.go`):
```go
type AppState struct {
    // ... existing fields
    CurrentVolume float32
    CurrentPreset int
    DeviceStatus  string
}
```

2. **Process device responses** (`server/server.go`):
```go
func (s *Server) processDeviceResponse(msg client.ReceivedMessage) {
    if msg.Header.MessageType == protocol.MessageTypeDeviceData {
        // Parse device data response
        // Update state with current volume, preset, etc.
        s.SetState(func(state *types.AppState) {
            state.CurrentVolume = parsedVolume
            state.CurrentPreset = parsedPreset
        })
    }
}
```

3. **Update templates** to show real-time values:
```go
templ StatusDisplay(state types.AppState) {
    <div class="alert alert-info">
        <strong>Volume:</strong> { fmt.Sprintf("%.1f", state.CurrentVolume) }
        <strong>Preset:</strong> { fmt.Sprintf("%d", state.CurrentPreset) }
    </div>
}
```

---

## Testing and Simulation

Testing audio equipment can be challenging - you don't always have physical devices available. The PPA control system includes a sophisticated simulation framework that lets you test your applications without real hardware.

### Device Simulation

#### Basic Simulated Device

The simulation system creates virtual PPA devices that respond to the same protocol as real hardware:

```go
// lib/simulation/client.go
type SimulatedDevice struct {
    SendChannel           chan Response
    ReceiveChannel        chan *bytes.Buffer
    Settings              SimulatedDeviceSettings
    currentlyActivePreset int
    currentVolume         float32
}

type SimulatedDeviceSettings struct {
    UniqueId    [4]byte  // Device identifier
    ComponentId byte     // Component ID
    Name        string   // Device name
    Address     string   // IP address to bind to
    Port        uint16   // Port to listen on
    Interface   string   // Network interface
}
```

#### Running a Simulated Device

```bash
# Start a simulated device on localhost:5001
go run cmd/ppa-cli/main.go simulate --address 0.0.0.0 --log-level info

# Start on specific interface
go run cmd/ppa-cli/main.go simulate --address 192.168.1.200 --interface eth0
```

The simulated device will:
- Respond to ping messages
- Accept volume changes
- Handle preset recalls
- Send device information responses

#### Simulation Message Handling

Here's how the simulator handles different message types:

```go
func (sd *SimulatedDevice) handlePing(req *Request) error {
    // Parse the incoming ping
    header, err := protocol.ParseHeader(req.Buffer.Bytes())
    if err != nil {
        return err
    }

    // Create response header
    responseHeader := protocol.NewBasicHeader(
        protocol.MessageTypePing,
        protocol.StatusResponseServer,  // We're responding as server
        sd.Settings.UniqueId,
        header.SequenceNumber,          // Echo sequence number
        sd.Settings.ComponentId,
    )

    // Send response
    buf := new(bytes.Buffer)
    protocol.EncodeHeader(buf, responseHeader)
    
    sd.SendChannel <- Response{
        Buffer: buf,
        Addr:   req.Addr,
    }
    
    return nil
}

func (sd *SimulatedDevice) handlePresetRecall(req *Request) error {
    // Parse preset recall message
    header, _ := protocol.ParseHeader(req.Buffer.Bytes())
    presetData, _ := protocol.ParsePresetRecall(req.Buffer.Bytes()[12:])
    
    // Update simulated state
    sd.currentlyActivePreset = int(presetData.IndexPosition)
    
    log.Info().
        Int("preset", sd.currentlyActivePreset).
        Msg("Simulated device: preset recalled")
    
    // Send acknowledgment
    responseHeader := protocol.NewBasicHeader(
        protocol.MessageTypePresetRecall,
        protocol.StatusResponseServer,
        sd.Settings.UniqueId,
        header.SequenceNumber,
        sd.Settings.ComponentId,
    )
    
    buf := new(bytes.Buffer)
    protocol.EncodeHeader(buf, responseHeader)
    
    sd.SendChannel <- Response{Buffer: buf, Addr: req.Addr}
    return nil
}
```

### Testing Strategies

#### Unit Testing Protocol Functions

Test the protocol encoding/decoding functions:

```go
// protocol_test.go
func TestBasicHeaderEncoding(t *testing.T) {
    header := protocol.NewBasicHeader(
        protocol.MessageTypePing,
        protocol.StatusRequestClient,
        [4]byte{1, 2, 3, 4},
        12345,
        0xFF,
    )
    
    buf := new(bytes.Buffer)
    err := protocol.EncodeHeader(buf, header)
    assert.NoError(t, err)
    
    // Decode and verify
    decoded, err := protocol.ParseHeader(buf.Bytes())
    assert.NoError(t, err)
    assert.Equal(t, header.MessageType, decoded.MessageType)
    assert.Equal(t, header.SequenceNumber, decoded.SequenceNumber)
}

func TestVolumeEncoding(t *testing.T) {
    // Test volume encoding formula
    volume := float32(0.5)  // 50% volume
    
    twentyDB := 0x3e8
    minusEightyDB := 0x00
    expectedGain := uint32(volume * float32(twentyDB-minusEightyDB))
    
    // This should equal 500 (halfway between 0 and 1000)
    assert.Equal(t, uint32(500), expectedGain)
}
```

#### Integration Testing with Simulation

Test your application against simulated devices:

```go
func TestVolumeControlIntegration(t *testing.T) {
    // Start simulated device
    settings := simulation.SimulatedDeviceSettings{
        UniqueId:    [4]byte{1, 2, 3, 4},
        ComponentId: 0xFF,
        Name:        "Test Speaker",
        Address:     "127.0.0.1",
        Port:        5001,
    }
    
    device := simulation.NewSimulatedDevice(settings)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    go device.Run(ctx)
    
    // Give device time to start
    time.Sleep(100 * time.Millisecond)
    
    // Create client and send volume command
    client := client.NewSingleDevice("127.0.0.1:5001", "", 0xFF)
    receiveCh := make(chan client.ReceivedMessage)
    
    go client.Run(ctx, receiveCh)
    
    // Send volume command
    client.SendMasterVolume(0.7)
    
    // Wait for response
    select {
    case msg := <-receiveCh:
        assert.NotNil(t, msg.Header)
        // Verify response
    case <-time.After(1 * time.Second):
        t.Fatal("No response received")
    }
}
```

#### Network Testing

Test discovery and multi-device scenarios:

```bash
# Terminal 1: Start multiple simulated devices
go run cmd/ppa-cli/main.go simulate --address 127.0.0.1 --port 5001 &
go run cmd/ppa-cli/main.go simulate --address 127.0.0.1 --port 5002 &
go run cmd/ppa-cli/main.go simulate --address 127.0.0.1 --port 5003 &

# Terminal 2: Test discovery
go run cmd/ppa-cli/main.go ping --discover --log-level info

# Terminal 3: Test volume control
go run cmd/ppa-cli/main.go volume --discover --volume 0.8
```

### Debugging Tools

#### Packet Capture

The system includes a packet capture tool for debugging:

```bash
# Capture all PPA traffic on interface eth0
go run cmd/pcap/main.go --interface eth0 --print-packets all

# Capture only volume commands
go run cmd/pcap/main.go --print-packets liveCmd --print-hexdump

# Analyze a saved capture file
go run cmd/pcap/main.go capture.pcap --print-hexdump
```

#### Logging Configuration

Use structured logging to debug issues:

```go
// Set log level for detailed debugging
zerolog.SetGlobalLevel(zerolog.DebugLevel)

// Log with context
log.Debug().
    Str("device", "192.168.1.100").
    Int("sequence", 123).
    Msg("Sending volume command")
```

---

## Advanced Topics

### Custom Message Types

You can extend the protocol to support custom functionality:

```go
// Add new message type
const MessageTypeCustom MessageType = 10

// Define custom payload structure
type CustomMessage struct {
    CustomField1 uint32
    CustomField2 [16]byte
    CustomString string
}

// Implement encoding/decoding
func EncodeCustomMessage(w io.Writer, cm *CustomMessage) error {
    // Implementation here
}

func ParseCustomMessage(buf []byte) (*CustomMessage, error) {
    // Implementation here
}
```

### Performance Optimization

#### Connection Pooling

For high-performance applications, implement connection pooling:

```go
type ConnectionPool struct {
    connections map[string]*SingleDevice
    mu          sync.RWMutex
    maxIdle     int
}

func (cp *ConnectionPool) GetConnection(addr string) *SingleDevice {
    cp.mu.RLock()
    if conn, exists := cp.connections[addr]; exists {
        cp.mu.RUnlock()
        return conn
    }
    cp.mu.RUnlock()
    
    // Create new connection
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    conn := NewSingleDevice(addr, "", 0xFF)
    cp.connections[addr] = conn
    return conn
}
```

#### Batch Operations

Send multiple commands efficiently:

```go
type BatchCommands struct {
    commands []Command
}

func (bc *BatchCommands) AddVolumeCommand(volume float32) {
    bc.commands = append(bc.commands, VolumeCommand{Volume: volume})
}

func (bc *BatchCommands) AddPresetCommand(preset int) {
    bc.commands = append(bc.commands, PresetCommand{Preset: preset})
}

func (bc *BatchCommands) Execute(client *MultiClient) error {
    for _, cmd := range bc.commands {
        if err := cmd.Execute(client); err != nil {
            return err
        }
    }
    return nil
}
```

### Security Considerations

#### Network Security

- Use VLANs to isolate audio equipment
- Implement firewall rules for UDP port 5001
- Consider VPN for remote access

#### Authentication

While the base protocol doesn't include authentication, you can add it at the application layer:

```go
type AuthenticatedClient struct {
    *SingleDevice
    token string
}

func (ac *AuthenticatedClient) SendAuthenticatedCommand(cmd Command) error {
    // Add authentication header
    authHeader := CreateAuthHeader(ac.token)
    
    // Send command with authentication
    return ac.sendWithAuth(cmd, authHeader)
}
```

### Monitoring and Metrics

#### Health Monitoring

Implement health checks for devices:

```go
type DeviceHealth struct {
    LastSeen    time.Time
    ResponseTime time.Duration
    ErrorCount   int
}

func (dh *DeviceHealth) IsHealthy() bool {
    return time.Since(dh.LastSeen) < 30*time.Second && 
           dh.ErrorCount < 5
}
```

#### Metrics Collection

Collect performance metrics:

```go
type Metrics struct {
    CommandsSent     int64
    ResponsesReceived int64
    AverageLatency   time.Duration
    ErrorRate        float64
}

func (m *Metrics) RecordCommand() {
    atomic.AddInt64(&m.CommandsSent, 1)
}

func (m *Metrics) RecordResponse(latency time.Duration) {
    atomic.AddInt64(&m.ResponsesReceived, 1)
    // Update average latency
}
```

---

## Troubleshooting

### Common Issues

#### 1. Device Discovery Not Working

**Symptoms:**
- No devices found during discovery
- Discovery times out

**Debugging Steps:**
```bash
# Check network connectivity
ping 192.168.1.100

# Verify UDP port is open
nmap -sU -p 5001 192.168.1.100

# Check firewall settings
sudo ufw status

# Test with specific interface
go run cmd/ppa-cli/main.go ping --discover --interfaces eth0 --log-level debug
```

**Common Causes:**
- Firewall blocking UDP port 5001
- Wrong network interface selected
- Device not responding to broadcast
- Network segmentation (VLANs)

#### 2. Commands Not Reaching Device

**Symptoms:**
- Commands sent but no response
- Device doesn't change state

**Debugging:**
```bash
# Enable packet capture
sudo tcpdump -i any -n port 5001

# Use simulation to test
go run cmd/ppa-cli/main.go simulate --address 0.0.0.0 &
go run cmd/ppa-cli/main.go volume --addresses 127.0.0.1 --volume 0.5
```

**Check:**
- Sequence numbers are incrementing
- Device ID matches
- Message encoding is correct

#### 3. Web Interface Not Updating

**Symptoms:**
- UI shows stale data
- Real-time updates not working

**Debugging:**
```javascript
// Check SSE connection in browser console
const eventSource = new EventSource('/discovery/events');
eventSource.onmessage = function(event) {
    console.log('SSE message:', event.data);
};
eventSource.onerror = function(event) {
    console.error('SSE error:', event);
};
```

**Common Causes:**
- SSE connection dropped
- Server not sending updates
- Browser blocking SSE

#### 4. Performance Issues

**Symptoms:**
- Slow response times
- High CPU usage
- Memory leaks

**Profiling:**
```bash
# CPU profiling
go run -cpuprofile=cpu.prof cmd/ppa-cli/main.go volume --discover

# Memory profiling
go run -memprofile=mem.prof cmd/ppa-cli/main.go volume --discover

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

### Debugging Techniques

#### 1. Enable Verbose Logging

```bash
# Maximum verbosity
go run cmd/ppa-cli/main.go ping --log-level debug

# Web server debugging
go run cmd/ppa-web/main.go --log-level debug
```

#### 2. Use Packet Capture

```bash
# Capture all PPA traffic
go run cmd/pcap/main.go --interface any --print-packets all --print-hexdump

# Save to file for analysis
go run cmd/pcap/main.go --interface any --timeout 60 > capture.log
```

#### 3. Simulate Network Issues

```bash
# Add network delay (Linux)
sudo tc qdisc add dev eth0 root netem delay 100ms

# Add packet loss
sudo tc qdisc add dev eth0 root netem loss 5%

# Remove simulation
sudo tc qdisc del dev eth0 root
```

### Best Practices

#### 1. Error Handling

Always handle errors gracefully:

```go
func sendCommandWithRetry(client *SingleDevice, cmd Command, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        if err := client.SendCommand(cmd); err != nil {
            log.Warn().Err(err).Int("attempt", i+1).Msg("Command failed, retrying")
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        }
        return nil
    }
    return fmt.Errorf("command failed after %d retries", maxRetries)
}
```

#### 2. Resource Management

Use contexts for cancellation and timeouts:

```go
func sendWithTimeout(ctx context.Context, client *SingleDevice, cmd Command) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    done := make(chan error, 1)
    go func() {
        done <- client.SendCommand(cmd)
    }()
    
    select {
    case err := <-done:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

#### 3. Testing

Write comprehensive tests:

```go
func TestDeviceDiscovery(t *testing.T) {
    // Test with simulated devices
    // Test network failures
    // Test timeout scenarios
    // Test concurrent access
}
```

---

## Conclusion

The PPA Control System provides a robust foundation for building professional audio control applications. Its modular architecture, comprehensive protocol implementation, and extensive tooling make it suitable for everything from simple command-line tools to complex web interfaces.

Key takeaways for developers:

1. **Start Simple**: Begin with the CLI examples to understand the basic patterns
2. **Use the Simulation**: Test your applications without physical hardware
3. **Follow the Patterns**: The CommandContext and MultiClient patterns handle most complexity
4. **Handle Errors**: Network communication is unreliable - plan for failures
5. **Monitor Performance**: Use logging and metrics to understand system behavior

Whether you're building a mobile app, web interface, or custom automation system, the patterns and examples in this guide will help you create reliable, professional-grade audio control applications.

For more examples and advanced usage, explore the `cmd/` directory and examine how the existing applications implement various features. The codebase is designed to be educational - each component demonstrates best practices for Go development and network programming.

Happy coding, and enjoy building amazing audio control systems! 