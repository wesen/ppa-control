# Preset Recall Analysis - PPA Control Protocol

## Overview
Analysis of preset recall operations based on PCAP captures from 2025-07-13.

## Capture Files Analyzed
- `ppa-ops-20250713-181344-07-preset-recall.pcap`
- `data/sniff preset 5-6.pcapng` (historical)

## Key Findings

### Preset Recall Message Structure
```
MessageType: MessageTypePresetRecall (4)
ProtocolId: 1
Status: [varies by phase]
DeviceUniqueId: [cc069300 for device responses]
SequenceNumber: [incremental]
ComponentId: [0 for device, ff for broadcast]
Reserved: 1
PresetRecall.CrtFlags: 0
PresetRecall.OptFlags: [varies]
PresetRecall.PresetId: [target preset]
```

### Preset Recall Sequence (2025-07-13 Capture)

#### Phase 1: Preset 4 Recall (14:58:09)
1. **Command Initiation**
   - Client → Device: `MessageTypePresetRecall(4)`, Status: `StatusCommandClient(2)`, PresetId: `4`
   - Device response: Status: `StatusWaitServer(41)`, PresetId: `2c` (44 decimal - processing indicator)

2. **Processing Phase**
   - Device takes 664ms to process the preset recall
   - Device → Client: Status: `StatusResponseClient(1)` (completion)

3. **Status Query**
   - Client → Device: Status: `StatusRequestServer(6)`, PresetId: `0` (query current state)
   - Device response: Status: `StatusResponseClient(1)`, OptFlags: `4`, PresetId: `4`

#### Phase 2: Preset 0 Recall (14:58:11)
1. **Command Initiation**
   - Client → Device: PresetId: `0`, OptFlags: `2`
   - Device response: Status: `StatusWaitServer(41)`, PresetId: `2c`

2. **Processing Phase**
   - Device takes 840ms to process (longer than preset 4)
   - Device → Client: Status: `StatusResponseClient(1)`

3. **Status Query**
   - Client query: PresetId: `0`
   - Device response: OptFlags: `6`, PresetId: `0`

### Protocol Patterns

#### Status Codes During Preset Recall
- `StatusCommandClient(2)`: Initial preset recall command
- `StatusWaitServer(41)`: Device processing acknowledgment
- `StatusResponseClient(1)`: Operation completion
- `StatusRequestServer(6)`: Status query
- `StatusErrorServer(9)`: Error conditions (observed during ping timeouts)

#### OptFlags Interpretation
- `0`: Base state
- `2`: Command with options
- `4`: Preset 4 active indicator
- `6`: Preset 0 active indicator with additional flags

#### Processing Times
- **Preset 4**: 664ms processing time
- **Preset 0**: 840ms processing time
- Indicates variable complexity or initialization requirements per preset

### Device Behavior
1. **Acknowledgment Pattern**: Device immediately acknowledges with `StatusWaitServer(41)`
2. **Processing Indicator**: Uses PresetId `2c` (44) during processing phase
3. **Completion Signal**: Returns to `StatusResponseClient(1)` when done
4. **State Persistence**: Preset state maintained and queryable after recall

### Error Handling
- No errors observed during preset recall operations
- Ping timeouts (`StatusErrorServer(9)`) occur during network issues but don't affect preset operations
- Device maintains operation integrity despite communication interruptions

## Historical Comparison Notes
The 2025-07-13 captures show more sophisticated preset handling compared to earlier captures:
- More detailed status reporting
- Consistent timing patterns
- Robust error recovery mechanisms

## Recommendations
1. **Timing Considerations**: Allow at least 1 second for preset recall operations
2. **Status Monitoring**: Poll device status after preset commands using StatusRequestServer(6)
3. **Error Recovery**: Implement retry logic for network-level errors
4. **State Validation**: Verify preset activation through OptFlags and PresetId fields
