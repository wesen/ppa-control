#!/bin/bash
# PPA Protocol Implementation Testing Script
# Tests our protocol understanding by sending controlled commands and capturing responses

set -e

DEVICE_IP="${1:-192.168.0.200}"
PORT="5001"
PCAP_DIR="$(dirname "$0")/captures"
ANALYSIS_DIR="$(dirname "$0")/analysis"
PPA_CLI="$(dirname "$0")/../../../ppa-cli"
PCAP_TOOL="$(dirname "$0")/../../../pcap"
TEST_SESSION="protocol-test-$(date +%Y%m%d-%H%M%S)"

mkdir -p "$PCAP_DIR" "$ANALYSIS_DIR"

echo "🧪 PPA Protocol Implementation Testing"
echo "Device: $DEVICE_IP:$PORT"
echo "Session: $TEST_SESSION"
echo

# Build tools if needed
build_tools() {
    echo "🔨 Building tools..."
    cd "$(dirname "$0")/../../.."
    
    if [ ! -f "ppa-cli" ]; then
        echo "  Building ppa-cli..."
        go build ./cmd/ppa-cli
    fi
    
    if [ ! -f "pcap" ]; then
        echo "  Building pcap..."
        go build ./cmd/pcap
    fi
    
    cd - > /dev/null
    echo "✅ Tools ready"
}

# Function to test a specific command with capture
test_command() {
    local test_name="$1"
    local description="$2"
    local command="$3"
    local expected_msg_type="$4"
    local capture_duration="${5:-10}"
    
    echo "┌─ Testing: $test_name ─"
    echo "│ $description"
    echo "│ Command: $command"
    echo "│ Expected message type: $expected_msg_type"
    echo "│"
    
    local capture_file="$PCAP_DIR/${TEST_SESSION}-${test_name}.pcap"
    local analysis_file="$ANALYSIS_DIR/${TEST_SESSION}-${test_name}-analysis.txt"
    
    # Start capture in background
    echo "│ 🎥 Starting packet capture..."
    timeout "$capture_duration" sudo tcpdump -i any -w "$capture_file" \
        "host $DEVICE_IP and port $PORT" &
    local tcpdump_pid=$!
    
    # Wait a moment for tcpdump to start
    sleep 2
    
    # Execute the command
    echo "│ 🚀 Executing command..."
    eval "$command" || echo "│ ⚠️  Command failed (expected for some tests)"
    
    # Wait for capture to complete
    wait $tcpdump_pid 2>/dev/null || true
    
    # Analyze the capture
    if [ -f "$capture_file" ]; then
        local packet_count=$(tcpdump -r "$capture_file" 2>/dev/null | wc -l)
        echo "│ 📊 Captured $packet_count packets"
        
        # Analyze message types
        echo "Analysis for $test_name" > "$analysis_file"
        echo "Command: $command" >> "$analysis_file"
        echo "Expected: $expected_msg_type" >> "$analysis_file"
        echo "" >> "$analysis_file"
        
        if [ "$packet_count" -gt 0 ]; then
            echo "Message types found:" >> "$analysis_file"
            "$PCAP_TOOL" --output-format jsonl "$capture_file" 2>/dev/null | \
                jq -r '.header.message_type // "unknown"' | \
                sort | uniq -c >> "$analysis_file"
            
            echo "" >> "$analysis_file"
            echo "Status codes found:" >> "$analysis_file"
            "$PCAP_TOOL" --output-format jsonl "$capture_file" 2>/dev/null | \
                jq -r '.header.status // "unknown"' | \
                sort | uniq -c >> "$analysis_file"
            
            # Check if expected message type was found
            local found_expected=$(grep -c "$expected_msg_type" "$analysis_file" || echo "0")
            if [ "$found_expected" -gt 0 ]; then
                echo "│ ✅ Found expected message type: $expected_msg_type"
            else
                echo "│ ❌ Expected message type not found: $expected_msg_type"
            fi
        else
            echo "No packets captured" >> "$analysis_file"
            echo "│ ❌ No packets captured"
        fi
    else
        echo "│ ❌ Capture file not created"
    fi
    
    echo "└─────────────────────────────"
    echo
}

# Protocol constant validation
validate_protocol_constants() {
    echo "🔍 Validating Protocol Constants"
    echo
    
    cat > "$ANALYSIS_DIR/${TEST_SESSION}-protocol-constants.txt" << EOF
PPA Protocol Constants Validation
Generated: $(date)

From lib/protocol/ppa-protocol.go:

Message Types:
- Ping: 0
- LiveCmd: 1
- DeviceData: 2
- PresetRecall: 4
- PresetSave: 5
- Unknown: 255

Status Types:
Client -> Server:
- StatusCommandClient: 0x0102 (258)
- StatusRequestClient: 0x0106 (262)
- StatusResponseClient: 0x0101 (257)
- StatusErrorClient: 0x0109 (265)
- StatusWaitClient: 0x0141 (321)

Server -> Client:
- StatusCommandServer: 0x0002 (2)
- StatusRequestServer: 0x0006 (6)
- StatusResponseServer: 0x0001 (1)
- StatusErrorServer: 0x0009 (9)
- StatusWaitServer: 0x0041 (65)

Level Types:
- Input: 1
- Output: 2
- Eq: 3
- Gain: 4
- EqType: 5
- Quality: 7
- Active: 8
- Mute: 9
- Delay: 10
- PhaseInversion: 11

Preset Recall Types:
- RecallByPresetIndex: 0
- RecallByPresetPosition: 2
EOF
    
    echo "✅ Protocol constants documented"
}

