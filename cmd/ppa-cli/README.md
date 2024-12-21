# ppa-cli

Command line interface for the PPA protocol, allowing interaction with PPA DSP devices.

## Global Flags

These flags are available for all subcommands:

- `--log-level string`: Set log level (debug, info, warn, error, fatal) (default "debug")
- `--log-format string`: Log format (json, text) (default "text")
- `--with-caller`: Log caller information
- `--dump-mem-profile string`: Dump memory profile to file
- `--track-leaks`: Track memory and goroutine leaks

## Subcommands

### ping

Send ping messages to discover and monitor PPA devices.

```bash
ppa-cli ping [flags]
```

#### Flags
- `-a, --addresses string`: Addresses to ping, comma separated
- `-d, --discover`: Send broadcast discovery messages (default false)
- `--interfaces []string`: Interfaces to use for discovery
- `-c, --componentId uint`: Component ID to use for devices (default 0xFF)
- `-p, --port uint`: Port to ping on (default 5001)

### recall

Recall a preset by index on one or more PPA devices.

```bash
ppa-cli recall [flags]
```

#### Flags
- `-a, --addresses string`: Addresses to ping, comma separated
- `-d, --discover`: Send broadcast discovery messages (default true)
- `-l, --loop`: Send recalls in a loop (default true)
- `-c, --componentId uint`: Component ID to use for devices (default 0xFF)
- `--preset int`: Preset to recall (default 0)
- `-p, --port uint`: Port to ping on (default 5001)

### simulate

Start a simulated PPA device for testing purposes.

```bash
ppa-cli simulate [flags]
```

#### Flags
- `-i, --interface string`: Bind listener to interface
- `-a, --address string`: Address to listen on (default "localhost")
- `-p, --port uint`: Port to listen on (default 5001)

### volume

Set the volume of one or more PPA devices.

```bash
ppa-cli volume [flags]
```

### udp-broadcast

Utility command to send and receive UDP broadcast messages. Useful for testing network connectivity.

```bash
ppa-cli udp-broadcast [flags]
```

#### Flags
- `-a, --address string`: Address to listen on (default "localhost")
- `-s, --server`: Run as server
- `-p, --port uint`: Port to listen on (default 5001)
- `-i, --interface string`: Interface to bind to

## Examples

### Discover Devices
```bash
# Discover devices on all interfaces
ppa-cli ping --discover

# Discover devices on specific interfaces
ppa-cli ping --discover --interfaces eth0,wlan0
```

### Recall Presets
```bash
# Recall preset 0 on a specific device
ppa-cli recall --addresses 192.168.1.100 --preset 0

# Recall presets in a loop with discovery enabled
ppa-cli recall --discover --loop
```

### Simulate a Device
```bash
# Start a simulated device on localhost
ppa-cli simulate

# Start a simulated device on a specific interface and address
ppa-cli simulate --interface eth0 --address 192.168.1.200
```

### Test Network Connectivity
```bash
# Start a UDP broadcast server
ppa-cli udp-broadcast --server

# Send a UDP broadcast message
ppa-cli udp-broadcast --address 192.168.1.255
``` 