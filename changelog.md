# Documentation Improvements

Enhanced documentation to provide better project understanding and usage instructions.

- Expanded README.md with comprehensive project description
- Added detailed installation and usage instructions
- Included development and contribution guidelines
- Added project structure documentation
- Enhanced CLI and GUI usage examples
- Updated pcap-dump tool documentation with accurate command-line flags and examples
- Added interface selection and timeout configuration documentation 

# Protocol Documentation

Added comprehensive protocol documentation in doc/protocol.md that details the PPA protocol structure, message types, and communication flow.

- Created detailed documentation of protocol header structure
- Documented all message types and their purposes
- Added status type documentation for client and server messages
- Included message-specific payload structures
- Added implementation notes and communication flow details

# Protocol Documentation Expansion

Significantly expanded the protocol documentation with detailed message layouts and encoding information.

- Added comprehensive Device Data message structure documentation
- Added detailed Live Command message documentation including path structure
- Added equalizer types and their values
- Added value encoding specifications for different parameter types
- Added hierarchical relationships between level types
- Added unknown message type documentation

# Add Packet Timestamps to pcap Output

Added packet timestamps to the pcap packet printing output to help with debugging timing-related issues and packet sequence analysis.

- Added timestamp display to packet output in cmd/pcap/packet_print.go

# Add Time Offset to pcap Output

Enhanced packet timestamp display by adding time offset from previous packet to help analyze timing patterns and packet intervals.

- Added time offset display (in ms/s) between consecutive packets in pcap output
- Added lastPacketTime tracking to PacketHandler

# Enhanced pcap Output with Styled Terminal Display

Added styled terminal output using lipgloss to improve readability and visual organization of packet information.

- Added color-coded styling for different message components
- Added message direction indication (Device → Client or Client → Device)
- Enhanced visual hierarchy with bold and italic styles for different fields
- Improved readability of hexdump output with distinct styling

# Add JSON/YAML Output to pcap Tool

Added JSON and YAML output formats to the pcap packet capture tool to support machine-readable output formats.

- Added --output-format flag to select between text, json, and yaml output
- Added structured PacketData type for consistent output formatting
- Refactored packet handler to use common data structure for all output formats
- Maintained backward compatibility with existing text output format

# Improve JSON/YAML Output Formats

Enhanced JSON and YAML output formats to better support streaming and multiple documents.

- Changed JSON output to use JSON Lines (JSONL) format for better streaming support
- Added YAML document separators (---) between packet entries
- Removed pretty-printing indentation from JSON for more compact output

# Add Separate JSON and JSONL Output Formats

Added distinct JSON array and JSON Lines output formats for different use cases.

- Added separate 'json' and 'jsonl' output formats
- JSON format outputs a properly formatted array of all packets
- JSONL format outputs one JSON object per line for streaming
- Updated command-line help to clarify the difference between formats

# Add ppa-cli Command Documentation

Added comprehensive documentation for the ppa-cli command in cmd/ppa-cli/README.md.

- Created detailed documentation of all subcommands (ping, recall, simulate, volume, udp-broadcast)
- Added documentation for global flags and command-specific flags
- Included practical usage examples for common tasks
- Added network testing and device simulation examples

# Add ui-test Command Documentation

Added comprehensive documentation for the ui-test command in cmd/ui-test/README.md.

- Created detailed documentation of the graphical user interface
- Added installation prerequisites and build instructions
- Documented all command-line flags and configuration options
- Added interface overview with component descriptions
- Included log management and upload functionality documentation
- Added practical usage examples

# Web Interface for PPA Control

Added a web-based interface for controlling PPA devices using htmx and templ, providing:
- Device connection management
- Preset recall buttons
- Volume control
- Real-time command logging

## Changes
- Created new web server in cmd/ppa-web
- Implemented templ templates for the UI
- Added HTMX for dynamic updates
- Integrated Bootstrap for styling
- Added real-time log window for command feedback

# Improved Web Interface with Background Ping

Enhanced the web interface with proper client management and background ping:
- Added Server struct for better state management
- Implemented background ping loop with status updates
- Added real-time connection status display
- Improved thread safety with mutex locks
- Added periodic status updates in the UI

## Changes
- Refactored main.go to use Server struct
- Added background ping loop from ppa-cli
- Added status bar with connection state
- Added automatic status updates every 2 seconds
- Improved error handling and state management

# Add Detailed Packet Logging to Web Interface

Enhanced the web interface with detailed packet logging similar to pcap:
- Added packet details logging to browser console
- Added styled console output for better readability
- Added hex dump display for packet payloads
- Added packet direction indicators
- Added timestamp and source/destination information

## Changes
- Added PacketInfo struct for structured packet information
- Added packet logging to JavaScript console
- Added styled console output with color coding
- Added detailed packet information display
- Added hex dump support for packet payloads

