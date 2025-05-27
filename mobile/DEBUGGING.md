# PPA Control Mobile App - Debugging Guide

## Overview

This guide covers debugging strategies, logging, and troubleshooting for the PPA Control mobile application.

## Current Error Analysis

### UDP Service Error: "Cannot read property 'createSocket' of null"

**Root Cause**: You're likely running the app in Expo Go or web mode, where the `react-native-udp` native module is not available.

**Solutions**:

1. **Create Development Build** (Recommended):
   ```bash
   # Install EAS CLI
   npm install -g eas-cli
   
   # Configure EAS project
   eas build:configure
   
   # Build for your platform
   eas build --profile development --platform ios
   # or
   eas build --profile development --platform android
   ```

2. **Use Simulation Mode** (For Testing):
   - The app now includes fallback handling for when UDP is not available
   - Check the debug panel for detailed error information

## Debugging Tools & Strategies

### 1. In-App Debug Panel

**Access**: Settings → Open Debug Panel

**Features**:
- Real-time log viewing with filtering
- App state inspection
- Log level control (Debug, Info, Warn, Error)
- Context-based filtering (UDP, Network, User actions, etc.)
- Log export functionality

**Usage**:
```javascript
// Access logger globally in development
global.logger.info('Context', 'Your message', { data: 'optional' });

// Change log level
global.logger.setLevel(global.LogLevel.DEBUG);
```

### 2. Console Logging

**Expo Dev Tools Console**:
```bash
# Start with logging
expo start --dev-client

# In another terminal, view logs
expo logs
```

**Log Format**:
```
[timestamp] LEVEL CONTEXT        message | Data: {...} | Error: ...
🐛 [12:34:56.789] DEBUG UDPService      Creating UDP socket... | Data: {...}
ℹ️  [12:34:56.790] INFO  UDPService      Socket created successfully
⚠️  [12:34:56.791] WARN  UDPService      Discovery timer already running
❌ [12:34:56.792] ERROR UDPService      Failed to bind socket | Error: ...
```

### 3. Metro Bundler Logs

```bash
# Start with verbose logging
npx expo start --dev-client --verbose

# Or with specific log level
npx expo start --dev-client --max-workers 1
```

### 4. Platform-Specific Debugging

#### iOS Debugging:
```bash
# View iOS simulator logs
xcrun simctl spawn booted log stream --predicate 'process == "Expo Go"'

# Or for development build
xcrun simctl spawn booted log stream --predicate 'process == "PPA Control"'
```

#### Android Debugging:
```bash
# View Android emulator logs
adb logcat -s "ReactNativeJS" "ExpoKit"

# Filter for specific tags
adb logcat | grep -i "ppa\|udp\|expo"
```

## Development Environment Setup

### 1. Development Build Setup

Create `eas.json`:
```json
{
  "cli": {
    "version": ">= 3.0.0"
  },
  "build": {
    "development": {
      "developmentClient": true,
      "distribution": "internal",
      "ios": {
        "simulator": true
      },
      "android": {
        "buildType": "apk"
      }
    },
    "preview": {
      "distribution": "internal"
    },
    "production": {}
  },
  "submit": {
    "production": {}
  }
}
```

### 2. Required Dependencies

```bash
# Core dependencies for UDP functionality
npm install react-native-udp expo-dev-client

# Development tools
npm install --save-dev @types/react-native
```

### 3. Metro Configuration

Ensure `metro.config.js` includes:
```javascript
const { getDefaultConfig } = require('expo/metro-config');

const config = getDefaultConfig(__dirname);

config.resolver.alias = {
  ...config.resolver.alias,
  'react-native-udp': require.resolve('react-native-udp'),
};

module.exports = config;
```

## Common Issues & Solutions

### Issue 1: UDP Module Not Found

**Symptoms**:
- "react-native-udp not available" error
- UDP functionality completely disabled

**Solutions**:
1. Ensure you're using a development build, not Expo Go
2. Check that `react-native-udp` is in `package.json`
3. Rebuild the development build after adding dependencies

