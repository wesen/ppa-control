# iPhone Application Design for DSP Speaker System Control

## 1. Purpose

Design a mobile (iPhone) application to control and monitor DSP speaker systems, focusing on volume and preset management, device discovery, and real-time feedback. The app will communicate directly with PPA DSP devices over UDP, providing a user-friendly interface for live control in professional or home audio environments.

## 2. System Architecture Overview

Based on the codebase analysis, the PPA control system uses:
- **UDP-based binary protocol** on port 5001 (default)
- **Device discovery** via UDP broadcast messages
- **Real-time control** through LiveCmd messages
- **Preset management** via PresetRecall messages
- **Device information** through DeviceData messages

## 3. User Stories
- As a user, I want to discover available speaker devices on my network so I can connect and control them.
- As a user, I want to view the status of each speaker (online/offline, current preset, volume level).
- As a user, I want to adjust the volume of individual or grouped speakers.
- As a user, I want to recall and apply presets to speakers.
- As a user, I want to see real-time updates when device states change.
- As a user, I want to manually enter device IP addresses when automatic discovery isn't available.
- As a user, I want a simple, reliable, and visually appealing mobile experience.

## 4. Key Features & UI Screens

### 4.1. Device Discovery Screen
- **Auto-discovery section**: List of discovered speakers via UDP broadcast
  - Device name, IP address, status indicator
  - Pull-to-refresh functionality
  - Real-time updates as devices come online/offline
- **Manual entry section**: 
  - Text input for IP address entry
  - "Add Device" button
  - Validation for IP format
- **Device list**: Combined view of discovered and manually added devices
  - Connect/disconnect buttons
  - Device status indicators (online/offline/connecting)

### 4.2. Device Control Screen
- **Device header**: Name, IP, connection status
- **Volume control**: 
  - Large slider (0-100%)
  - Mute/unmute toggle button
  - Real-time volume feedback
- **Preset management**:
  - Grid or list of available presets (0-255)
  - Current preset indicator
  - Recall preset buttons
- **Advanced controls** (expandable section):
  - Individual input/output controls
  - EQ settings
  - Gain adjustments
- **Status feedback**: Success/error messages for commands

### 4.3. Multi-Device Control (Group Control)
- **Device selection**: Checkboxes to select multiple devices
- **Synchronized volume control**: Master volume slider affecting all selected devices
- **Preset synchronization**: Apply same preset to all selected devices
- **Status overview**: Grid showing status of all selected devices

### 4.4. Settings Screen
- **Network settings**: Default port, timeout values
- **Discovery settings**: Interface selection, broadcast intervals
- **App preferences**: Theme, notifications, logging level
- **About**: App version, protocol information

## 5. Technical Architecture

### 5.1. Communication Protocol Implementation

#### 5.1.1. UDP Socket Management
- **Library**: Use `react-native-udp` for UDP socket functionality
- **Socket lifecycle**: Create, bind, send, receive, close
- **Error handling**: Connection timeouts, network errors, malformed packets

#### 5.1.2. PPA Protocol Implementation
```javascript
// Message structure (12-byte header + payload)
const BasicHeader = {
  messageType: 1,      // MessageType (Ping=0, LiveCmd=1, DeviceData=2, PresetRecall=4)
  protocolId: 1,       // Always 1
  status: 0x0102,      // StatusType (Client/Server, Command/Request/Response)
  deviceId: [0,0,0,0], // 4-byte device unique ID
  sequenceNumber: 1,   // 2-byte sequence
  componentId: 0xFF,   // Component ID
  reserved: 0          // Reserved byte
};
```

#### 5.1.3. Message Types Implementation
- **Ping Messages**: Device discovery and keepalive
- **LiveCmd Messages**: Volume control, gain adjustments
- **DeviceData Messages**: Device information requests/responses
- **PresetRecall Messages**: Preset loading by index

### 5.2. Device Discovery Implementation
```javascript
// Discovery process
1. Send UDP broadcast ping on port 5001
2. Listen for responses from devices
3. Parse device information from responses
4. Maintain device list with timeout handling
5. Update UI with discovered devices
```

### 5.3. Volume Control Implementation
```javascript
// Volume encoding: volume = 1 -> 0dB, volume = 0 -> -72dB
const twentyDB = 0x3e8;
const minusEightyDB = 0x00;
const encodedGain = volume * (twentyDB - minusEightyDB);

// LiveCmd path for master volume: [0, Input, 0, Gain]
const volumePath = [0, 1, 0, 4, 0, 0, 0, 0, 0, 0];
```

### 5.4. State Management
- **Redux/Context**: Centralized state for devices, connections, UI state
- **Real-time updates**: WebSocket-like pattern for UDP message handling
- **Persistence**: Save device list, preferences to AsyncStorage

### 5.5. Technology Stack
- **Framework**: React Native with Expo (if UDP support available) or bare React Native
- **UDP Library**: `react-native-udp` for socket communication
- **State Management**: Redux Toolkit or React Context
- **UI Components**: React Native Elements or NativeBase
- **Navigation**: React Navigation
- **Storage**: AsyncStorage for persistence

## 6. Network Communication Details