# Reorganize Web Interface Code Structure

Improved code organization of the web interface:
- Created separate server package for better modularity
- Moved AppState and PacketInfo to their own types file
- Improved template integration with server package
- Added better type safety and package boundaries

## Changes
- Created cmd/ppa-web/server package
- Moved server logic to dedicated package
- Separated types into types.go
- Updated templates to use server package types
- Improved code organization and maintainability

# WebMIDI Test Application

Added a test application in test/web-midi to demonstrate WebMIDI capabilities. The application provides a web interface for:
- Detecting and listing MIDI devices
- Monitoring MIDI input messages
- Sending test MIDI notes to output devices
- Real-time device connection monitoring

# Improve WebMIDI Security Handling

Enhanced WebMIDI test application with better security error handling and documentation:
- Added detailed security error messages with fix instructions
- Updated documentation with browser-specific configuration steps
- Added security requirements section to README
- Disabled sysex access for better security

Add templ generation directive
Added a generate.go file with a go:generate directive to automate templ template generation, making it easier to maintain and update templates.

# Improve Timer Handling in Ping Command

Enhanced timer handling and cleanup in the ping command to prevent resource leaks and ensure proper cleanup on context cancellation.

## Changes
- Added proper timer cleanup in ping loop
- Added timer stop on context cancellation
- Added timer cleanup before channel processing
- Improved resource management in ping loop

# Add Volume Control Command

Added a new volume control command to the CLI that allows setting master volume levels on PPA devices:
- Volume range from 0.0 to 1.0 (maps to -72dB to +20dB)
- Support for multiple devices via discovery or direct addressing
- Optional continuous volume updates with loop mode
- Immediate volume application to newly discovered devices

## Changes
- Implemented volume command in cmd/ppa-cli/cmds/volume.go
- Added volume level validation
- Added loop mode for continuous volume updates
- Added proper timer and context cancellation handling
- Added volume command documentation

# Refactor CLI Commands for Better Code Organization

Improved code organization and reliability of CLI commands by extracting common functionality:
- Created shared command setup and execution infrastructure
- Unified discovery and multiclient handling
- Standardized error handling and context cancellation
- Improved resource cleanup across all commands

## Changes
- Added common.go with shared command functionality
- Refactored ping, recall, and volume commands to use common code
- Improved timer cleanup and context cancellation
- Added consistent error handling patterns
- Standardized command initialization and cleanup

# Improve Command Context Management

Enhanced command context management with a new CommandContext type:
- Added CommandContext struct to encapsulate command state and resources
- Added idiomatic methods for context and resource management
- Improved error handling and cleanup patterns
- Simplified command implementation with better abstraction

## Changes
- Added CommandContext type with resource management methods
- Refactored commands to use the new CommandContext
- Added proper cleanup and error handling methods
- Improved code readability and maintainability
- Added consistent context access patterns

# Client Architecture Documentation

Added comprehensive documentation explaining the client architecture, including single device clients, multi-client management, and the discovery system. This documentation helps developers understand the system's components and their interactions.

- Added `lib/client/client.md` with detailed explanations of client components
- Documented channel handling and lifecycle management
- Added code snippets and best practices

# Improve Client Interface Design

Enhanced the client interface design for better separation of concerns and more idiomatic Go:

- Split Client interface into Commander and Client interfaces
- Fixed channel ownership and direction in interfaces
- Updated channel types to use send-only channels where appropriate
- Improved code organization with dedicated interface file

# Improve Error Handling in Client Package

Enhanced error handling across the client package for better reliability and debugging:

- Added custom error types for domain-specific errors
- Replaced panics with proper error handling
- Added error propagation through error channels
- Added safe command execution with panic recovery
- Improved error logging and context
- Added proper error handling for client lifecycle operations

# Refactor Web Server to Use CommandContext

The web server has been refactored to use the common CommandContext functionality instead of reimplementing its own client management. This improves code reuse and consistency across the codebase.

- Added FromCobraCommand constructor to create server from cobra commands
- Added normal constructor for direct use from ppa-web
- Added SetupContext and GetMultiClient methods to CommandContext
- Refactored Server struct to use CommandContext for all client operations

# Extract Web Server Handler

Improved code organization by extracting HTTP handlers into a dedicated Handler struct:
- Created Handler struct to encapsulate all HTTP handlers
- Added TemplateProvider interface to break circular dependencies
- Improved separation of concerns between server and handler logic
- Added proper interface abstractions for template rendering

# Add Device Discovery UI

Added a device discovery UI to the web interface that allows users to:
- Start/stop device discovery
- See discovered devices in real-time
- Connect to discovered devices with one click
- Monitor network interfaces

## Changes
- Added discovery section to web UI with SSE updates
- Added device discovery state management to server
- Added real-time device list updates
- Added interface monitoring
- Added discovery command-line flags