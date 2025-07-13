# PPA-Control CLI Commands Reference

This document describes all available commands in the PPA-Control CLI with their parameters and usage examples.

## Overview

The PPA-Control CLI provides commands for managing Programmable Power Array (PPA) devices with support for both human-readable logging and structured data output.

All commands support the glazed framework with:
- **Structured Output**: JSON, CSV, table, YAML, markdown, etc.
- **Dual Mode**: Classic logging and structured data modes
- **Parameter Layers**: Consistent PPA configuration across commands
- **Discovery**: Automatic device discovery via UDP broadcast

## Global Flags

Available across all commands:

```bash
--dump-mem-profile string   Dump memory profile to file
--log-format string         Log format (json, text) (default "text")
--log-level string          Log level (default "debug")
--track-leaks               Track memory and goroutine leaks
--with-caller               Log caller information
```

## Common PPA Parameters

Most commands share these PPA connection parameters:

```bash
-a, --addresses string       Addresses to connect to, comma separated
-c, --component-id int       Component ID to use for devices (default 255)
    --componentId int        Legacy alias for component-id
-d, --discover              Send broadcast discovery messages (default true)
    --interfaces strings     Interfaces to use for discovery
-p, --port int              Port to connect to (default 5001)
```

## Glazed Output Parameters

Commands support extensive output formatting options:

```bash
--structured-output         Switch to structured output mode
--output string            Output format: table, json, csv, yaml, etc. (default "table")
--fields strings           Fields to include in output (default [all])
--filter strings           Fields to remove from output
--sort-by strings          Sort by field(s)
--select string            Select single field for output
--template string          Custom Go template for output
```

---

## Commands

### 1. ping

Send periodic ping messages to PPA servers with structured output support.

**Usage:**
```bash
ppa-cli ping [flags]
```

**Description:**
Sends ping messages every 5 seconds to specified PPA devices. Supports both classic logging mode and structured output for monitoring and automation.

**Examples:**
```bash
# Classic mode - human readable logs
ppa-cli ping --addresses 192.168.1.100

# Structured JSON output
ppa-cli ping --addresses 192.168.1.100 --structured-output --output json

# Discovery mode with structured table output
ppa-cli ping --discover --structured-output --output table

# Multiple addresses with CSV output
ppa-cli ping --addresses 192.168.1.100,192.168.1.101 --structured-output --output csv
```

**Structured Output Events:**
- `message_received`: Ping responses from devices
- `discovery_message`: Device discovery events
- `unknown_message`: Unrecognized message types

---

### 2. volume

Set the volume of one or more PPA clients with structured output support.

**Usage:**
```bash
ppa-cli volume [flags]
```

**Volume-Specific Parameters:**
```bash
-v, --volume float          Volume level (0.0 to 1.0, required)
-l, --loop                  Loop continuously sending volume commands
```

**Description:**
Sets volume levels on PPA devices. Volume must be between 0.0 (silent) and 1.0 (maximum). Can loop continuously for testing.

**Examples:**
```bash
# Set volume to 50% on specific device
ppa-cli volume --addresses 192.168.1.100 --volume 0.5

# Set volume with structured JSON output
ppa-cli volume --addresses 192.168.1.100 --volume 0.8 --structured-output --output json

# Loop volume commands with discovery
ppa-cli volume --volume 0.3 --loop --discover --structured-output --output table

# Set volume on multiple devices
ppa-cli volume --addresses 192.168.1.100,192.168.1.101 --volume 0.6
```

**Structured Output Events:**
- `volume_command_sent`: When volume command is transmitted
- `message_received`: Device responses to volume changes
- `discovery_message`: New device discoveries
- `new_client_volume_sent`: Volume sent to newly discovered clients

---

### 3. recall

Recall presets on PPA servers with structured output support.

**Usage:**
```bash
ppa-cli recall [flags]
```

**Recall-Specific Parameters:**
```bash
    --preset int            Preset index to recall (required)
-l, --loop                  Loop continuously sending recall commands
```

**Description:**
Recalls (activates) a specific preset by index on PPA devices. Presets are device-specific configurations.

**Examples:**
```bash
# Recall preset 5 on specific device
ppa-cli recall --addresses 192.168.1.100 --preset 5

# Recall preset with structured output
ppa-cli recall --addresses 192.168.1.100 --preset 3 --structured-output --output json

# Loop recall commands with discovery
ppa-cli recall --preset 1 --loop --discover --structured-output --output table

# Recall on multiple devices
ppa-cli recall --addresses 192.168.1.100,192.168.1.101 --preset 2
```

**Structured Output Events:**
- `recall_initiated`: When recall command starts
- `recall_sent`: Each time recall message is transmitted
- `recall_response`: Device responses to recall commands
- `discovery_message`: Device discovery events
- `new_client_recall`: Recall sent to newly discovered clients

---

### 4. udp-broadcast

Send and receive UDP broadcast messages with structured output support.

**Usage:**
```bash
ppa-cli udp-broadcast [flags]
```

**UDP-Specific Parameters:**
```bash
-a, --address string        Address to listen on or send to
-p, --port int             UDP port to use (default 8080)
-s, --server               Run as server (listen mode)
-i, --interface string     Interface to bind to
```

**Description:**
Sends UDP broadcast messages or runs as a UDP server for testing network connectivity and broadcast functionality.

**Examples:**
```bash
# Send UDP broadcast to localhost:8080
ppa-cli udp-broadcast --address 127.0.0.1 --port 8080

# Run as UDP server with structured output
ppa-cli udp-broadcast --server --port 9999 --structured-output --output json

# Send to specific address with table output
ppa-cli udp-broadcast --address 192.168.1.255 --port 5555 --structured-output --output table

# Server mode on specific interface
ppa-cli udp-broadcast --server --interface eth0 --port 8080
```

**Structured Output Events:**
- `socket_created`: When UDP socket is created
- `interface_bind_attempt`: Interface binding attempts
- `sending_message`: Before sending UDP message
- `message_sent`: After successful UDP transmission
- `message_received`: When UDP message is received
- `read_error`/`send_error`: Network error events

---

## Output Formats

The CLI supports multiple output formats via the `--output` flag:

| Format | Description | Use Case |
|--------|-------------|----------|
| `table` | Human-readable ASCII table | Terminal viewing |
| `json` | JSON array of objects | API integration, scripting |
| `csv` | Comma-separated values | Spreadsheets, data analysis |
| `yaml` | YAML format | Configuration files |
| `markdown` | Markdown table | Documentation |
| `template` | Custom Go templates | Custom formatting |

## Logging Configuration

All commands support comprehensive logging configuration:

```bash
--log-level string          Set logging level: trace, debug, info, warn, error, fatal
--log-format string         Log format: text (human-readable) or json (structured)
--log-file string           Write logs to file instead of stderr
--with-caller              Include source file and line numbers in logs
--logstash-enabled         Send logs to Logstash for centralized collection
```

## Best Practices

1. **Use structured output for automation**: `--structured-output --output json`
2. **Filter output for specific data**: `--fields timestamp,from,status`
3. **Sort results for analysis**: `--sort-by timestamp`
4. **Use discovery for unknown networks**: `--discover`
5. **Save output to files**: `--output-file results.json`
6. **Monitor with JSON logs**: `--log-format json --log-level info`

## Error Handling

- Invalid parameters show helpful error messages
- Network errors are logged with context
- Discovery failures are non-fatal
- Structured output includes error events
- Exit codes indicate success (0) or failure (non-zero)
