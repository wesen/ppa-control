# PPA Control Mobile App

A React Native mobile application for controlling PPA DSP speaker systems using UDP communication.

## Features

- **Device Discovery**: Automatic discovery of PPA devices on the network via UDP broadcast
- **Manual Device Entry**: Add devices by IP address when auto-discovery isn't available
- **Volume Control**: Real-time volume adjustment with visual feedback
- **Preset Management**: Recall and manage presets (1-16) across devices
- **Multi-Device Control**: Synchronized control of multiple speakers
- **Settings**: Comprehensive configuration options for network and app behavior

## Architecture

### Protocol Implementation
- Full PPA binary protocol implementation in TypeScript
- UDP socket communication using `react-native-udp`
- Message encoding/decoding with proper binary formats
- Support for ping, volume, preset, and device data messages

### State Management
- Redux Toolkit for centralized state management
- Separate slices for devices, control, and settings
- Optimized for real-time updates and concurrent device management

### UI/UX Design
- Professional audio-focused interface suitable for live environments
- Dark/light theme support
- Haptic feedback and accessibility features
- Real-time status feedback and error handling

## Technical Requirements

### Development Build Required
This app uses `react-native-udp` which requires native code. You'll need to create a development build:

```bash
# Install Expo CLI and EAS CLI
npm install -g @expo/cli eas-cli

# Install dependencies
npm install

# Create development build
eas build --profile development --platform ios
# or
eas build --profile development --platform android
```

### Network Permissions
The app requires network access permissions:
- **Android**: Internet, network state, and WiFi state permissions
- **iOS**: Local network access (configured automatically)

## Getting Started

### 1. Install Dependencies
```bash
npm install
```

### 2. Start Development Server
```bash
npm run start
```

### 3. Run on Device/Simulator

#### For iOS Simulator (macOS only)
```bash
npm run ios
```

#### For Android Emulator
```bash
npm run android
```

#### For Web Development
```bash
npm run web
```

**Note**: UDP functionality won't work in web or Expo Go. You need a development build for full functionality.

## Project Structure

```
src/
├── hooks/           # Custom React hooks
│   └── useUDPService.ts
├── navigation/      # Navigation configuration
│   └── AppNavigator.tsx
├── protocol/        # PPA protocol implementation
│   ├── types.ts
│   └── encoding.ts
├── screens/         # App screens
│   ├── DiscoveryScreen.tsx
│   ├── ControlScreen.tsx
│   ├── MultiDeviceScreen.tsx
│   └── SettingsScreen.tsx
├── services/        # Business logic services
│   └── UDPService.ts
└── store/          # Redux store and slices
    ├── index.ts
    ├── deviceSlice.ts
    ├── controlSlice.ts
    └── settingsSlice.ts
```

## Usage

### Device Discovery
1. Open the app and navigate to the "Discovery" tab
2. Toggle discovery on to automatically find devices
3. Or manually enter device IP addresses
4. Tap on a device to connect and control it

### Volume Control
1. Select a device from the discovery screen
2. Use the volume slider for real-time control
3. Tap mute/unmute for instant audio control
4. View real-time feedback and status

### Preset Management
1. In the control screen, scroll to the preset section
2. Tap preset buttons (1-16) to recall saved configurations
3. Current preset is highlighted in blue

### Multi-Device Control
1. Navigate to the "Multi-Device" tab
2. Select multiple devices using the switches
3. Use master volume control to adjust all selected devices
4. Sync presets across all selected devices

### Settings
- Configure network settings (ports, timeouts, broadcast address)
- Adjust UI preferences (theme, haptic feedback)
- Set app behavior (auto-discovery, confirmations)
- Enable debug logging for troubleshooting

## Network Configuration

### Default Settings
- **Port**: 5001 (UDP)
- **Discovery Interval**: 5 seconds
- **Device Timeout**: 30 seconds
- **Broadcast Address**: 255.255.255.255

### Firewall Considerations
Ensure that UDP port 5001 is open on your network. The app needs to:
- Send UDP broadcasts for device discovery
- Receive UDP responses from devices
- Send UDP commands to specific device IPs

## Testing with Simulation

You can test the app using the PPA control system's simulation feature:

```bash
# In the main ppa-control directory
go run cmd/ppa-cli/main.go simulate --address 0.0.0.0 --log-level info
```

This creates a simulated PPA device that responds to the mobile app's commands.

## Protocol Implementation Details

### Message Structure
- 12-byte header + optional payload
- Little-endian encoding for multi-byte values
- Sequence numbers for request/response tracking
- Device unique IDs for targeting specific devices

### Supported Commands
- **Ping**: Device discovery and keepalive
- **Volume Control**: Master volume adjustment (0-100%)
- **Preset Recall**: Load saved presets (0-255)
- **Device Data**: Query device information

### Volume Encoding
- Range: 0.0 (muted) to 1.0 (maximum)
- Internal: 0x00 (-80dB) to 0x3E8 (+20dB)
- Real-time updates with optimistic UI

## Troubleshooting

### Common Issues

#### App Can't Find Devices
- Ensure devices are on the same network
- Check firewall settings for UDP port 5001
- Verify network permissions are granted
- Try manual IP entry

#### Commands Not Working
- Check device connection status
- Verify UDP service is running
- Enable packet logging in settings
- Test with simulated device

#### Performance Issues
- Reduce discovery interval in settings
- Disable packet logging
- Clear app data if needed

### Debug Logging
Enable debug logging in Settings > Debug & Logging to see detailed network activity and protocol messages.

## Building for Production

### iOS
```bash
eas build --platform ios
```

### Android
```bash
eas build --platform android
```

### App Store Deployment
- Update version number in `app.json`
- Add proper app icons and splash screens
- Configure bundle identifiers
- Submit through EAS Submit or manually

## Contributing

1. Follow TypeScript best practices
2. Use Redux for state management
3. Implement proper error handling
4. Add loading states for network operations
5. Test with both real and simulated devices

## License

This project is part of the PPA Control System. See the main project license for details.