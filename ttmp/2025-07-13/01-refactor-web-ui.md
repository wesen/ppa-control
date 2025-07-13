# PPA Control Web UI Refactoring Specification

**Date:** 2025-07-13  
**Version:** 1.0  
**Scope:** Complete web UI simplification for single client/device architecture

## Executive Summary

Complete refactoring of ppa-web to a simplified single-client, single-device WebSocket-based architecture. Eliminates multi-client complexity, removes backwards compatibility, and provides real-time bidirectional control.

## 1. Architecture Principles

### 1.1 Core Assumptions
- **Single Device**: One PPA device connection at all times
- **Single Client**: One WebSocket connection maximum (second connection = HTTP 409 Conflict)
- **No Backwards Compatibility**: Complete rewrite, no legacy support
- **Real-time First**: WebSocket as primary communication method
- **Stateless Web**: No server-side session management beyond device connection
- **Connection Lifecycle**: Device connection drops when WebSocket disconnects

### 1.2 Simplified Architecture
```
Browser (WebSocket) ↔ Go Web Server ↔ Single PPA Client ↔ Single Device
```

**Eliminated Complexity:**
- MultiClient management
- Discovery system (manual device connection)
- SSE streaming
- HTMX templates  
- State synchronization
- Connection pooling

## 2. Technical Specification

### 2.1 WebSocket Protocol

**Connection Model:**
- Single WebSocket connection per browser session
- Automatic reconnection on disconnect
- Direct 1:1 mapping to PPA protocol messages

**Message Format:**
```json
// Client → Server (Commands)
{
  "type": "connect",
  "address": "192.168.1.100:5001"
}
{
  "type": "volume", 
  "value": 0.75
}
{
  "type": "preset",
  "action": "recall|save",
  "index": 1
}
{
  "type": "parameter",
  "path": ["input", "1", "gain"],
  "value": -3.0
}

// Server → Client (Events)  
{
  "type": "connected",
  "device": {
    "id": "00:00:00:01",
    "name": "PPA Device",
    "address": "192.168.1.100:5001"
  }
}
{
  "type": "disconnected", 
  "reason": "timeout|websocket_closed|device_error"
}
{
  "type": "parameterChanged",
  "path": ["volume"],
  "value": 0.8
}
{
  "type": "error",
  "category": "transport|protocol|command|internal",
  "message": "Device not responding"
}
```

**Keep-alive**: Uses WebSocket ping/pong frames (10s timeout), no application-level ping

### 2.2 Go Backend Structure

**Simplified Package Structure:**
```
cmd/ppa-web/
├── main.go                 # Entry point with WebSocket server
├── websocket/
│   ├── handler.go         # WebSocket connection handler  
│   ├── messages.go        # Message types and parsing
│   └── bridge.go          # PPA protocol bridge
├── device/
│   ├── client.go          # Single device client wrapper
│   └── connection.go      # Connection management
└── static/
    ├── index.html         # Single-page application
    ├── app.js             # WebSocket client
    └── style.css          # Styling
```

**Core Components:**
- **WebSocketHandler**: Manages single WebSocket connection (rejects second connections)
- **DeviceClient**: Wraps PPA client for single device with context cancellation
- **MessageBridge**: Converts between WebSocket JSON and PPA binary
- **ConnectionManager**: Handles device connect/disconnect lifecycle

**Dependencies:**
- `github.com/gorilla/websocket` for WebSocket upgrade
- Existing `ppa-control/lib` for PPA protocol

### 2.3 Frontend Architecture

**Single-Page Application:**
- Vanilla JavaScript (no frameworks)
- WebSocket client with auto-reconnect
- Real-time UI updates
- Simple control interface

**UI Components:**
- Connection panel (address input)
- Device status indicator  
- Volume slider (real-time)
- Preset buttons (1-10)
- Parameter controls (gain, mute, etc.)
- Log/status area

## 3. Implementation Steps

### 3.1 Phase 1: Core WebSocket Infrastructure

