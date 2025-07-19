# PPA Ping/Keepalive Protocol Analysis

## Overview
Analysis of PPA ping/keepalive packet capture from `ppa-ops-20250713-180400-03-ping-keepalive.pcap`.

## Key Findings

### Message Types Identified
1. **Ping Messages** (Type 0x01): Device→Client communication with payload structure
2. **Discovery Messages** (Type 0x02): Connection establishment 
3. **Live Commands** (Type 0x09): Real-time control messages
4. **Device Data** (Type 0x03): Device status and metadata
5. **Preset Recall** (Type 0x04): Preset management

### Communication Pattern
- **Source**: 192.168.0.200:5001 (PPA Device)
- **Destination**: f.50071 (Client - hostname redacted)
- **Protocol**: UDP
- **Port**: 5001 (PPA standard)

### Timing Analysis

#### Initial Connection Sequence (18:08:22-18:08:23)
```
1752444502.877096: Device ping (12 bytes)
1752444503.032893: Client discovery request (16 bytes) 
1752444503.037230: Device discovery response (142 bytes)
1752444503.049782: Client live command (20 bytes)
1752444503.066485: Device ping ACK (12 bytes)
```

**Connection establishment time**: ~189ms

#### Keepalive Pattern
Regular ping messages from device:
- Interval: ~1.9-2.8 seconds
- Size: 12 bytes consistently
- Response: Client acknowledges with 16-20 byte responses

### Message Structure Analysis

#### Ping Messages (Type 0x01)
```
Hex: 0001 0101 0000 8700 XXXX 0001 0000 0000 0000
```
- Header: `0001 0101`
- Status: `0000 8700` (device ready)
- Sequence: `XXXX` (increments: 0000, 0100, 0200, etc.)
- Footer: `0001 0000 0000 0000`

#### Discovery Response (Type 0x02)
```
Length: 142 bytes
Header: 0201 0101 0000 8700 0000 0001
Device Info: Contains "SICA R" (device model)
```

#### Live Commands (Type 0x09)
```
Header: 0901 0201 0000 0000
Command data varies by operation
Size: 20 bytes typically
```

### Response Codes
- `0x8700`: Device ready/operational status
- `0x0101`: Standard message flag
- `0x0001`: Message sequence/ID

### Timing Correlations

#### Command-Response Pairs
1. **Ping → ACK**: 10-17ms typical response time
2. **Discovery → Response**: 4-5ms response time
3. **Live Command → Status**: 16-20ms response time

#### Keepalive Intervals
- Shortest: 1.758s
- Longest: 2.814s  
- Average: ~2.2s
- No missed keepalives observed in capture

### Status Monitoring
Device maintains connection health through:
1. Regular ping broadcasts
2. Immediate ACK requirements
3. Status code embedding in all messages
4. Sequence number tracking

### Network Behavior
- Clean UDP communication
- No packet loss observed
- Consistent message timing
- Proper connection state management

## Recommendations
1. Implement 3-second timeout for keepalive detection
2. Use sequence numbers for message ordering
3. Monitor status code `0x8700` for device health
4. Handle discovery handshake during connection loss recovery