### 6.1. Device Discovery Flow
```
1. App starts → Create UDP socket on random port
2. Send broadcast ping to 255.255.255.255:5001
3. Listen for responses with DeviceData messages
4. Parse device information (name, IP, capabilities)
5. Add to device list with timestamp
6. Repeat every 5 seconds, timeout devices after 30 seconds
```

### 6.2. Device Control Flow
```
1. User selects device → Establish dedicated connection
2. Send initial ping to verify connectivity
3. Request device status via DeviceData request
4. User adjusts volume → Send LiveCmd with volume path
5. User recalls preset → Send PresetRecall message
6. Listen for responses and update UI accordingly
```

### 6.3. Message Encoding/Decoding
```javascript
// Encode basic header
function encodeHeader(messageType, status, deviceId, seqNum, componentId) {
  const buffer = new ArrayBuffer(12);
  const view = new DataView(buffer);
  view.setUint8(0, messageType);
  view.setUint8(1, 1); // Protocol ID
  view.setUint16(2, status, true); // Little endian
  // ... continue encoding
  return new Uint8Array(buffer);
}

// Decode received messages
function decodeMessage(data) {
  const view = new DataView(data.buffer);
  return {
    messageType: view.getUint8(0),
    protocolId: view.getUint8(1),
    status: view.getUint16(2, true),
    // ... continue decoding
  };
}
```

## 7. UI/UX Design Considerations

### 7.1. Design Principles
- **Professional audio focus**: Clean, functional interface suitable for live environments
- **Real-time feedback**: Immediate visual confirmation of actions
- **Error resilience**: Clear error messages and recovery options
- **Accessibility**: Support for VoiceOver, large text, high contrast

### 7.2. Visual Design
- **Color scheme**: Dark theme for low-light environments, light theme option
- **Typography**: Clear, readable fonts with good contrast
- **Icons**: Intuitive audio-related iconography
- **Feedback**: Loading states, success/error animations

### 7.3. Interaction Patterns
- **Swipe gestures**: Swipe to connect/disconnect devices
- **Long press**: Access advanced options
- **Pull to refresh**: Update device discovery
- **Haptic feedback**: Confirm important actions

## 8. Implementation Phases

### Phase 1: Core Protocol Implementation
- [ ] UDP socket management
- [ ] Basic PPA protocol encoding/decoding
- [ ] Device discovery functionality
- [ ] Simple ping/pong communication

### Phase 2: Basic Device Control
- [ ] Device list UI
- [ ] Manual IP entry
- [ ] Basic volume control
- [ ] Preset recall functionality

### Phase 3: Enhanced Features
- [ ] Multi-device control
- [ ] Advanced audio controls (EQ, gain)
- [ ] Settings and preferences
- [ ] Error handling and recovery

### Phase 4: Polish and Optimization
- [ ] UI/UX refinements
- [ ] Performance optimization
- [ ] Testing with real hardware
- [ ] App store preparation

## 9. Technical Challenges & Solutions

### 9.1. UDP in React Native
- **Challenge**: Limited UDP support in Expo managed workflow
- **Solution**: Use bare React Native with `react-native-udp` library
- **Alternative**: Expo development build with custom native modules

### 9.2. Binary Protocol Handling
- **Challenge**: JavaScript's limited binary data handling
- **Solution**: Use ArrayBuffer, DataView, and Uint8Array for binary operations
- **Encoding**: Implement proper little-endian encoding for protocol compliance

### 9.3. Real-time Communication
- **Challenge**: Maintaining responsive UI during network operations
- **Solution**: Use background threads for network operations, async/await patterns
- **Timeout handling**: Implement proper timeout and retry mechanisms

### 9.4. Device State Synchronization
- **Challenge**: Keeping UI in sync with actual device state
- **Solution**: Periodic status polling, optimistic UI updates with rollback

## 11. Deployment Considerations

### 11.1. Platform Requirements
- **iOS**: Minimum iOS 12.0 for React Native compatibility
- **Network permissions**: Required for UDP socket access
- **Background execution**: Limited background network access

### 11.2. App Store Guidelines
- **Network usage**: Clearly document network requirements
- **Privacy**: Explain local network access needs
- **Functionality**: Demonstrate core features work without server dependency

## 14. Relevant Files from Codebase

### Core Protocol Implementation
- `lib/protocol/ppa-protocol.go` - Protocol message structures and encoding
- `tutorial-protocol.md` - Protocol documentation and examples

### Client Implementation Reference
- `lib/client/single-device.go` - Single device UDP communication
- `lib/client/discovery/discovery.go` - Device discovery implementation
- `cmd/ppa-cli/cmds/` - CLI command implementations for reference

### Example Usage
- `cmd/ppa-cli/cmds/volume.go` - Volume control implementation
- `cmd/ppa-cli/cmds/recall.go` - Preset recall implementation
- `cmd/ppa-cli/cmds/ping.go` - Device discovery and ping

---

**Summary:**
This iPhone app will provide a modern, real-time interface for controlling DSP speaker systems using the existing PPA UDP protocol. The app will focus on device discovery, volume control, and preset management, with a clean and responsive UI optimized for professional audio environments. The implementation will use React Native with native UDP socket support to communicate directly with devices without requiring a backend server. 