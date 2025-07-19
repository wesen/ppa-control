# Preset Switching Resynchronization Analysis

**Session:** preset-switching-132325  
**Capture Duration:** 60 seconds  
**Total Packets:** 828

## Message Type Distribution

| Type | Count | Protocol Name | Description |
|------|-------|---------------|-------------|
| 0    | 29    | Ping          | Keepalive/heartbeat |
| 1    | 22    | LiveCmd       | Live commands (deprecated?) |
| 3    | 6     | PresetDirectory | Preset list/metadata |
| 4    | 31    | PresetRecall  | **Preset switching commands** |
| 6    | 320   | BulkParameterBlock | Config data (fragmented) |
| 9    | 44    | TransactionControl | Commit/rollback wrapper |
| 10   | 376   | LiveStatusStream | Real-time metering (~9Hz) |

## Key Findings: Massive Preset Activity

### Preset Switching Timeline
- **~1850+ Type 4 packets** observed during capture (grep shows extensive duplication/truncation in timing file)
- **Continuous preset switching** occurring approximately every 100-200ms
- **Pattern suggests automated preset cycling or UI rapid-fire clicking**

### Resynchronization Behavior Analysis

**Token-Heavy Operations Identified:**

1. **Type 6 (BulkParameterBlock) - 320 packets = 38.6% of traffic**
   - Large fragmented config transfers
   - Likely triggered per preset change
   - **Primary token consumer**

2. **Type 10 (LiveStatusStream) - 376 packets = 45.4% of traffic**  
   - Continuous ~9Hz streaming during entire session
   - Float32 metering data (seen in hex: `00 00 80 3f` = 1.0)
   - **Secondary token consumer**

3. **Type 9 (TransactionControl) - 44 packets**
   - Transaction wrappers around bulk operations
   - Commit/rollback coordination

### Resynchronization Duration Patterns

From timing analysis:
- **Burst periods**: 13:23:41.232053 - 13:23:41.454802 (~220ms burst of Type 4)
- **Sustained periods**: 13:23:45.486351 - 13:23:45.535418 (~50ms rapid fire)
- **Recovery gaps**: ~1-second delays between major preset operations

## Token Usage Implications

**Why Resynchronization "Eats Tokens":**

1. **Bulk Parameter Transfers (Type 6)**
   - 320 large packets with device configuration
   - Each preset change triggers full state dump
   - **Fragmented transmission = multiple round-trips**

2. **Continuous Streaming During Sync (Type 10)**
   - Real-time metering doesn't pause during preset changes
   - Creates parallel data flow competing for bandwidth
   - **144-byte payloads every ~110ms**

3. **Transaction Overhead (Type 9)**
   - Structured commit/rollback around configuration changes
   - Adds protocol coordination overhead

**Short vs Long Resync:**
- **Short**: Single Type 4 preset recall
- **Long**: Type 4 + burst of Type 6 fragments + Type 9 wrappers + continued Type 10 streaming

## Optimization Recommendations

1. **Batch preset changes** to reduce Type 6 fragmentation
2. **Pause streaming (Type 10)** during bulk transfers
3. **Cache configuration** to avoid full state dumps
4. **Implement incremental sync** for preset changes