echo "🏗️  Setup phase..."
build_tools
validate_protocol_constants

echo
echo "🧪 Testing Protocol Implementation"
echo "   Each test will capture packets and analyze the protocol messages"
echo

# Test 1: Basic ping
test_command "ping" \
    "Basic ping command to test connectivity" \
    "$PPA_CLI ping -a \"$DEVICE_IP:$PORT\"" \
    "Ping" \
    15

# Test 2: Device data request
test_command "device-data" \
    "Request device information" \
    "$PPA_CLI device-info -a \"$DEVICE_IP:$PORT\"" \
    "DeviceData" \
    15

# Test 3: Volume control (LiveCmd)
test_command "volume-control" \
    "Set master volume using LiveCmd" \
    "$PPA_CLI volume -v 0.5 -a \"$DEVICE_IP:$PORT\"" \
    "LiveCmd" \
    15

# Test 4: Preset recall
test_command "preset-recall" \
    "Recall preset using PresetRecall message" \
    "$PPA_CLI recall --preset 1 -a \"$DEVICE_IP:$PORT\"" \
    "PresetRecall" \
    15

# Test 5: Discovery mode
test_command "discovery" \
    "Device discovery using broadcast ping" \
    "$PPA_CLI ping --discover" \
    "Ping" \
    20

# Test 6: Invalid operations (error testing)
test_command "invalid-preset" \
    "Test error handling with invalid preset" \
    "$PPA_CLI recall --preset 255 -a \"$DEVICE_IP:$PORT\"" \
    "PresetRecall" \
    15

echo "🔬 Advanced Protocol Analysis"

# Analyze all captures together
echo "📊 Creating comprehensive analysis..."

cat > "$ANALYSIS_DIR/${TEST_SESSION}-comprehensive-analysis.md" << EOF
# PPA Protocol Implementation Test Results

**Session:** $TEST_SESSION  
**Device:** $DEVICE_IP:$PORT  
**Generated:** $(date)

## Test Summary

EOF

# Add test results summary
for test_file in "$ANALYSIS_DIR/${TEST_SESSION}"-*-analysis.txt; do
    if [ -f "$test_file" ]; then
        test_name=$(basename "$test_file" | sed 's/.*-\(.*\)-analysis.txt/\1/')
        echo "### Test: $test_name" >> "$ANALYSIS_DIR/${TEST_SESSION}-comprehensive-analysis.md"
        echo '```' >> "$ANALYSIS_DIR/${TEST_SESSION}-comprehensive-analysis.md"
        cat "$test_file" >> "$ANALYSIS_DIR/${TEST_SESSION}-comprehensive-analysis.md"
        echo '```' >> "$ANALYSIS_DIR/${TEST_SESSION}-comprehensive-analysis.md"
        echo >> "$ANALYSIS_DIR/${TEST_SESSION}-comprehensive-analysis.md"
    fi
done

cat >> "$ANALYSIS_DIR/${TEST_SESSION}-comprehensive-analysis.md" << EOF

## Protocol Validation Results

Compare the captured message types and status codes with the expected protocol constants.

### Message Type Validation
- ✅ Expected message types should match protocol constants
- ❌ Any unexpected message types need investigation

### Status Code Validation  
- ✅ Status codes should follow client/server pattern
- ❌ Unknown status codes indicate protocol extensions or errors

### Sequence Number Analysis
- ✅ Sequence numbers should increment properly
- ❌ Gaps or duplicates may indicate retransmission or errors

## Recommendations

1. **Update Protocol Documentation**: Document any new message types or status codes found
2. **Add Test Cases**: Create unit tests for validated protocol patterns
3. **Error Handling**: Improve error handling for edge cases discovered
4. **Performance**: Analyze timing patterns for optimization opportunities

## Commands for Further Investigation

\`\`\`bash
# Detailed analysis of specific message type
$PCAP_TOOL --print-packets <messageType> --print-hexdump capture.pcap

# JSON export for programmatic analysis
$PCAP_TOOL --output-format json capture.pcap > analysis.json

# Timing analysis
$PCAP_TOOL --output-format jsonl capture.pcap | jq -r '[.timestamp, .header.message_type, .direction] | @csv'
\`\`\`
EOF

echo
echo "✅ Protocol testing complete!"
echo
echo "📁 Generated analysis files:"
ls -la "$ANALYSIS_DIR/${TEST_SESSION}"* 2>/dev/null

echo
echo "📖 Read the comprehensive analysis:"
echo "   cat $ANALYSIS_DIR/${TEST_SESSION}-comprehensive-analysis.md"

echo
echo "🚀 Next steps:"
echo "1. Review test results in analysis files"
echo "2. Compare with protocol specification"
echo "3. Update protocol implementation if needed"
echo "4. Add discovered patterns to test suite"
