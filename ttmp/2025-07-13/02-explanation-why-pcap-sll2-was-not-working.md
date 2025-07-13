# PCAP Analysis Issue: Linux Cooked v2 Format and gopacket Compatibility

**Date:** July 13, 2025  
**Issue:** PPA protocol PCAP analysis not working with captured packets  
**Status:** RESOLVED  

## Problem Summary

The PCAP analysis tool (`cmd/pcap`) was unable to parse captured PPA protocol packets, showing no output despite valid network traffic being captured. The root cause was an incompatibility between the packet capture format (Linux cooked v2) and the gopacket library's parsing capabilities.

## Technical Background

### What is Linux Cooked Format?

Linux cooked format (also known as SLL - Linux "cooked" link layer) is a special packet capture format used by tcpdump/libpcap when:

1. **Capturing on "any" interface** (`-i any`)
2. **Capturing on multiple interfaces simultaneously**
3. **Capturing on certain virtual or non-Ethernet interfaces**

There are two versions:
- **SLL (v1)**: Original Linux cooked format
- **SLL2 (v2)**: Enhanced version with additional metadata

### Why Linux Cooked Format Exists

When capturing on "any" interface, tcpdump needs to handle packets from different interface types (Ethernet, WiFi, loopback, etc.) with potentially different frame formats. The Linux cooked format provides a **unified header structure** that abstracts away the underlying link-layer differences.

#### Standard Ethernet Frame vs Linux Cooked Frame

**Ethernet Frame:**
```
[Ethernet Header] [IP Header] [UDP Header] [PPA Payload]
    14 bytes         20 bytes    8 bytes     variable
```

**Linux Cooked v2 Frame:**
```
[SLL2 Header] [IP Header] [UDP Header] [PPA Payload]
   20 bytes     20 bytes    8 bytes     variable
```

## Investigation Process

### 1. Initial Symptoms

```bash
$ ./pcap --output-format jsonl captured-file.pcap
Opening captured-file.pcap
# No output despite 43 packets in file
```

```bash
$ tcpdump -r captured-file.pcap | wc -l
43  # Confirmed packets exist
```

### 2. Link Type Analysis

```bash
$ tcpdump -r captured-file.pcap -c 1 2>&1
reading from file captured-file.pcap, link-type LINUX_SLL2 (Linux cooked v2), snapshot length 262144
```

**Key Discovery:** The capture used Linux cooked v2 format, not standard Ethernet.

### 3. gopacket Behavior Investigation

Created test program to examine packet parsing:

```go
packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
for packet := range packetSource.Packets() {
    for _, layer := range packet.Layers() {
        fmt.Printf("Layer: %v\n", layer.LayerType())
    }
}
```

**Result:**
```
Link type: UnknownLinkType
Packet 1 layers: DecodeFailure
Packet 2 layers: DecodeFailure
```

**Root Cause Identified:** gopacket library couldn't decode Linux cooked v2 format, returning `DecodeFailure` instead of extracting IPv4/UDP layers.

### 4. PPA Tool Requirements Analysis

The `cmd/pcap/packet-handler.go` expects:

```go
func (ph *PacketHandler) handlePacket(packet gopacket.Packet) {
    ip4Layer := packet.Layer(layers.LayerTypeIPv4)
    if ip4Layer == nil {
        return  // Packet ignored
    }
    udpLayer := packet.Layer(layers.LayerTypeUDP)
    if udpLayer == nil {
        return  // Packet ignored
    }
    // Process PPA payload...
}
```

Since gopacket couldn't extract IPv4/UDP layers from the cooked format, all packets were silently ignored.

## Root Cause Analysis

### Why Our Scripts Generated Linux Cooked Format

Original capture scripts used:

```bash
sudo tcpdump -i any -w capture.pcap "host $DEVICE_IP and port $PORT"
```

The `-i any` flag forces Linux cooked format because:
- It captures from all interfaces simultaneously
- Different interfaces may have different link-layer formats
- tcpdump uses cooked format as a common denominator

### Why gopacket Failed

1. **Version Compatibility:** The gopacket version in use doesn't support Linux cooked v2 (SLL2)
2. **Layer Detection:** gopacket returned `UnknownLinkType` for the link layer
3. **Fallback Mechanism:** No fallback to attempt SLL2 parsing
4. **Silent Failure:** Packets were silently dropped rather than generating errors

## Solutions Implemented

### 1. Fix Capture Format (Primary Solution)

**Changed from:**
```bash
sudo tcpdump -i any -w capture.pcap "host $DEVICE_IP and port $PORT"
```

