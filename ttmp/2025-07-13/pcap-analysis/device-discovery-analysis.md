# PPA Device Discovery Analysis

Generated: analyze.py
Analysis Date: 2025-07-13

## Executive Summary

- **Total Packets**: 12
- **Parsed PPA Packets**: 12
- **Unique Message Types**: 5
- **Client→Device**: 0
- **Device→Client**: 12

## Message Type Analysis

| Type | Name | Count | Percentage |
|------|------|-------|------------|
| 0 | Ping | 3 | 25.0% |
| 9 | Unknown | 3 | 25.0% |
| 3 | Unknown | 3 | 25.0% |
| 2 | DeviceData | 2 | 16.7% |
| 4 | PresetRecall | 1 | 8.3% |

## Status Type Analysis

| Status | Name | Count |
|--------|------|-------|
| 0x0101 (ResponseClient) | 9 |
| 0x1300 (Unknown) | 1 |
| 0x4100 (Unknown) | 1 |
| 0x2300 (Unknown) | 1 |

## Packet Size Distribution

| Size (bytes) | Count |
|--------------|-------|
| 12 | 6 |
| 16 | 3 |
| 124 | 1 |
| 142 | 2 |

## Unknown/Undocumented Message Types

- **Type 9** at 18:05:53.816516
  - Payload: `090101010000870001000001000000000000`
- **Type 3** at 18:05:53.820306
  - Payload: `0301130000008700030000006c0000000000`
- **Type 3** at 18:05:53.820306
  - Payload: `030141000000870003000001000064000000`
- **Type 3** at 18:05:53.823563
  - Payload: `030123000000870003000000000001000000420200000053`
- **Type 9** at 18:05:56.801504
  - Payload: `090101010000870006000001000000000000`
- **Type 9** at 18:05:57.777684
  - Payload: `090101010000870007000001000000000000`

## Protocol Sequence Patterns

### Timing Analysis
- Total message exchanges: 12
- Average interval analysis shows regular patterns

### Message Flow Pattern
```
18:05:53.710553 → 0 (Ping) (ResponseClient) Seq:0 Len:12
18:05:53.794107 → 2 (DeviceData) (ResponseClient) Seq:0 Len:142
18:05:53.816516 → 9 (Unknown) (ResponseClient) Seq:256 Len:12
18:05:53.820306 → 3 (Unknown) (Unknown) Seq:768 Len:16
18:05:53.820306 → 3 (Unknown) (Unknown) Seq:768 Len:16
18:05:53.823563 → 3 (Unknown) (Unknown) Seq:768 Len:124
18:05:53.828995 → 2 (DeviceData) (ResponseClient) Seq:1024 Len:142
18:05:53.831816 → 4 (PresetRecall) (ResponseClient) Seq:1280 Len:16
18:05:55.685153 → 0 (Ping) (ResponseClient) Seq:256 Len:12
18:05:56.801504 → 9 (Unknown) (ResponseClient) Seq:1536 Len:12
18:05:57.676654 → 0 (Ping) (ResponseClient) Seq:512 Len:12
18:05:57.777684 → 9 (Unknown) (ResponseClient) Seq:1792 Len:12
```

## Key Findings

- **Ping/Keepalive Activity**: 3 ping messages detected, indicating active connection monitoring
- **Device Discovery**: 2 device data exchanges, likely initial device enumeration
- **Preset Operations**: 1 preset-related messages (recall/save)
- **Unknown Messages**: 6 undocumented message types require investigation

## Anomalies and Interesting Behaviors

- **Unusual Packet Sizes**: Sizes [124] differ from standard PPA message sizes
- **Status Imbalance**: Unequal request/response status distribution
