# EQ Adjustments Analysis

**File**: ppa-ops-20250713-181344-06-eq-adjustments.pcap  
**Analysis Date**: 2025-07-13  
**Total Packets**: 1073 (estimated from file size)

## Overview

Analysis of PPA EQ parameter adjustment commands captured during user interface testing. This PCAP contains bidirectional UDP traffic showing clear command-response patterns between client and PPA device (192.168.0.200:5001).

## Traffic Pattern Analysis

### Packet Direction Distribution
- **Device → Client**: 144-byte periodic status updates (~100ms intervals)
- **Client → Device**: Command packets (20, 28, 44 bytes)
- **Device → Client**: 12-byte acknowledgment responses

### Command-Response Sequence
Clear request-acknowledgment pattern observed:
1. Client sends command packet
2. Device responds with 12-byte ACK within 50-200ms
3. Device continues periodic 144-byte status updates

## EQ Command Structure Analysis

### Command Types Identified

#### 1. EQ Parameter Set Commands (44 bytes)
```
Pattern: 0101 0201 0000 0000 XXXX fe01 0001 YYYY 0305 0100
Structure:
- Protocol header: 0101 0201 0000 0000
- Message ID: XXXX (increments: 26→27→28...)
- Command type: fe01 (parameter set)
- Parameter count: 0001 (single parameter)
- Parameter type: YYYY (varies)
- Parameter class: 0305 0100 (EQ-related)
- Parameter value: 32-bit data
- Additional parameters for dual adjustments
```

#### 2. Single Parameter Commands (28 bytes)
```
Pattern: 0101 0201 0000 0000 XXXX fe01 0000 YYYY 030X 0100
Structure:
- Same header structure
- Parameter count: 0000 (query/single adjustment)
- Class variations: 0305, 0304, 0307, 0300, 0303
```

#### 3. Commit/Apply Commands (20 bytes)
```
Pattern: 0901 0201 0000 0000 XXXX ff01 0000 581b 0a00 0000
Structure:
- Protocol type: 0901 (commit operation)
- Command: ff01 (apply/commit)
- Fixed suffix: 581b 0a00 0000
```

### EQ Parameter Encoding

#### Parameter Values Observed
From the hex analysis of 44-byte packets:

| Command | Parameter Type | Value (hex) | Value (decimal) | Description |
|---------|---------------|-------------|-----------------|-------------|
| 27 | 0600 | 9407 0000 | 1940 | EQ Band adjustment |
| 28 | 0600 | 5207 0000 | 1874 | EQ Band adjustment |
| 29 | 0600 | 8805 0000 | 1416 | EQ Band adjustment |
| 2A | 0600 | 1205 0000 | 1298 | EQ Band adjustment |
| 2B | 0600 | 3406 0000 | 1588 | EQ Band adjustment |

#### Parameter Class Analysis
- **0305**: Primary EQ parameters (gain/frequency)
- **0304**: EQ band selection
- **0303**: EQ curve parameters  
- **0300**: EQ enable/bypass
- **0307**: EQ preset selection

### EQ Value Encoding

#### Frequency Values (16-bit, little-endian)
- Range appears to be 0-65535 mapping to EQ frequency spectrum
- Examples: 1940 (07 94), 1874 (52 07), 1416 (05 88)

#### Gain Values (typically follow frequency)
- Appear as secondary parameters in 44-byte commands
- Examples: 0309 (777), 0370 (880), 0394 (916), 036F (879)

## Device Response Analysis

### Acknowledgment Pattern
All client commands receive 12-byte responses:
```
Pattern: 0101 0101 0000 8700 XXXX 0001 0000 0000 0000
Structure:
- Response header: 0101 0101 0000 8700
- Message ID echo: XXXX (matches command)
- Status: 0001 (success)
- Padding: 0000 0000 0000
```

### Status Message Changes
Device data messages show parameter changes in real-time:
- Message ID increments from 8700 f700 → 8700 fd00
- Status field changes indicate parameter application
- Timestamp updates reflect processing completion

## EQ Operation Sequence

### Typical Command Flow
1. **Band Selection**: 28-byte command (class 0304)
2. **Frequency Adjustment**: 44-byte command (dual parameter)
3. **Gain Adjustment**: 44-byte command (dual parameter)  
4. **Apply Changes**: 20-byte commit command
5. **Device ACK**: 12-byte response for each step

### Multi-Parameter Updates
Some 44-byte commands contain dual parameters:
- Primary: EQ frequency/band (offset 0x1C-0x1F)
- Secondary: EQ gain/slope (offset 0x28-0x2B)

## Parameter Value Ranges

### Observed Frequency Range
- Minimum: ~1000 (03 E8)
- Maximum: ~7000 (1B 58)  
- Resolution: Appears to be 1 Hz precision

### Observed Gain Range
- Center: ~800-900 (neutral/0dB equivalent)
- Range: 300-1400 (approximate ±6dB equivalent)
- Resolution: Single digit precision

## Command Acknowledgment Timing

### Response Latency
- Typical ACK delay: 50-150ms
- Commit operation delay: 30-80ms
- Parameter application visible in next device data packet

### Command Rate Limiting
- Commands spaced 100-200ms apart
- No rejected commands observed
- Device handles rapid parameter changes gracefully

## Technical Implementation Notes

### Message ID Sequencing
- Client uses incremental message IDs (26, 27, 28...)
- Device echoes message ID in ACK responses
- Enables command correlation and timeout detection

### Parameter Persistence
- Changes reflected immediately in device status
- Parameters survive across multiple adjustments
- No explicit save command observed (auto-persist)

### Error Handling
- All commands in capture were successful (status 0001)
- No error response patterns observed
- Device appears robust to parameter adjustments

## Protocol Insights

1. **Real-time Updates**: EQ changes immediately visible in device status
2. **Atomic Operations**: Individual parameter changes are discrete
3. **Stateful Protocol**: Device maintains EQ configuration
4. **Reliable Transport**: UDP with application-level ACKs
5. **Structured Parameters**: Clear type/class/value encoding

## Recommendations for Implementation

1. **Command Batching**: Group related EQ adjustments
2. **Response Correlation**: Use message IDs for command tracking
3. **Parameter Validation**: Respect observed value ranges
4. **Rate Limiting**: Space commands 100ms apart minimum
5. **Status Monitoring**: Watch device data for change confirmation
