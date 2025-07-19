# Oracle Protocol Analysis - PPA v3 Evolution
*Generated: 2025-07-13*

## Executive Summary

The Oracle's analysis reveals that the PPA protocol has **evolved significantly** since the 2021 v2 documentation into what we can call **"PPA v3"** - a hybrid command-response + streaming system with fragmented bulk data transfer capabilities.

## Key Revelations

### 1. Protocol Evolution (v2 → v3)
- **v2 (2021)**: Simple command-response with LiveCmd for individual parameters
- **v3 (2025)**: Hybrid architecture with bulk transfers and real-time streaming
- **LiveCmd (Type 1) deprecated** in favor of more efficient bulk operations

### 2. Message Type Classification

#### **Documented (v2 Spec)**
- **Type 0x00**: Ping/KeepAlive - unchanged
- **Type 0x01**: LiveCmd - **deprecated in current firmware**
- **Type 0x02**: DeviceData - enhanced with ACK functionality  
- **Type 0x04**: PresetRecall - unchanged

#### **New/Undocumented (v3 Extensions)**
- **Type 0x03**: "PresetDirectory" - delivers preset list (name + index)
- **Type 0x06**: "BulkParameterBlock" - complete device state dump/restore
- **Type 0x09**: "TransactionControl" - commit/undo operations
- **Type 0x0A**: "LiveStatusStream" - real-time channel metering (9Hz)

### 3. Status Code Evolution

The Oracle identified **expanded status codes** beyond the v2 spec:

```
0x0000 = Unsolicited push/stream (device→client)
0x0001 = ResponseServer (spec's 0x0001)  
0x0101 = ResponseClient (device ACK)
0x0102 = CommandClient (client → device)
0x0106 = RequestClient (client → device info)
0x1300 = Fragment START marker
0x2300 = Fragment DATA marker  
0x4100 = Fragment END marker
```

### 4. Fragmentation Protocol

**Critical Discovery**: Types 3 and 6 use a **fragmentation system** for large data:
- **START frame** (0x1300): 16 bytes, initiates transfer
- **DATA frames** (0x2300): 528 bytes (0x200 payload + 0x10 header)
- **END frame** (0x4100): 16 bytes, completes transfer

### 5. Type 10 (0x0A) Structure Decoded

The Oracle decoded the **LiveStatusStream** structure:

```
Offset 0x00: uint16 FlagsA (0104)
Offset 0x02: uint16 FlagsB (0602)  
Offset 0x04: uint16 FlagsC (0104)
Offset 0x06: uint16 FlagsD (0408)
Offset 0x08: uint64 Device-ticks (~0.1s step)
Offset 0x10: 12×float32 RMS/Peak pairs (64 bytes)
Offset 0x50: Additional channel data (zeros in capture)
```

## Protocol Architecture (v3)

### Three-Layer System
1. **Command Layer**: Traditional command-response (Types 0,2,4,9)
2. **Bulk Transfer Layer**: Fragmented data operations (Types 3,6)  
3. **Streaming Layer**: Real-time unsolicited data (Type 10)

### Data Flow Patterns
- **Device Discovery**: Ping → DeviceData → PresetDirectory → BulkParameterBlock
- **Parameter Changes**: BulkParameterBlock fragments → TransactionControl commit
- **Real-time Monitoring**: Continuous LiveStatusStream (9Hz)

## Implementation Impact

### Why LiveCmd (Type 1) Disappeared
Modern GUI software uses **BulkParameterBlock** instead of hundreds of individual LiveCmd requests for efficiency.

### Fragmentation Necessity  
Device configuration data exceeds UDP packet limits, requiring the START-DATA-END fragmentation protocol.

### Stream vs Poll Architecture
v3 shifted from polling individual parameters to **push-based streaming** for real-time data.

## Next Steps for Protocol Implementation

### 1. Immediate Actions
- **Update decoder** to handle fragmentation (0x1300/0x2300/0x4100)
- **Implement Type 10 parser** for real-time metering
- **Add Type 9 transaction support** for atomic operations

### 2. Testing Strategy
- **Capture older software** that still uses Type 1 (LiveCmd)
- **Trigger preset save operations** to observe Type 5
- **Map channel indices** in Type 10 during live audio changes

### 3. Protocol Documentation
- **Document v3 extensions** as unofficial spec addendum
- **Create TLV dictionary** for Type 6 sub-structures
- **Build compatibility matrix** (v2 vs v3 support)

## Critical Success Factors

1. **Backward Compatibility**: Must support both v2 command-response and v3 streaming
2. **Fragmentation Handling**: Essential for bulk operations
3. **Stream Processing**: Real-time Type 10 data for responsive UI
4. **Transaction Semantics**: Proper Type 9 commit/rollback support

This analysis provides the foundation for implementing a **complete PPA v3 client** that can handle both legacy v2 devices and modern streaming-capable devices.