### Issue 2: Network Discovery Not Working

**Debugging Steps**:
1. Check debug panel for UDP service status
2. Verify network permissions (especially on Android)
3. Test with simulated device:
   ```bash
   # In main project directory
   go run cmd/ppa-cli/main.go simulate --address 0.0.0.0
   ```

**Check Network Configuration**:
- Ensure device and speakers are on same network
- Check firewall settings for UDP port 5001
- Verify broadcast address in settings

### Issue 3: App Crashes on Startup

**Debugging**:
1. Check console for JavaScript errors
2. Use debug panel to see initialization logs
3. Start with minimal functionality:
   ```javascript
   // Disable auto-discovery in settings
   settings.autoDiscovery = false;
   ```

### Issue 4: Performance Issues

**Debugging**:
1. Monitor log frequency in debug panel
2. Adjust discovery interval in settings
3. Reduce log level to INFO or WARN
4. Check for memory leaks in device list

## Log Categories and Contexts

### Main Contexts:
- **UDPService**: Core UDP communication and socket management
- **Network**: HTTP requests and responses
- **User**: User interactions and actions
- **State**: Redux state changes
- **Logger**: Logging system itself

### Log Levels:
- **DEBUG**: Detailed execution flow (development only)
- **INFO**: Important events and status changes
- **WARN**: Potential issues that don't break functionality
- **ERROR**: Errors that affect functionality

## Testing Strategies

### 1. Unit Testing

```bash
# Run TypeScript checks
npm run type-check

# Future: Add Jest tests
npm test
```

### 2. Integration Testing with Simulation

```bash
# Terminal 1: Start simulated device
go run cmd/ppa-cli/main.go simulate --address 127.0.0.1

# Terminal 2: Run mobile app and test discovery
expo start --dev-client
```

### 3. Network Testing

```bash
# Test UDP connectivity
nc -u 127.0.0.1 5001

# Monitor network traffic
sudo tcpdump -i any port 5001
```

### 4. Device Testing Checklist

- [ ] Auto-discovery finds simulated device
- [ ] Manual IP entry works
- [ ] Volume commands send successfully
- [ ] Preset recall functions
- [ ] Error handling displays properly
- [ ] Debug panel shows relevant logs

## Performance Monitoring

### Memory Usage:
- Monitor device list growth
- Check for retained UDP sockets
- Clear logs periodically

### Network Usage:
- Monitor UDP packet frequency
- Adjust discovery intervals for battery life
- Track failed network operations

### UI Performance:
- Check for excessive re-renders
- Monitor Redux state updates
- Optimize log display frequency

## Advanced Debugging

### Remote Debugging:
```bash
# Enable remote debugging in development build
expo start --dev-client --tunnel
```

### React DevTools:
```bash
# Install React DevTools
npm install -g react-devtools

# Connect to running app
react-devtools
```

### Redux DevTools:
- Use Flipper for React Native debugging
- Monitor Redux state changes
- Time-travel debugging

## Troubleshooting Workflow

1. **Check App State**: Use debug panel to verify current state
2. **Review Logs**: Filter by error level and relevant context
3. **Test Components**: Isolate issues to specific functionality
4. **Network Analysis**: Verify UDP communication works
5. **Platform Testing**: Test on different devices/simulators
6. **Fallback Mode**: Test app behavior without UDP

## Getting Help

### Information to Include:
- Platform (iOS/Android/Simulator)
- Expo CLI version
- Development build or Expo Go
- Full error logs from debug panel
- Network configuration
- Steps to reproduce

### Export Debug Information:
1. Open Debug Panel
2. Set log level to DEBUG
3. Reproduce the issue
4. Use "Export" button to share logs

## Production Considerations

### Before Release:
- Set default log level to WARN or ERROR
- Disable packet logging
- Remove debug panel from production builds
- Add crash reporting (Sentry, Bugsnag)
- Test on real network environments

### Monitoring:
- Add analytics for key user flows
- Monitor UDP service success rates
- Track device discovery performance
- Log critical errors to external service