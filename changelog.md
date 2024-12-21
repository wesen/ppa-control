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