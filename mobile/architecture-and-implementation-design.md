# PPA Control Mobile Application - Architecture, Design & Implementation

## Table of Contents

1. [Overview & Design Philosophy](#overview--design-philosophy)
2. [Technical Architecture](#technical-architecture)
3. [Project Structure](#project-structure)
4. [Core Components](#core-components)
5. [Protocol Implementation](#protocol-implementation)
6. [State Management](#state-management)
7. [Network Communication](#network-communication)
8. [User Interface Architecture](#user-interface-architecture)
9. [Debugging & Logging System](#debugging--logging-system)
10. [Development Workflow](#development-workflow)
11. [Design Patterns](#design-patterns)
12. [Error Handling Strategy](#error-handling-strategy)
13. [Performance Considerations](#performance-considerations)
14. [Testing Architecture](#testing-architecture)
15. [Deployment & Build Process](#deployment--build-process)

---

## Overview & Design Philosophy

### Purpose

The PPA Control Mobile Application is a React Native application built with Expo that provides professional audio engineers and users with a mobile interface to control PPA (Professional PA) DSP speaker systems. The application communicates directly with hardware devices over UDP using a custom binary protocol, eliminating the need for intermediate servers or cloud services.

### Core Design Principles

**Real-time Performance**: The application prioritizes low-latency communication suitable for live audio environments. All network operations are designed to provide immediate feedback with optimistic UI updates and proper rollback mechanisms.

**Professional Reliability**: Given its use in professional audio contexts, the application emphasizes robust error handling, graceful degradation, and comprehensive logging to ensure reliability during critical live events.

**Modular Architecture**: The codebase is structured with clear separation of concerns, making it easy to extend functionality, add new device types, or modify the protocol implementation without affecting other components.

**Native Platform Integration**: While built with React Native for cross-platform compatibility, the application leverages native networking capabilities through `react-native-udp` to achieve the precise UDP communication required by the PPA protocol.

### Target Use Cases

- **Live Sound Engineers**: Real-time volume and preset control during live performances
- **Installation Technicians**: Configuration and testing of distributed speaker systems
- **Home Audio Enthusiasts**: Control of residential DSP speaker installations
- **System Integrators**: Multi-device synchronization and management

---

## Technical Architecture

### High-Level Architecture

The application follows a layered architecture pattern with clear boundaries between presentation, business logic, and data layers:

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                       │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │ Discovery   │ │ Control     │ │ Multi-Device│ │Settings│ │
│  │ Screen      │ │ Screen      │ │ Screen      │ │ Screen │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    Business Logic Layer                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │ Redux Store │ │ Custom Hooks│ │ Navigation  │ │ Logger │ │
│  │ & Slices    │ │ (useUDP)    │ │ System      │ │ System │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │ UDP Service │ │ Protocol    │ │ Device      │ │ Debug  │ │
│  │             │ │ Encoding    │ │ Management  │ │ Panel  │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                     Protocol Layer                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │ UDP Sockets │ │ Binary      │ │ Message     │ │ Device │ │
│  │ (Native)    │ │ Encoding    │ │ Types       │ │ Types  │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

**Framework & Runtime**:
- **React Native**: Cross-platform mobile development framework
- **Expo**: Development platform providing build tools and services
- **TypeScript**: Type-safe JavaScript for enhanced developer experience

**State Management**:
- **Redux Toolkit**: Modern Redux implementation with simplified syntax
- **React Redux**: React bindings for Redux state management

**Navigation**:
- **React Navigation**: Tab-based navigation with stack navigators
- **Bottom Tab Navigator**: Primary navigation pattern

**Networking**:
- **react-native-udp**: Native UDP socket implementation for real-time communication
- **ArrayBuffer/DataView**: Binary data manipulation for protocol compliance

**UI Components**:
- **React Native Elements**: Consistent UI component library
- **React Native Community Slider**: Volume control components
- **Expo Vector Icons**: Icon library for navigation and UI elements

**Development Tools**:
- **Expo Dev Client**: Custom development builds with native modules
- **EAS Build**: Cloud-based build service for development and production
- **Metro**: JavaScript bundler with custom resolver configuration

---

## Project Structure

### Directory Organization

The project follows a feature-based organization with clear separation of concerns:

```
mobile/
├── src/
│   ├── components/          # Reusable UI components
│   │   └── DebugPanel.tsx   # In-app debugging interface
│   ├── hooks/               # Custom React hooks
│   │   └── useUDPService.ts # UDP service management hook
│   ├── navigation/          # Navigation configuration
│   │   └── AppNavigator.tsx # Tab-based navigation setup
│   ├── protocol/            # PPA protocol implementation
│   │   ├── types.ts         # Protocol type definitions
│   │   └── encoding.ts      # Binary encoding/decoding
│   ├── screens/             # Main application screens
│   │   ├── DiscoveryScreen.tsx    # Device discovery and connection
│   │   ├── ControlScreen.tsx      # Single device control
│   │   ├── MultiDeviceScreen.tsx  # Multi-device management
│   │   └── SettingsScreen.tsx     # App configuration
│   ├── services/            # Business logic services
│   │   └── UDPService.ts    # Core UDP communication service
│   ├── store/               # Redux state management
│   │   ├── index.ts         # Store configuration
│   │   ├── deviceSlice.ts   # Device state management
│   │   ├── controlSlice.ts  # Audio control state
│   │   └── settingsSlice.ts # App settings state
│   └── utils/               # Utility functions
│       └── logger.ts        # Comprehensive logging system
├── App.tsx                  # Root application component
├── app.json                 # Expo configuration
├── metro.config.js          # Metro bundler configuration
├── package.json             # Dependencies and scripts
├── README.md                # Development guide
├── DEBUGGING.md             # Debugging strategies
└── tsconfig.json            # TypeScript configuration
```

### File Naming Conventions

**Components**: PascalCase with descriptive names (`DiscoveryScreen.tsx`, `DebugPanel.tsx`)
**Hooks**: camelCase with `use` prefix (`useUDPService.ts`)
**Services**: PascalCase with service suffix (`UDPService.ts`)
**Types**: PascalCase for interfaces and enums (`DeviceInfo`, `MessageType`)
**Utilities**: camelCase descriptive names (`logger.ts`)

---

## Core Components

### Screen Components

#### DiscoveryScreen (`src/screens/DiscoveryScreen.tsx`)

The `DiscoveryScreen` serves as the primary entry point for device management, combining automatic discovery with manual device entry capabilities.

**Key Features**:
- **Auto-Discovery Toggle**: Users can enable/disable UDP broadcast discovery
- **Manual IP Entry**: Text inputs for IP address and port with validation
- **Device List**: Real-time display of discovered and manually added devices
- **Device Status**: Visual indicators showing online/offline status with last-seen timestamps
- **Pull-to-Refresh**: Manual refresh of device discovery process

**State Integration**:
```typescript
const { devices, isDiscovering, discoveryError } = useSelector((state: RootState) => state.devices);
const { autoDiscovery } = useSelector((state: RootState) => state.settings);
```

**Core Functions**:
- `handleDeviceSelect()`: Dispatches device selection and navigates to control screen
- `handleAddDevice()`: Validates and adds manually entered devices
- `handleRefresh()`: Restarts discovery process with visual feedback
- `handlePingDevice()`: Sends test ping to verify device connectivity

#### ControlScreen (`src/screens/ControlScreen.tsx`)

The `ControlScreen` provides comprehensive control over a single selected device, featuring volume control, preset management, and real-time feedback.

**Volume Control Architecture**:
- **Local State Management**: Maintains `localVolume` for responsive UI updates
- **Optimistic Updates**: Immediate UI response with server confirmation
- **Slider Integration**: Custom slider component with percentage and dB display
- **Mute Functionality**: Toggle mute with visual state indication

**Preset Management**:
```typescript
const renderPresetButtons = () => {
  const presets = Array.from({ length: 16 }, (_, i) => i + 1);
  return presets.map(preset => (
    <TouchableOpacity
      key={preset}
      style={[styles.presetButton, currentPreset === preset && styles.presetButtonActive]}
      onPress={() => handlePresetRecall(preset)}
    >
      <Text>{preset}</Text>
    </TouchableOpacity>
  ));
};
```

**Real-time Feedback System**:
- **Command Status**: Visual indicators for pending, success, and error states
- **Feedback Messages**: Scrollable log of recent device interactions
- **Error Handling**: Clear error messages with dismiss functionality

#### MultiDeviceScreen (`src/screens/MultiDeviceScreen.tsx`)

The `MultiDeviceScreen` enables synchronized control across multiple devices, essential for distributed audio systems.

**Device Selection Model**:
- **Checkbox Selection**: Individual device selection with visual feedback
- **Select All/None**: Bulk selection operations for efficiency
- **Connected Device Filtering**: Only displays devices that are currently online

**Master Control Implementation**:
- **Synchronized Volume**: Single slider controlling multiple devices simultaneously
- **Preset Synchronization**: Apply same preset across all selected devices
- **Batch Command Execution**: Parallel command sending with Promise.all()

**Status Overview Grid**:
```typescript
const statusGrid = selectedDeviceList.map(device => (
  <View key={device.address} style={styles.statusItem}>
    <Text>{device.name}</Text>
    <StatusIndicator connected={device.isConnected} />
  </View>
));
```

#### SettingsScreen (`src/screens/SettingsScreen.tsx`)

The `SettingsScreen` provides comprehensive configuration options with immediate effect application and debug panel access.

**Configuration Categories**:
- **Network Settings**: UDP port, discovery intervals, broadcast configuration
- **UI Preferences**: Theme selection, haptic feedback, notification settings
- **App Behavior**: Auto-discovery, screen management, command confirmations
- **Debug Options**: Log levels, packet logging, debug panel access

**Dynamic Setting Updates**:
```typescript
const handleLogLevelChange = (value: string) => {
  dispatch(setLogLevel(value as any));
  // Immediate logger configuration update
  const logLevelMap = { 'error': LogLevel.ERROR, 'warn': LogLevel.WARN, ... };
  logger.setLevel(logLevelMap[value]);
};
```

### Service Components

#### DebugPanel (`src/components/DebugPanel.tsx`)

The `DebugPanel` provides a comprehensive in-app debugging interface accessible through the settings screen.

**Log Management**:
- **Real-time Updates**: Live log stream with configurable auto-scroll
- **Filtering System**: Filter by log level (Debug, Info, Warn, Error) and context
- **Export Functionality**: Share logs via system share dialog for remote debugging

**State Inspection**:
- **Redux State Display**: Current state of all Redux slices with formatted output
- **Device Status Summary**: Connected devices, discovery status, selection state
- **Control State Overview**: Current volume, mute status, active presets

**User Interface**:
```typescript
const renderLogEntry = (entry: LogEntry, index: number) => (
  <View key={index} style={styles.logEntry}>
    <View style={styles.logHeader}>
      <Text style={styles.logTimestamp}>{entry.timestamp.toLocaleTimeString()}</Text>
      <Text style={[styles.logLevel, { color: levelColors[entry.level] }]}>
        {levelIcons[entry.level]} {LogLevel[entry.level]}
      </Text>
      <Text style={styles.logContext}>{entry.context}</Text>
    </View>
    <Text style={styles.logMessage}>{entry.message}</Text>
  </View>
);
```

---

## Protocol Implementation

### Binary Protocol Structure

The PPA protocol implementation in `src/protocol/` provides a complete TypeScript implementation of the binary communication protocol used by PPA DSP devices.

#### Type Definitions (`src/protocol/types.ts`)

**Message Type Enumeration**:
```typescript
export enum MessageType {
  Ping = 0,           // Device discovery and keepalive
  LiveCmd = 1,        // Real-time control commands
  DeviceData = 2,     // Device information exchange
  PresetRecall = 4,   // Load saved presets
  PresetSave = 5,     // Save current settings
}
```

**Status Type System**:
The status type encoding indicates message direction and purpose:
```typescript
export enum StatusType {
  RequestClient = 0x0101,   // Client requesting from server
  RequestServer = 0x0102,   // Server requesting from client
  CommandClient = 0x0201,   // Client commanding server
  CommandServer = 0x0202,   // Server commanding client
  ResponseClient = 0x0301,  // Client responding
  ResponseServer = 0x0302,  // Server responding
}
```

**Header Structure**:
The `BasicHeader` interface defines the 12-byte header present in all messages:
```typescript
export interface BasicHeader {
  messageType: MessageType;      // 1 byte - message type identifier
  protocolId: number;           // 1 byte - always 1 for PPA protocol
  status: StatusType;           // 2 bytes - direction and purpose
  deviceUniqueId: Uint8Array;   // 4 bytes - target device identifier
  sequenceNumber: number;       // 2 bytes - request/response tracking
  componentId: number;          // 1 byte - device component identifier
  reserved: number;             // 1 byte - reserved for future use
}
```

#### Encoding Implementation (`src/protocol/encoding.ts`)

**Header Encoding**:
The `encodeHeader()` function converts the header structure to binary format:
```typescript
export function encodeHeader(header: BasicHeader): ArrayBuffer {
  const buffer = new ArrayBuffer(12);
  const view = new DataView(buffer);
  
  view.setUint8(0, header.messageType);
  view.setUint8(1, header.protocolId);
  view.setUint16(2, header.status, true); // Little endian
  
  // Device unique ID (4 bytes)
  for (let i = 0; i < 4; i++) {
    view.setUint8(4 + i, header.deviceUniqueId[i] || 0);
  }
  
  view.setUint16(8, header.sequenceNumber, true);
  view.setUint8(10, header.componentId);
  view.setUint8(11, header.reserved);
  
  return buffer;
}
```

**Volume Encoding Algorithm**:
The volume encoding implements the PPA specification where volume ranges from -80dB to +20dB:
```typescript
export function encodeVolume(volume: number): number {
  const twentyDB = 0x3e8;      // +20dB encoded value (1000)
  const minusEightyDB = 0x00;  // -80dB encoded value (0)
  
  // Clamp volume between 0 and 1
  const clampedVolume = Math.max(0, Math.min(1, volume));
  
  // Linear mapping: 0.0 -> 0x00 (-80dB), 1.0 -> 0x3e8 (+20dB)
  return Math.round(clampedVolume * (twentyDB - minusEightyDB));
}
```

**Message Construction**:
The `createVolumeMessage()` function demonstrates the complete message creation process:
```typescript
export function createVolumeMessage(volume: number, sequenceNumber: number, componentId: number = 0xFF): ArrayBuffer {
  const header = createBasicHeader(
    MessageType.DeviceData,     // Volume uses DeviceData type
    StatusType.CommandClient,   // Client commanding device
    new Uint8Array([0, 0, 0, 0]), // Broadcast device ID
    sequenceNumber,
    componentId
  );
  
  // Create volume payload with path: [0, Input, 0, Gain]
  const payload = new ArrayBuffer(8);
  const payloadView = new DataView(payload);
  
  payloadView.setInt8(0, 1);  // Path component 1
  payloadView.setInt8(1, 0);  // Path component 2
  payloadView.setInt8(2, 3);  // Path component 3
  payloadView.setInt8(3, 6);  // Path component 4
  
  const encodedGain = encodeVolume(volume);
  payloadView.setUint32(4, encodedGain, true); // Little endian
  
  return createMessage(header, payload);
}
```

### Protocol Message Flow

**Discovery Process**:
1. Client sends ping broadcast to `255.255.255.255:5001`
2. Devices respond with ping responses containing device information
3. Client maintains device list with timeout-based cleanup

**Volume Control Process**:
1. User adjusts volume slider triggering `handleVolumeChangeEnd()`
2. Volume value encoded using `encodeVolume()` function
3. LiveCmd message constructed and sent via UDP
4. Device responds with acknowledgment
5. UI updated with success/error status

**Preset Recall Process**:
1. User selects preset button triggering `handlePresetRecall()`
2. PresetRecall message created with preset index
3. Message sent to device with command status
4. Device loads preset and responds with confirmation

---

## State Management

### Redux Architecture

The application uses Redux Toolkit for centralized state management, organized into three main slices that handle different aspects of application state.

#### Device State (`src/store/deviceSlice.ts`)

The device slice manages all device-related state including discovery, selection, and device metadata.

**State Structure**:
```typescript
export interface DeviceState {
  devices: DeviceInfo[];              // Array of discovered/added devices
  selectedDevice: DeviceInfo | null;  // Currently selected device for control
  selectedDevices: string[];          // Array of device addresses for multi-control
  isDiscovering: boolean;             // Discovery process status
  discoveryError: string | null;      // Discovery error messages
  connectionStatus: 'disconnected' | 'connecting' | 'connected';
}
```

**Key Actions**:
- `addDevice()`: Adds or updates device in the devices array with duplicate detection
- `selectDevice()`: Sets the currently selected device and updates connection status
- `toggleDeviceSelection()`: Manages multi-device selection for synchronized control
- `startDiscovery()/stopDiscovery()`: Controls the discovery process state

**Device Management Logic**:
```typescript
addDevice: (state, action: PayloadAction<DeviceInfo>) => {
  const deviceIndex = state.devices.findIndex(
    d => `${d.address}:${d.port}` === `${action.payload.address}:${action.payload.port}`
  );
  
  if (deviceIndex >= 0) {
    // Update existing device
    state.devices[deviceIndex] = action.payload;
  } else {
    // Add new device
    state.devices.push(action.payload);
  }
}
```

#### Control State (`src/store/controlSlice.ts`)

The control slice manages audio control parameters and command execution status.

**State Structure**:
```typescript
export interface ControlState {
  volume: number;                     // Current volume level (0.0-1.0)
  isMuted: boolean;                   // Mute status
  currentPreset: number | null;       // Active preset number
  isProcessingCommand: boolean;       // Command execution status
  lastCommandStatus: 'success' | 'error' | 'pending' | null;
  errorMessage: string | null;        // Last error message
  feedbackMessages: string[];         // Command feedback history
}
```

**Volume Management**:
```typescript
setVolume: (state, action: PayloadAction<number>) => {
  state.volume = Math.max(0, Math.min(1, action.payload)); // Clamp to valid range
  state.isMuted = false; // Unmute when volume is changed
}
```

**Command Status Tracking**:
The control slice tracks command execution with proper state transitions:
```typescript
setCommandProcessing: (state, action: PayloadAction<boolean>) => {
  state.isProcessingCommand = action.payload;
  if (action.payload) {
    state.lastCommandStatus = 'pending';
    state.errorMessage = null;
  }
}
```

#### Settings State (`src/store/settingsSlice.ts`)

The settings slice manages application configuration with immediate effect application.

**Configuration Categories**:
```typescript
export interface SettingsState {
  // Network configuration
  defaultPort: number;            // UDP port for device communication
  discoveryInterval: number;      // Time between discovery broadcasts
  deviceTimeout: number;          // Device offline timeout
  broadcastAddress: string;       // Network broadcast address
  
  // UI preferences
  theme: 'light' | 'dark' | 'auto';
  enableHapticFeedback: boolean;
  enableNotifications: boolean;
  
  // Debug configuration
  logLevel: 'debug' | 'info' | 'warn' | 'error';
  enablePacketLogging: boolean;
  
  // App behavior
  autoDiscovery: boolean;
  keepScreenOn: boolean;
  confirmCommands: boolean;
}
```

### Store Configuration

The store configuration in `src/store/index.ts` includes custom middleware for handling non-serializable data:

```typescript
export const store = configureStore({
  reducer: {
    devices: deviceReducer,
    control: controlReducer,
    settings: settingsReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // Ignore device timestamps and binary data
        ignoredActions: ['devices/addDevice', 'devices/updateDevice'],
        ignoredActionsPaths: ['payload.uniqueId', 'payload.lastSeen'],
        ignoredPaths: ['devices.devices.uniqueId', 'devices.devices.lastSeen'],
      },
    }),
});
```

---

## Network Communication

### UDP Service Architecture (`src/services/UDPService.ts`)

The `UDPService` class provides a comprehensive abstraction over UDP socket communication, handling device discovery, message sending, and response processing.

#### Service Configuration

**Configuration Interface**:
```typescript
export interface UDPServiceConfig {
  discoveryPort: number;        // Port for device discovery (default: 5001)
  discoveryInterval: number;    // Discovery broadcast interval (default: 5000ms)
  deviceTimeout: number;        // Device timeout period (default: 30000ms)
  broadcastAddress: string;     // Broadcast address (default: 255.255.255.255)
}
```

**Callback System**:
```typescript
export interface UDPServiceCallbacks {
  onDeviceDiscovered?: (device: DeviceInfo) => void;
  onDeviceTimeout?: (device: DeviceInfo) => void;
  onMessageReceived?: (message: ReceivedMessage) => void;
  onError?: (error: Error) => void;
}
```

#### Socket Management

**Socket Initialization**:
The service handles socket creation with comprehensive error handling and fallback support:
```typescript
async start(): Promise<void> {
  if (!dgram) {
    const error = new Error('react-native-udp not available - requires development build');
    logger.error('UDPService', 'Cannot start UDP service without native module', null, error);
    throw error;
  }
  
  this.socket = dgram.createSocket({
    type: 'udp4',
    reusePort: true,
  });
  
  this.setupSocketListeners();
  
  // Bind to random port and enable broadcast
  await this.bindSocket();
  this.socket.setBroadcast(true);
}
```

**Event Handling**:
```typescript
private setupSocketListeners(): void {
  this.socket.on('message', (buffer: Buffer, rinfo: any) => {
    logger.debug('UDPService', 'Received UDP message', {
      from: `${rinfo.address}:${rinfo.port}`,
      size: buffer.length
    });
    this.handleReceivedMessage(buffer, rinfo);
  });
  
  this.socket.on('error', (err: any) => {
    logger.error('UDPService', 'UDP socket error', {
      error: err.message,
      code: err.code
    }, err);
    this.handleError(new Error(`Socket error: ${err.message}`));
  });
}
```

#### Device Discovery

**Discovery Process**:
The discovery system uses periodic UDP broadcasts to locate devices:
```typescript
private sendDiscoveryPing(): void {
  const message = createPingMessage(this.getNextSequenceNumber());
  const buffer = Buffer.from(message);
  
  this.socket.send(
    buffer, 0, buffer.length,
    this.config.discoveryPort,
    this.config.broadcastAddress,
    (err: any) => {
      if (err) {
        this.handleError(new Error(`Failed to send discovery ping: ${err.message}`));
      } else {
        logger.debug('UDPService', 'Discovery ping broadcast sent');
      }
    }
  );
}
```

**Device Timeout Management**:
```typescript
private checkDeviceTimeouts(): void {
  const now = new Date();
  const timeoutThreshold = this.config.deviceTimeout;
  
  for (const [address, device] of this.devices.entries()) {
    const timeSinceLastSeen = now.getTime() - device.lastSeen.getTime();
    
    if (timeSinceLastSeen > timeoutThreshold && device.isConnected) {
      device.isConnected = false;
      logger.warn('UDPService', `Device ${address} timed out`);
      
      if (this.callbacks.onDeviceTimeout) {
        this.callbacks.onDeviceTimeout(device);
      }
    }
  }
}
```

#### Command Transmission

**Volume Command Sending**:
```typescript
sendVolumeCommand(deviceAddress: string, volume: number): void {
  logger.info('UDPService', 'Sending volume command', {
    deviceAddress,
    volume,
    volumePercent: Math.round(volume * 100)
  });
  
  const [host, portStr] = deviceAddress.split(':');
  const port = parseInt(portStr, 10) || this.config.discoveryPort;
  
  const message = createVolumeMessage(volume, this.getNextSequenceNumber());
  const buffer = Buffer.from(message);
  
  this.socket.send(buffer, 0, buffer.length, port, host, (err: any) => {
    if (err) {
      logger.error('UDPService', 'Failed to send volume command', {
        deviceAddress, volume, error: err.message
      }, err);
    } else {
      logger.info('UDPService', 'Volume command sent successfully');
    }
  });
}
```

### Hook Integration (`src/hooks/useUDPService.ts`)

The `useUDPService` hook provides a React-friendly interface to the UDP service, integrating with Redux state management.

**Service Initialization**:
```typescript
const initializeService = useCallback(() => {
  if (udpServiceRef.current) {
    return udpServiceRef.current;
  }
  
  const callbacks: UDPServiceCallbacks = {
    onDeviceDiscovered: (device) => {
      dispatch(addDevice(device));
      dispatch(addFeedbackMessage(`Device discovered: ${device.address}`));
    },
    onDeviceTimeout: (device) => {
      const updatedDevice = { ...device, isConnected: false };
      dispatch(updateDevice(updatedDevice));
    },
    onError: (error) => {
      dispatch(setDiscoveryError(error.message));
      dispatch(setErrorMessage(error.message));
    },
  };
  
  const config = {
    discoveryPort: settings.defaultPort,
    discoveryInterval: settings.discoveryInterval,
    deviceTimeout: settings.deviceTimeout,
    broadcastAddress: settings.broadcastAddress,
  };
  
  udpServiceRef.current = new UDPService(config, callbacks);
  return udpServiceRef.current;
}, [dispatch, settings]);
```

**Command Execution with State Management**:
```typescript
const sendVolumeCommand = useCallback(async (deviceAddress: string, volume: number) => {
  const service = udpServiceRef.current;
  if (!service) {
    dispatch(setErrorMessage('UDP service not available'));
    return;
  }
  
  try {
    dispatch(setCommandProcessing(true));
    service.sendVolumeCommand(deviceAddress, volume);
    dispatch(setCommandStatus('success'));
    dispatch(addFeedbackMessage(`Volume set to ${Math.round(volume * 100)}%`));
  } catch (error) {
    dispatch(setCommandStatus('error'));
    dispatch(setErrorMessage(error instanceof Error ? error.message : 'Unknown error'));
  }
}, [dispatch]);
```

---

## User Interface Architecture

### Navigation System

The application uses React Navigation with a tab-based structure providing intuitive access to major functionality areas.

#### Navigation Configuration (`src/navigation/AppNavigator.tsx`)

**Tab Structure**:
```typescript
<Tab.Navigator
  screenOptions={({ route }) => ({
    tabBarIcon: ({ focused, color, size }) => {
      let iconName: keyof typeof Ionicons.glyphMap;
      
      switch (route.name) {
        case 'Discovery': iconName = focused ? 'search' : 'search-outline'; break;
        case 'Control': iconName = focused ? 'volume-high' : 'volume-high-outline'; break;
        case 'MultiDevice': iconName = focused ? 'grid' : 'grid-outline'; break;
        case 'Settings': iconName = focused ? 'settings' : 'settings-outline'; break;
      }
      
      return <Ionicons name={iconName} size={size} color={color} />;
    },
    tabBarActiveTintColor: '#007AFF',
    tabBarInactiveTintColor: 'gray',
  })}
>
```

**Stack Navigation Integration**:
Each tab contains a stack navigator allowing for future expansion with detailed screens:
```typescript
const ControlStack = () => (
  <Stack.Navigator>
    <Stack.Screen 
      name="ControlMain" 
      component={ControlScreen}
      options={{ title: 'Device Control' }}
    />
  </Stack.Navigator>
);
```

### Design System

#### Color Palette

**Primary Colors**:
- **Primary Blue**: `#007AFF` - Interactive elements, active states
- **Success Green**: `#4CAF50` - Success states, online indicators
- **Warning Orange**: `#FF9500` - Warning states, debug elements
- **Error Red**: `#FF3B30` - Error states, offline indicators

**Neutral Colors**:
- **Background**: `#f5f5f5` - Screen backgrounds
- **Card Background**: `white` - Component backgrounds
- **Text Primary**: `#333` - Primary text content
- **Text Secondary**: `#666` - Secondary text content
- **Text Tertiary**: `#999` - Disabled text, timestamps

#### Typography System

**Font Hierarchy**:
```typescript
const styles = StyleSheet.create({
  screenTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#333',
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
  },
  bodyText: {
    fontSize: 16,
    fontWeight: '400',
    color: '#333',
  },
  captionText: {
    fontSize: 14,
    fontWeight: '400',
    color: '#666',
  },
  labelText: {
    fontSize: 12,
    fontWeight: '500',
    color: '#999',
  },
});
```

#### Component Styling Patterns

**Card Component Pattern**:
Consistent styling for content sections across screens:
```typescript
const cardStyles = {
  section: {
    backgroundColor: 'white',
    marginHorizontal: 16,
    marginVertical: 8,
    padding: 16,
    borderRadius: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
};
```

**Interactive Element Pattern**:
Consistent touch target styling with proper accessibility:
```typescript
const buttonStyles = {
  primaryButton: {
    backgroundColor: '#007AFF',
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderRadius: 8,
    alignItems: 'center',
    minHeight: 44, // Minimum touch target size
  },
  secondaryButton: {
    backgroundColor: 'transparent',
    borderWidth: 1,
    borderColor: '#007AFF',
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderRadius: 8,
    alignItems: 'center',
  },
};
```

### Responsive Design

**Screen Size Adaptation**:
The interface adapts to different screen sizes using flexible layouts:
```typescript
const responsiveStyles = {
  container: {
    flex: 1,
    paddingHorizontal: Platform.select({
      ios: 16,
      android: 16,
      web: '10%', // Wider margins on web
    }),
  },
  presetGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
    justifyContent: 'space-between',
  },
  presetButton: {
    width: Dimensions.get('window').width > 400 ? 70 : 60, // Larger on tablets
    height: 60,
  },
};
```

### Accessibility Implementation

**Screen Reader Support**:
```typescript
<TouchableOpacity
  accessible={true}
  accessibilityLabel={`Volume ${Math.round(volume * 100)} percent`}
  accessibilityHint="Adjust device volume"
  accessibilityRole="adjustable"
  onPress={handleVolumePress}
>
```

**High Contrast Support**:
```typescript
const getContrastColor = (isDarkMode: boolean) => ({
  text: isDarkMode ? '#FFFFFF' : '#000000',
  background: isDarkMode ? '#000000' : '#FFFFFF',
  accent: '#007AFF', // Maintains contrast in both modes
});
```

---

## Debugging & Logging System

### Logger Architecture (`src/utils/logger.ts`)

The logging system provides comprehensive debugging capabilities with structured output and in-app visualization.

#### Core Logger Implementation

**Log Level System**:
```typescript
export enum LogLevel {
  DEBUG = 0,  // Detailed execution flow
  INFO = 1,   // Important events
  WARN = 2,   // Potential issues
  ERROR = 3,  // Critical errors
}
```

**Log Entry Structure**:
```typescript
export interface LogEntry {
  timestamp: Date;
  level: LogLevel;
  context: string;    // Component or module name
  message: string;    // Human-readable message
  data?: any;         // Structured data
  error?: Error;      // Error object with stack trace
}
```

**Context-Aware Logging**:
```typescript
class Logger {
  private log(level: LogLevel, context: string, message: string, data?: any, error?: Error): void {
    if (level < this.currentLevel) return;
    
    const entry: LogEntry = {
      timestamp: new Date(),
      level, context, message, data, error,
    };
    
    // Store for in-app viewing
    this.logs.push(entry);
    
    // Console output with formatting
    const timestamp = entry.timestamp.toISOString().substr(11, 12);
    const levelStr = LogLevel[level].padEnd(5);
    const contextStr = context.padEnd(15);
    
    console.log(`[${timestamp}] ${levelStr} ${contextStr} ${message}`);
  }
}
```

#### Specialized Logging Methods

**Network Activity Logging**:
```typescript
udpPacket(direction: 'sent' | 'received', address: string, type: string, data?: any): void {
  this.debug('UDP', `${direction.toUpperCase()} ${type} ${direction === 'sent' ? 'to' : 'from'} ${address}`, data);
}

networkRequest(url: string, method: string, data?: any): void {
  this.debug('Network', `${method} ${url}`, data);
}
```

**State Change Tracking**:
```typescript
stateChange(context: string, before: any, after: any): void {
  this.debug('State', `${context} state change`, { before, after });
}

userAction(action: string, data?: any): void {
  this.info('User', action, data);
}
```

#### Global Development Access

**Development Mode Integration**:
```typescript
if (__DEV__) {
  (global as any).logger = logger;
  (global as any).LogLevel = LogLevel;
  
  logger.info('Logger', 'Logger initialized in development mode');
  logger.info('Logger', 'Access logger globally with: global.logger');
}
```

### Debug Panel Implementation

The debug panel provides real-time log viewing and app state inspection within the application.

#### Log Filtering and Display

**Dynamic Filtering**:
```typescript
const filteredLogs = logs.filter(entry => {
  if (selectedLevel !== undefined && entry.level < selectedLevel) return false;
  if (selectedContext && entry.context !== selectedContext) return false;
  return true;
});
```

**Real-time Updates**:
```typescript
useEffect(() => {
  const updateLogs = () => {
    setLogs(logger.getLogs(selectedLevel, selectedContext || undefined));
  };
  
  updateLogs();
  logger.addListener(updateLogs);
  
  return () => logger.removeListener(updateLogs);
}, [selectedLevel, selectedContext]);
```

#### State Inspection

**Redux State Display**:
```typescript
const renderAppState = () => (
  <View style={styles.stateContainer}>
    <Text style={styles.stateTitle}>App State Summary</Text>
    
    <View style={styles.stateSection}>
      <Text style={styles.stateSectionTitle}>Devices ({devices.devices.length})</Text>
      <Text>Connected: {devices.devices.filter(d => d.isConnected).length}</Text>
      <Text>Selected: {devices.selectedDevice?.name || 'None'}</Text>
      <Text>Discovery: {devices.isDiscovering ? 'Active' : 'Inactive'}</Text>
    </View>
  </View>
);
```

#### Log Export Functionality

**Share Integration**:
```typescript
const handleExportLogs = async () => {
  try {
    const logText = logger.exportLogs();
    await Share.share({
      message: logText,
      title: 'PPA Control Debug Logs',
    });
  } catch (error) {
    Alert.alert('Error', 'Failed to export logs');
  }
};
```

### Integration with Components

**Service-Level Logging**:
```typescript
// In UDPService.ts
logger.info('UDPService', 'Starting UDP service...', {
  config: this.config,
  isRunning: this.isRunning
});

logger.udpPacket('sent', deviceAddress, 'VolumeCommand', {
  volume, bufferLength: buffer.length
});
```

**Hook-Level Logging**:
```typescript
// In useUDPService.ts
const startService = useCallback(async () => {
  logger.info('useUDPService', 'Starting UDP service...', { autoDiscovery });
  
  try {
    await service.start();
    logger.info('useUDPService', 'UDP service started successfully');
  } catch (error) {
    logger.error('useUDPService', 'Failed to start UDP service', { error }, error);
  }
}, []);
```

---

## Development Workflow

### Environment Setup

**Prerequisites**:
- Node.js 18+ with npm
- Expo CLI: `npm install -g @expo/cli`
- EAS CLI: `npm install -g eas-cli`
- iOS Simulator (macOS) or Android Emulator

**Development Build Creation**:
```bash
# Initial setup
cd mobile
npm install

# Configure EAS project
eas build:configure

# Create development build
eas build --profile development --platform ios
# or
eas build --profile development --platform android
```

### Development Scripts

**Package.json Scripts**:
```json
{
  "scripts": {
    "start": "expo start",
    "android": "expo start --android",
    "ios": "expo start --ios",
    "build:dev": "eas build --profile development",
    "build:preview": "eas build --profile preview",
    "build:production": "eas build --profile production",
    "type-check": "tsc --noEmit",
    "lint": "eslint . --ext .ts,.tsx"
  }
}
```

### Testing Strategy

**Manual Testing with Simulation**:
```bash
# Terminal 1: Start simulated device
cd ..  # Main project directory
go run cmd/ppa-cli/main.go simulate --address 0.0.0.0

# Terminal 2: Run mobile app
cd mobile
expo start --dev-client
```

**Type Checking**:
```bash
npm run type-check
```

**Testing Checklist**:
- [ ] Auto-discovery finds simulated device
- [ ] Manual IP entry validation works
- [ ] Volume control sends UDP commands
- [ ] Preset recall functions properly
- [ ] Multi-device selection works
- [ ] Settings persist correctly
- [ ] Debug panel displays logs
- [ ] Error handling shows appropriate messages

### Debugging Workflow

**Standard Debugging Process**:
1. **Enable Debug Logging**: Set log level to DEBUG in settings
2. **Use Debug Panel**: Monitor real-time logs during operation
3. **Check Network**: Verify UDP service status and device connectivity
4. **Export Logs**: Share logs for external analysis if needed

**Common Debug Commands**:
```javascript
// In development console
global.logger.setLevel(global.LogLevel.DEBUG);
global.logger.info('Test', 'Custom log message');
global.logger.getLogs().filter(l => l.context === 'UDPService');
```

---

## Design Patterns

### Architecture Patterns

#### Service Layer Pattern

The application uses a service layer to abstract business logic from UI components:

```typescript
// Service definition
class UDPService {
  // Business logic for UDP communication
  public sendVolumeCommand(address: string, volume: number): void;
  public startDiscovery(): void;
}

// Hook integration
export const useUDPService = () => {
  // React integration and state management
  const serviceRef = useRef<UDPService | null>(null);
  return { sendVolumeCommand, startDiscovery, ... };
};

// Component usage
const ControlScreen = () => {
  const { sendVolumeCommand } = useUDPService();
  // UI logic only
};
```

#### Observer Pattern

The logging system implements the observer pattern for real-time updates:

```typescript
class Logger {
  private listeners: ((entry: LogEntry) => void)[] = [];
  
  addListener(listener: (entry: LogEntry) => void): void {
    this.listeners.push(listener);
  }
  
  private notifyListeners(entry: LogEntry): void {
    this.listeners.forEach(listener => listener(entry));
  }
}
```

#### Factory Pattern

Protocol message creation uses factory functions:

```typescript
export function createVolumeMessage(volume: number, sequenceNumber: number): ArrayBuffer;
export function createPresetRecallMessage(presetIndex: number, sequenceNumber: number): ArrayBuffer;
export function createPingMessage(sequenceNumber: number): ArrayBuffer;
```

#### Strategy Pattern

Different logging strategies based on environment:

```typescript
if (__DEV__) {
  // Development logging strategy
  logger.setLevel(LogLevel.DEBUG);
  logger.enableConsoleOutput(true);
} else {
  // Production logging strategy
  logger.setLevel(LogLevel.ERROR);
  logger.enableConsoleOutput(false);
}
```

### React Patterns

#### Custom Hooks Pattern

Custom hooks encapsulate complex logic and provide clean APIs:

```typescript
export const useUDPService = () => {
  // Complex UDP service management
  const serviceRef = useRef<UDPService | null>(null);
  
  // Simple API for components
  return {
    startService,
    sendVolumeCommand,
    sendPresetCommand,
    isServiceRunning: !!serviceRef.current,
  };
};
```

#### Compound Component Pattern

The debug panel uses compound components for flexibility:

```typescript
<DebugPanel visible={visible} onClose={onClose}>
  <DebugPanel.Header />
  <DebugPanel.Controls />
  <DebugPanel.LogList />
  <DebugPanel.StateInspector />
</DebugPanel>
```

#### Render Props Pattern

Log entry rendering uses render props for customization:

```typescript
const renderLogEntry = (entry: LogEntry, index: number) => (
  <LogEntry
    key={index}
    entry={entry}
    renderHeader={(entry) => <LogHeader {...entry} />}
    renderContent={(entry) => <LogContent {...entry} />}
  />
);
```

---

## Error Handling Strategy

### Error Categories

#### Network Errors

**UDP Module Unavailable**:
```typescript
try {
  dgram = require('react-native-udp');
} catch (error) {
  logger.error('UDPService', 'react-native-udp not available - requires development build', null, error);
  throw new Error('UDP functionality requires development build');
}
```

**Socket Operation Errors**:
```typescript
this.socket.send(buffer, 0, buffer.length, port, host, (err: any) => {
  if (err) {
    logger.error('UDPService', 'Failed to send command', { deviceAddress, error: err.message }, err);
    this.handleError(new Error(`Failed to send command: ${err.message}`));
  }
});
```

#### Protocol Errors

**Message Parsing Errors**:
```typescript
try {
  header = parseHeader(arrayBuffer);
} catch (error) {
  logger.warn('UDPService', 'Failed to parse message header', { error: error.message }, error);
  return; // Skip malformed messages
}
```

**Encoding Errors**:
```typescript
export function encodeHeader(header: BasicHeader): ArrayBuffer {
  try {
    const buffer = new ArrayBuffer(12);
    // ... encoding logic
    return buffer;
  } catch (error) {
    logger.error('Protocol', 'Failed to encode header', { header }, error);
    throw new Error(`Header encoding failed: ${error.message}`);
  }
}
```

#### Application Errors

**State Validation Errors**:
```typescript
const handleVolumeChange = (volume: number) => {
  if (volume < 0 || volume > 1 || isNaN(volume)) {
    logger.error('ControlScreen', 'Invalid volume value', { volume });
    dispatch(setErrorMessage('Invalid volume value'));
    return;
  }
  // ... valid processing
};
```

**Navigation Errors**:
```typescript
useEffect(() => {
  if (!selectedDevice) {
    logger.warn('ControlScreen', 'No device selected, navigating back');
    navigation.goBack();
  }
}, [selectedDevice, navigation]);
```

### Error Recovery Strategies

#### Graceful Degradation

**UDP Service Fallback**:
```typescript
if (!dgram) {
  // Disable UDP features but keep app functional
  return {
    startService: () => Promise.reject(new Error('UDP not available')),
    sendVolumeCommand: () => { /* No-op */ },
    // ... other disabled methods
  };
}
```

#### Retry Mechanisms

**Command Retry Logic**:
```typescript
const sendCommandWithRetry = async (command: Command, maxRetries: number = 3) => {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      await sendCommand(command);
      return; // Success
    } catch (error) {
      logger.warn('Service', `Command failed (attempt ${attempt}/${maxRetries})`, { error });
      
      if (attempt === maxRetries) {
        throw error; // Final attempt failed
      }
      
      await new Promise(resolve => setTimeout(resolve, attempt * 1000)); // Exponential backoff
    }
  }
};
```

#### User Feedback

**Error Message Display**:
```typescript
const ErrorBoundary = ({ children }: { children: React.ReactNode }) => {
  const [hasError, setHasError] = useState(false);
  
  return hasError ? (
    <View style={styles.errorContainer}>
      <Text style={styles.errorText}>Something went wrong</Text>
      <TouchableOpacity onPress={() => setHasError(false)}>
        <Text>Try Again</Text>
      </TouchableOpacity>
    </View>
  ) : (
    children
  );
};
```

---

## Performance Considerations

### Memory Management

#### Device List Optimization

**Efficient Device Storage**:
```typescript
// Use Map for O(1) device lookup
private devices: Map<string, DeviceInfo> = new Map();

// Cleanup inactive devices
private cleanupInactiveDevices(): void {
  const cutoffTime = Date.now() - this.config.deviceTimeout * 2;
  
  for (const [address, device] of this.devices.entries()) {
    if (device.lastSeen.getTime() < cutoffTime && !device.isConnected) {
      this.devices.delete(address);
      logger.debug('UDPService', `Cleaned up inactive device: ${address}`);
    }
  }
}
```

#### Log Management

**Log Rotation**:
```typescript
private log(level: LogLevel, context: string, message: string): void {
  // Add to internal log storage
  this.logs.push(entry);
  
  // Prevent memory leak with log rotation
  if (this.logs.length > this.maxLogs) {
    this.logs = this.logs.slice(-this.maxLogs); // Keep only recent logs
  }
}
```

### Network Performance

#### UDP Optimization

**Sequence Number Management**:
```typescript
private getNextSequenceNumber(): number {
  this.sequenceNumber = (this.sequenceNumber + 1) % 65536; // Prevent overflow
  return this.sequenceNumber;
}
```

**Batch Operations**:
```typescript
const sendBatchCommands = async (commands: Command[]) => {
  const promises = commands.map(cmd => sendCommand(cmd));
  const results = await Promise.allSettled(promises);
  
  // Process results and handle partial failures
  results.forEach((result, index) => {
    if (result.status === 'rejected') {
      logger.error('BatchOperation', `Command ${index} failed`, { error: result.reason });
    }
  });
};
```

### UI Performance

#### Render Optimization

**Memoized Components**:
```typescript
const DeviceListItem = React.memo(({ device }: { device: DeviceInfo }) => (
  <View style={styles.deviceItem}>
    <Text>{device.name}</Text>
    <StatusIndicator connected={device.isConnected} />
  </View>
));
```

**Virtualized Lists for Large Datasets**:
```typescript
// For future implementation with many devices
import { FlatList } from 'react-native';

<FlatList
  data={devices}
  renderItem={({ item }) => <DeviceListItem device={item} />}
  keyExtractor={(item) => `${item.address}:${item.port}`}
  getItemLayout={(data, index) => ({ length: 70, offset: 70 * index, index })}
  removeClippedSubviews={true}
/>
```

#### State Update Optimization

**Selective Re-rendering**:
```typescript
// Use specific selectors to prevent unnecessary re-renders
const selectedDevice = useSelector((state: RootState) => state.devices.selectedDevice);
const isDiscovering = useSelector((state: RootState) => state.devices.isDiscovering);

// Instead of
const deviceState = useSelector((state: RootState) => state.devices);
```

### Battery Optimization

#### Discovery Interval Management

**Adaptive Discovery**:
```typescript
const adaptDiscoveryInterval = (deviceCount: number) => {
  if (deviceCount === 0) {
    return 3000; // Faster when no devices found
  } else if (deviceCount < 5) {
    return 5000; // Normal interval
  } else {
    return 10000; // Slower when many devices found
  }
};
```

#### Background Behavior

**App State Management**:
```typescript
useEffect(() => {
  const handleAppStateChange = (nextAppState: string) => {
    if (nextAppState === 'background') {
      // Reduce discovery frequency
      udpService.setDiscoveryInterval(30000);
    } else if (nextAppState === 'active') {
      // Resume normal frequency
      udpService.setDiscoveryInterval(5000);
    }
  };
  
  AppState.addEventListener('change', handleAppStateChange);
  return () => AppState.removeEventListener('change', handleAppStateChange);
}, []);
```

---

## Testing Architecture

### Testing Strategy

While comprehensive automated testing is not yet implemented, the architecture supports future testing implementation with clear separation of concerns and dependency injection patterns.

#### Unit Testing Structure

**Service Testing**:
```typescript
// Future UDPService.test.ts
describe('UDPService', () => {
  let service: UDPService;
  let mockSocket: jest.Mocked<any>;
  
  beforeEach(() => {
    mockSocket = createMockSocket();
    service = new UDPService(testConfig, testCallbacks);
    service.setSocket(mockSocket); // Dependency injection
  });
  
  test('should send volume command with correct format', async () => {
    await service.sendVolumeCommand('192.168.1.100:5001', 0.5);
    
    expect(mockSocket.send).toHaveBeenCalledWith(
      expect.any(Buffer),
      0,
      expect.any(Number),
      5001,
      '192.168.1.100',
      expect.any(Function)
    );
  });
});
```

**Protocol Testing**:
```typescript
// Future protocol.test.ts
describe('Protocol Encoding', () => {
  test('should encode volume correctly', () => {
    const encoded = encodeVolume(0.5);
    expect(encoded).toBe(500); // 50% of range 0-1000
  });
  
  test('should create valid header', () => {
    const header = createBasicHeader(MessageType.Ping, StatusType.RequestServer, ...);
    const buffer = encodeHeader(header);
    expect(buffer.byteLength).toBe(12);
  });
});
```

#### Integration Testing

**Hook Testing**:
```typescript
// Future useUDPService.test.ts
import { renderHook, act } from '@testing-library/react-hooks';

describe('useUDPService', () => {
  test('should initialize service correctly', async () => {
    const { result } = renderHook(() => useUDPService());
    
    await act(async () => {
      await result.current.startService();
    });
    
    expect(result.current.isServiceRunning).toBe(true);
  });
});
```

#### Manual Testing Guidelines

**Device Discovery Testing**:
1. Start simulated device: `go run cmd/ppa-cli/main.go simulate`
2. Enable discovery in app
3. Verify device appears in list
4. Check device status updates correctly

**Protocol Testing**:
1. Send volume commands from app
2. Monitor simulated device logs for received commands
3. Verify command format matches protocol specification
4. Test edge cases (min/max volume, invalid inputs)

**Error Handling Testing**:
1. Disconnect network during operation
2. Send malformed UDP packets to app
3. Test with UDP module unavailable (Expo Go)
4. Verify graceful error messages and recovery

---

## Deployment & Build Process

### Build Configuration

#### EAS Build Configuration (`eas.json`)

```json
{
  "cli": { "version": ">= 3.0.0" },
  "build": {
    "development": {
      "developmentClient": true,
      "distribution": "internal",
      "ios": { "simulator": true },
      "android": { "buildType": "apk" }
    },
    "preview": {
      "distribution": "internal"
    },
    "production": {}
  }
}
```

#### App Configuration (`app.json`)

**Platform-specific Settings**:
```json
{
  "expo": {
    "name": "PPA Control",
    "slug": "ppa-control",
    "ios": {
      "supportsTablet": true,
      "bundleIdentifier": "com.ppacontrol.mobile"
    },
    "android": {
      "package": "com.ppacontrol.mobile",
      "permissions": [
        "android.permission.INTERNET",
        "android.permission.ACCESS_NETWORK_STATE",
        "android.permission.ACCESS_WIFI_STATE"
      ]
    },
    "plugins": [
      ["expo-dev-client", { "addGeneratedScheme": false }]
    ]
  }
}
```

### Development Build Process

**Build Commands**:
```bash
# Development build for testing
eas build --profile development --platform ios

# Preview build for stakeholder testing
eas build --profile preview --platform ios

# Production build for app store
eas build --profile production --platform ios
```

**Build Optimization**:
```javascript
// metro.config.js - Optimized for production
const { getDefaultConfig } = require('expo/metro-config');

const config = getDefaultConfig(__dirname);

// Production optimizations
if (process.env.NODE_ENV === 'production') {
  config.transformer.minifierConfig = {
    keep_fnames: false,
    mangle: { keep_fnames: false },
    output: { comments: false },
  };
}
```

### Deployment Strategy

#### Pre-deployment Checklist

**Code Quality**:
- [ ] TypeScript compilation without errors: `npm run type-check`
- [ ] All console.log statements removed or replaced with logger
- [ ] Debug panel access restricted in production builds
- [ ] Error boundaries implemented for crash prevention

**Performance**:
- [ ] Log level set to WARN or ERROR for production
- [ ] Packet logging disabled in production
- [ ] Discovery intervals optimized for battery life
- [ ] Memory leak testing completed

**Testing**:
- [ ] Tested on real devices with actual PPA hardware
- [ ] Network edge cases tested (WiFi switching, poor connectivity)
- [ ] Multi-device scenarios tested with multiple speakers
- [ ] Battery usage profiled during extended use

#### Production Considerations

**Security**:
- Remove development-only code paths
- Implement certificate pinning for future API endpoints
- Validate all user inputs thoroughly
- Implement rate limiting for UDP commands

**Monitoring**:
- Integrate crash reporting (Sentry, Bugsnag)
- Add analytics for user behavior tracking
- Implement performance monitoring
- Set up error alerting for critical failures

**Maintenance**:
- Plan for protocol version updates
- Implement automatic update checks
- Design backward compatibility strategy
- Create rollback procedures for failed deployments

---

## Conclusion

This document provides a comprehensive overview of the PPA Control mobile application architecture. The modular design, comprehensive logging system, and clear separation of concerns create a maintainable and extensible codebase suitable for professional audio control applications.

Key architectural strengths include:

- **Protocol Implementation**: Complete TypeScript implementation of the PPA binary protocol with proper encoding/decoding
- **Network Layer**: Robust UDP communication with fallback handling and comprehensive error management
- **State Management**: Well-structured Redux implementation with optimized selectors and actions
- **Debugging System**: Advanced logging and in-app debugging capabilities for development and troubleshooting
- **User Experience**: Professional-grade interface suitable for live audio environments

Future developers should focus on:

1. **Extending Protocol Support**: Adding new message types and device capabilities as the PPA protocol evolves
2. **Performance Optimization**: Implementing virtualized lists and optimizing render cycles for large device deployments
3. **Testing Implementation**: Adding comprehensive unit and integration tests using the established patterns
4. **Platform Features**: Leveraging platform-specific capabilities like background processing and push notifications

The architecture provides a solid foundation for both current functionality and future enhancements while maintaining the reliability and performance requirements of professional audio applications.