**Step 1.1: WebSocket Handler**
```go
// websocket/handler.go
type Handler struct {
    conn       *websocket.Conn
    device     *device.Client
    sendCh     chan []byte
    ctx        context.Context
    cancel     context.CancelFunc
    connected  *atomic.Bool // Single connection guard
}

func (h *Handler) HandleConnection(w http.ResponseWriter, r *http.Request) {
    // Return HTTP 409 if already connected
}
func (h *Handler) readLoop(ctx context.Context)  // WebSocket → Device
func (h *Handler) writeLoop(ctx context.Context) // Device → WebSocket
func (h *Handler) handleVolumeDebounce(volume float32) // Rate-limit volume changes
```

**Step 1.2: Message Types**
```go
// websocket/messages.go
type ClientMessage struct {
    Type string          `json:"type"`
    Data json.RawMessage `json:"data,omitempty"`
}

type ServerMessage struct {
    Type string      `json:"type"`
    Data interface{} `json:"data,omitempty"`
}

type ConnectRequest struct {
    Address string `json:"address"`
}

type VolumeRequest struct {
    Value float32 `json:"value"`
}
```

**Step 1.3: Device Client Wrapper**
```go
// device/client.go
type Client struct {
    ppaClient lib.Client
    eventCh   chan ServerMessage
    connected bool
    mu        sync.RWMutex
}

func (c *Client) Connect(address string) error
func (c *Client) Disconnect() error  
func (c *Client) SendVolume(volume float32) error
func (c *Client) Events() <-chan ServerMessage
```

### 3.2 Phase 2: Protocol Bridge

**Step 2.1: Message Bridge Implementation**
```go
// websocket/bridge.go
type Bridge struct {
    device *device.Client
}

func (b *Bridge) HandleClientMessage(msg ClientMessage) error {
    switch msg.Type {
    case "connect":
        var req ConnectRequest
        json.Unmarshal(msg.Data, &req)
        return b.device.Connect(req.Address)
    case "volume":
        var req VolumeRequest  
        json.Unmarshal(msg.Data, &req)
        return b.device.SendVolume(req.Value)
    }
}
```

**Step 2.2: PPA Event Conversion**
```go
func (b *Bridge) convertPPAMessage(ppaMsg lib.ReceivedMessage) ServerMessage {
    switch ppaMsg.Header.MessageType {
    case lib.MessageTypeDeviceData:
        return ServerMessage{
            Type: "connected",
            Data: convertDeviceData(ppaMsg.DeviceData),
        }
    case lib.MessageTypeLiveCmd:
        return ServerMessage{
            Type: "parameterChanged", 
            Data: convertLiveCmd(ppaMsg.LiveCmd),
        }
    }
}
```

### 3.3 Phase 3: Frontend Implementation

**Step 3.1: WebSocket Client**
```javascript
// static/app.js
class PPAWebClient {
    constructor() {
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000; // Start at 1s, exponential backoff
    }
    
    connect() {
        const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
        this.ws = new WebSocket(`${protocol}//${location.host}/ws`);
        this.ws.onmessage = (event) => this.handleMessage(JSON.parse(event.data));
        this.ws.onclose = () => this.handleDisconnect();
        this.ws.onerror = () => this.handleError();
    }
    
    handleDisconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            setTimeout(() => this.connect(), this.reconnectDelay);
            this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000);
            this.reconnectAttempts++;
        }
    }
    
    send(type, data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ type, data }));
        }
    }
}
```

**Step 3.2: UI Controls**
```javascript
class UIController {
    constructor(client) {
        this.client = client;
        this.setupControls();
    }
    