**Changed to:**
```bash
DEFAULT_INTERFACE=$(ip route | grep default | head -1 | awk '{print $5}')
sudo tcpdump -i "$DEFAULT_INTERFACE" -w capture.pcap "host $DEVICE_IP and port $PORT"
```

**Result:** Generates standard Ethernet format that gopacket can parse.

**Verification:**
```bash
$ tcpdump -r new-capture.pcap -c 1 2>&1
reading from file new-capture.pcap, link-type EN10MB (Ethernet), snapshot length 262144
```

### 2. Fix JSON Output Issue (Secondary Fix)

**Problem:** Debug messages contaminated JSON output:
```bash
$ ./pcap --output-format jsonl file.pcap | jq
Opening file.pcap
{"timestamp":"..."}  # jq failed to parse due to "Opening" line
```

**Solution:** Removed debug output for non-text formats:
```go
// Only print opening message for text format
if ph.outputFormat == "text" {
    fmt.Printf("Opening %s\n", fileName)
}
```

### 3. Updated All Capture Scripts

Modified scripts:
- `quick-capture.sh`
- `record-ppa-session.sh` 
- `record-specific-operations.sh`
- `test-protocol-implementation.sh`
- `protocol-fuzzer.sh`

All now use specific network interface instead of `-i any`.

## Results After Fix

### Successful Packet Analysis

```bash
$ ./pcap --output-format jsonl capture.pcap | jq -r '.header.MessageType' | sort | uniq -c
      2 0    # Ping messages
     37 10   # Unknown message type!
      4 9    # Unknown message type!
```

### Protocol Discovery

The fix revealed **previously unknown PPA message types**:
- **MessageType 9:** 4 packets (undocumented)
- **MessageType 10:** 37 packets (undocumented)
- **Status 0:** 37 packets (undocumented status code)

This suggests the PPA protocol has more message types than currently documented in `lib/protocol/ppa-protocol.go`.

## Technical Lessons Learned

### 1. Capture Interface Selection Matters

| Interface Flag | Link Type | gopacket Support | Use Case |
|---|---|---|---|
| `-i any` | Linux cooked v2 | ❌ No | Multi-interface capture |
| `-i eth0` | Ethernet | ✅ Yes | Single interface |
| `-i wlan0` | Ethernet | ✅ Yes | WiFi interface |

### 2. gopacket Limitations

- **Limited SLL2 support** in current version
- **Silent failures** when encountering unknown link types
- **No automatic fallback** mechanisms

### 3. PCAP Format Compatibility

Different tools handle various PCAP formats differently:
- **tcpdump:** Can read/display any format
- **Wireshark:** Can parse most formats including SLL2
- **gopacket:** Limited to supported link types

## Recommendations

### 1. For Future Captures

- **Always use specific network interface** instead of `-i any`
- **Verify link type** after capture: `tcpdump -r file.pcap -c 1 2>&1 | grep link-type`
- **Test with analysis tools** before lengthy capture sessions

### 2. For Tool Development

- **Add link type validation** in packet analysis tools
- **Implement graceful degradation** for unsupported formats
- **Provide clear error messages** instead of silent failures

### 3. For Protocol Analysis

- **Document discovered message types** (9 and 10)
- **Investigate unknown status codes**
- **Update protocol constants** in `lib/protocol/ppa-protocol.go`

## Code Changes Made

### Files Modified:

1. **`cmd/pcap/main.go`:** Removed debug output for JSON formats
2. **`cmd/pcap/packet-handler.go`:** Conditional debug output
3. **`ttmp/2025-07-13/pcap/quick-capture.sh`:** Use specific interface
4. **Other capture scripts:** Similar interface selection fixes

### Testing Verification:

```bash
# Before fix
$ ./pcap --output-format jsonl old-capture.pcap
Opening old-capture.pcap
# No output

# After fix  
$ ./pcap --output-format jsonl new-capture.pcap | head -2
{"timestamp":"17:52:50.177362","time_offset":"+0ms",...}
{"timestamp":"17:52:50.287711","time_offset":"+110ms",...}
```

## Conclusion

The PCAP analysis failure was caused by an incompatibility between Linux cooked v2 packet format (generated by `-i any`) and the gopacket library's parsing capabilities. The solution was to modify capture scripts to use specific network interfaces, generating standard Ethernet format packets that gopacket can properly decode.

This fix not only resolved the immediate analysis issue but also **enabled discovery of new PPA protocol message types**, significantly advancing our protocol reverse engineering efforts.

The incident highlights the importance of understanding the relationship between packet capture methods, file formats, and analysis tool capabilities in network protocol analysis workflows.
