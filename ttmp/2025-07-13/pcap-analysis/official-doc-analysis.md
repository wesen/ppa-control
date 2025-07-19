# Official PPA Protocol Documentation Analysis
*Cross-referenced with PCAP findings - 2025-07-13*

## Documented Message Types (from official Four Audio documentation)

### Confirmed Types
- **Type 0x00**: Ping - BasicHeader only, check device availability
- **Type 0x01**: LiveCmd - Set/query individual parameters (gain, EQ, etc.)
- **Type 0x02**: DeviceData - Query/set general device information  
- **Type 0x04**: PresetRecall - Recall presets or query active preset

### Status Codes (documented)
- **0x0001**: Response (success)
- **0x0002**: Command (changes device state)
- **0x0006**: Request (queries parameters)
- **0x0009**: Error
- **0x0041**: Wait (processing delay)

## Cross-Reference with PCAP Findings

### What We Found vs Documentation

#### ✅ **Matches Official Docs**
- Type 0 (Ping): ✅ Confirmed in both
- Type 2 (DeviceData): ✅ Confirmed in both
- Type 4 (PresetRecall): ✅ Confirmed in both
- Status 0x0001: ✅ Our "status code 0" likely maps to 0x0001 (little endian)

#### ❌ **Missing from Official Docs**
- **Type 1 (LiveCmd)**: ABSENT from our captures - No live parameter changes recorded
- **Type 3**: Unknown - NOT documented
- **Type 6**: Unknown - NOT documented  
- **Type 9**: Unknown - NOT documented
- **Type 10**: Unknown - NOT documented (90% of traffic!)

## Analysis of Unknown Types

### Type 10 (High-Frequency Streaming)
**Hypothesis**: Real-time audio/DSP data streaming
- **Evidence**: 299 occurrences, 9Hz frequency, 144-byte packets
- **Theory**: Undocumented streaming protocol for real-time audio monitoring
- **Payload Structure**: Likely contains multi-channel audio levels, timestamps
- **Not in Official Doc**: This suggests streaming was added after v2 protocol

### Type 6 (Configuration)  
**Hypothesis**: Extended configuration/setup commands
- **Evidence**: Appears during device initialization
- **Theory**: Advanced device configuration beyond basic DeviceData
- **Not in Official Doc**: Possibly proprietary extensions

### Type 3 & 9 (Rare Occurrences)
**Hypothesis**: Event notifications or error reporting
- **Evidence**: Sporadic appearance in connection logs
- **Theory**: Async notifications, diagnostics, or extended error handling
- **Not in Official Doc**: Likely internal/diagnostic messages

## Protocol Evolution Analysis

### Official Doc (2021) vs Current Implementation
The official documentation (v2, 2021-04-07) appears to describe a **command-response** protocol, but our captures reveal a **hybrid system**:

1. **Command Layer**: Types 0,2,4 (documented) 
2. **Streaming Layer**: Type 10 (undocumented, 90% of traffic)
3. **Extended Layer**: Types 3,6,9 (undocumented)

### Missing Type 1 (LiveCmd) Analysis
**Critical Finding**: No Type 1 (LiveCmd) messages in our captures
- **Possible Reasons**:
  - Commands were issued before capture started
  - Different client software bypasses LiveCmd
  - Type 10 streaming replaces individual parameter queries
  - Protocol evolution deprecated LiveCmd for streaming

## Key Insights

### 1. Documentation is Incomplete
The official doc covers ~10% of actual protocol usage. Type 10 streaming dominates but is completely undocumented.

### 2. Protocol Has Evolved
Evidence suggests significant evolution since 2021 v2 documentation:
- Addition of high-frequency streaming (Type 10)
- Extended configuration system (Type 6)  
- Enhanced diagnostic/event system (Types 3,9)

### 3. Two-Layer Architecture
- **Control Layer**: Traditional command-response (Types 0,2,4)
- **Data Layer**: Real-time streaming (Type 10)

## Next Steps for Investigation

1. **Decode Type 10 payload structure** - likely IEEE 754 floats for audio data
2. **Trigger Type 1 (LiveCmd)** - use official examples to generate this traffic
3. **Reverse engineer Types 3,6,9** - analyze payload patterns
4. **Compare with Type 1 examples** - implement official LiveCmd examples

## Impact on Protocol Implementation

Our implementation should support **both** documented and undocumented features:
- Official command-response protocol (documented)
- Real-time streaming protocol (undocumented but dominant)
- Extended configuration and diagnostics (undocumented)
