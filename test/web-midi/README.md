# WebMIDI Test Application

This is a simple web application that demonstrates the capabilities of the Web MIDI API, built using Go, HTMX, and Bootstrap. The application allows you to interact with MIDI devices connected to your computer through your web browser.

## Security Requirements

WebMIDI API has specific security requirements:

1. **Secure Context**: The application must be accessed via:
   - `localhost` (recommended for testing)
   - HTTPS connection (for production)

2. **Browser Configuration**:
   - **Chrome/Edge**: Works by default on localhost
   - **Firefox**: Requires enabling `dom.webmidi.enabled` in `about:config`
   - **Opera**: Works by default on localhost
   - **Safari**: Not supported

## What is WebMIDI?

The Web MIDI API is a specification that allows web browsers to interact with MIDI (Musical Instrument Digital Interface) devices. It provides a standardized way to:

- Enumerate and select MIDI input and output devices
- Send and receive MIDI messages
- Listen for MIDI device connection/disconnection events

MIDI messages are used to communicate various musical events, such as:
- Note On/Off events (when a key is pressed/released)
- Control Change messages (for knobs, sliders, etc.)
- Program Change messages (for selecting instruments/patches)
- System Exclusive messages (device-specific commands)

## Features

This application demonstrates several key WebMIDI features:

1. **Device Detection**
   - Automatically detects all connected MIDI input and output devices
   - Shows device names and manufacturers
   - Updates in real-time when devices are connected or disconnected

2. **MIDI Input Monitoring**
   - Enable/disable monitoring for each input device
   - Displays incoming MIDI messages with timestamps
   - Decodes common MIDI message types (Note On/Off, Control Change)

3. **MIDI Output Testing**
   - Send test notes to any connected output device
   - Demonstrates basic MIDI message construction and timing

## Technical Implementation

### Frontend (HTML/JavaScript)

The frontend uses:
- **Bootstrap** for styling and layout
- **HTMX** for potential future dynamic updates
- **WebMIDI API** for MIDI device interaction

Key JavaScript functions:
- `requestMIDIAccess()`: Initializes WebMIDI and requests device access
- `onMIDIMessage()`: Handles incoming MIDI messages
- `sendTestNote()`: Demonstrates sending MIDI messages
- `toggleInput()`: Enables/disables MIDI input monitoring

### Backend (Go)

A simple Go HTTP server that:
- Serves static files
- Provides the main HTML page
- Could be extended for more complex MIDI processing

## Getting Started

1. Clone the repository
2. Run the server:
   ```bash
   go run test/web-midi/main.go
   ```
3. Open `http://localhost:8080` in your browser
4. Connect MIDI devices to your computer
5. Allow MIDI access when prompted by the browser

## Browser Support

WebMIDI is supported in:
- Chrome/Chromium (all platforms)
- Edge (all platforms)
- Opera (all platforms)
- Firefox (behind a flag)

Note: Safari currently does not support WebMIDI.

## Security Considerations

The Web MIDI API requires explicit user permission to access MIDI devices. This permission must be granted by the user when the application first requests MIDI access.

## MIDI Message Format

MIDI messages in this application are displayed in a human-readable format:

```
[timestamp] Channel X: Message Type: Data
```

Common message types:
- Note On: `[time] Channel X: Note On: note (velocity: val)`
- Note Off: `[time] Channel X: Note Off: note (velocity: val)`
- Control Change: `[time] Channel X: Control Change: controller (value: val)`

Raw MIDI data is also shown for other message types in hexadecimal format.

## Future Enhancements

Possible improvements could include:
- MIDI file playback
- Virtual MIDI keyboard
- MIDI routing between inputs and outputs
- MIDI message filtering and transformation
- Recording and exporting MIDI data

## Contributing

Feel free to submit issues and enhancement requests! 