    setupControls() {
        // Volume slider with debounced updates (20ms)
        const volumeSlider = document.getElementById('volume');
        let volumeTimeout;
        volumeSlider.addEventListener('input', (e) => {
            clearTimeout(volumeTimeout);
            volumeTimeout = setTimeout(() => {
                this.client.send('volume', { value: parseFloat(e.target.value) });
            }, 20);
        });
        
        // Preset buttons
        for (let i = 1; i <= 10; i++) {
            document.getElementById(`preset-${i}`).addEventListener('click', () => {
                this.client.send('preset', { action: 'recall', index: i });
            });
        }
    }
}
```

### 3.4 Phase 4: Simplified Main Application

**Step 4.1: Minimal main.go**
```go
// main.go
func main() {
    mux := http.NewServeMux()
    
    // Static files
    mux.Handle("/", http.FileServer(http.Dir("static/")))
    
    // WebSocket endpoint
    wsHandler := websocket.NewHandler()
    mux.HandleFunc("/ws", wsHandler.HandleConnection)
    
    log.Printf("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

**Step 4.2: Remove Existing Complexity**
- Delete `cmd/ppa-web/handler/` (replace with websocket/)
- Delete `cmd/ppa-web/router/` (use simple mux)
- Delete `cmd/ppa-web/server/` (merge into device/)
- Delete `cmd/ppa-web/templates/` (replace with static/)
- Remove SSE, HTMX, discovery system

## 4. File Structure Changes

### 4.1 Files to Delete
```
cmd/ppa-web/handler/
cmd/ppa-web/router/  
cmd/ppa-web/server/
cmd/ppa-web/templates/
cmd/ppa-web/static/ (existing)
```

### 4.2 Files to Create
```
cmd/ppa-web/websocket/
├── handler.go
├── messages.go
└── bridge.go

cmd/ppa-web/device/
├── client.go
└── connection.go

cmd/ppa-web/static/
├── index.html
├── app.js
└── style.css
```

### 4.3 Files to Modify
```
cmd/ppa-web/main.go (complete rewrite)
```

## 5. Benefits of Simplified Architecture

### 5.1 Reduced Complexity
- **90% fewer lines of code** in web layer
- **No state management** beyond single connection
- **No concurrent coordination** between clients
- **Direct message mapping** between WebSocket and PPA

### 5.2 Improved Performance  
- **Lower latency** (direct WebSocket → PPA bridge)
- **Reduced memory usage** (single client state)
- **Simplified debugging** (linear message flow)
- **Real-time responsiveness** (bidirectional WebSocket)

### 5.3 Enhanced Maintainability
- **Clear separation** of concerns
- **Testable components** (simple interfaces)
- **No legacy compatibility** burden
- **Modern web standards** (WebSocket, vanilla JS)

## 6. Migration Strategy

### 6.1 Development Approach
1. **Parallel Development**: Build new architecture alongside existing
2. **Feature Parity**: Implement core PPA operations first
3. **Testing**: Validate against single device
4. **Replacement**: Switch entry point to new main.go

### 6.2 Rollback Plan
- Keep existing code in `cmd/ppa-web-legacy/`
- Switch back via symbolic link if needed
- No data migration required (stateless)

## 7. Testing Strategy

### 7.1 Component Testing
- **WebSocket Handler**: Connection lifecycle, message parsing
- **Device Client**: PPA protocol integration
- **Message Bridge**: JSON ↔ Binary conversion

### 7.2 Integration Testing  
- **End-to-End**: Browser → WebSocket → Device → Response
- **Error Handling**: Device disconnection, invalid messages
- **Performance**: Message latency, reconnection speed

### 7.3 Manual Testing
- **Real Device**: Connect to actual PPA hardware
- **Volume Control**: Real-time slider responsiveness  
- **Preset Operations**: Save/recall functionality
- **Connection Recovery**: Network interruption handling

## 8. Success Criteria

### 8.1 Functional Requirements
- ✅ Connect to single PPA device via IP address
- ✅ Real-time volume control with immediate feedback
- ✅ Preset save/recall operations
- ✅ Device status monitoring
- ✅ Automatic reconnection on disconnect

### 8.2 Performance Requirements
- ✅ < 50ms WebSocket message round-trip
- ✅ < 100ms volume change response time
- ✅ Graceful handling of device disconnection
- ✅ Auto-reconnect within 5 seconds

### 8.3 Code Quality Requirements
- ✅ < 1000 lines of Go code (excluding static files and generated code)
- ✅ 90%+ test coverage for core components with mocked PPA client
- ✅ Minimal external dependencies (gorilla/websocket only)
- ✅ Clear error messages for all failure modes
- ✅ Second WebSocket connection rejected with HTTP 409
- ✅ UI survives browser refresh without crashing server
- ✅ Optional TLS support via command-line flag

### 8.4 Additional Considerations
- **Device Discovery**: Users must obtain device IP via DHCP logs, device labels, or network scanning
- **Static Assets**: Embedded with `embed.FS` for single-binary deployment
- **Build Requirements**: Go 1.21+, gorilla/websocket dependency
- **Security**: Optional TLS and basic auth flags for network deployment

This specification provides a complete roadmap for transforming the ppa-web application from a complex multi-client system to a streamlined, real-time, single-device control interface.
