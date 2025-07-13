#!/bin/bash
# PPA Protocol Fuzzing and Edge Case Testing Script
# Tests protocol robustness and discovers edge cases

set -e

DEVICE_IP="${1:-192.168.0.200}"
PORT="5001"
PCAP_DIR="$(dirname "$0")/captures"
ANALYSIS_DIR="$(dirname "$0")/analysis"
PPA_CLI="$(dirname "$0")/../../../ppa-cli"
PCAP_TOOL="$(dirname "$0")/../../../pcap"
FUZZ_SESSION="fuzz-test-$(date +%Y%m%d-%H%M%S)"

mkdir -p "$PCAP_DIR" "$ANALYSIS_DIR"

echo "🔥 PPA Protocol Fuzzing and Edge Case Testing"
echo "Device: $DEVICE_IP:$PORT"
echo "Session: $FUZZ_SESSION"
echo "⚠️  This script tests edge cases and may cause unexpected device behavior!"
echo

read -p "Continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

# Build tools if needed
if [ ! -f "$PPA_CLI" ] || [ ! -f "$PCAP_TOOL" ]; then
    echo "🔨 Building tools..."
    cd "$(dirname "$0")/../../.."
    go build ./cmd/ppa-cli ./cmd/pcap
    cd - > /dev/null
fi

# Function to run fuzz test with capture
fuzz_test() {
    local test_name="$1"
    local description="$2"
    local command="$3"
    local duration="${4:-15}"
    
    echo "┌─ Fuzzing: $test_name ─"
    echo "│ $description"
    echo "│ Command: $command"
    echo "│"
    
    local capture_file="$PCAP_DIR/${FUZZ_SESSION}-${test_name}.pcap"
    local analysis_file="$ANALYSIS_DIR/${FUZZ_SESSION}-${test_name}.txt"
    
    # Start capture
    timeout "$duration" sudo tcpdump -i any -w "$capture_file" \
        "host $DEVICE_IP and port $PORT" &
    local tcpdump_pid=$!
    
    sleep 2
    
    # Execute test command (expect it may fail)
    echo "│ 🚀 Executing..."
    eval "$command" 2>&1 || echo "│ ⚠️  Command failed (expected for some fuzz tests)"
    
    wait $tcpdump_pid 2>/dev/null || true
    
    # Analyze results
    if [ -f "$capture_file" ]; then
        local packet_count=$(tcpdump -r "$capture_file" 2>/dev/null | wc -l)
        echo "│ 📊 Captured $packet_count packets"
        
        # Quick analysis
        {
            echo "Fuzz test: $test_name"
            echo "Description: $description"
            echo "Command: $command"
            echo "Packets captured: $packet_count"
            echo ""
            
            if [ "$packet_count" -gt 0 ]; then
                echo "Message types:"
                "$PCAP_TOOL" --output-format jsonl "$capture_file" 2>/dev/null | \
                    jq -r '.header.message_type // "unknown"' | sort | uniq -c
                
                echo ""
                echo "Status codes:"
                "$PCAP_TOOL" --output-format jsonl "$capture_file" 2>/dev/null | \
                    jq -r '.header.status // "unknown"' | sort | uniq -c
                
                # Look for error status codes
                local error_count=$(grep -c "Error" "$analysis_file" 2>/dev/null || echo "0")
                if [ "$error_count" -gt 0 ]; then
                    echo ""
                    echo "⚠️  Error responses detected!"
                fi
            else
                echo "No protocol responses captured"
            fi
        } > "$analysis_file"
        
        # Check for errors in response
        local has_errors=$("$PCAP_TOOL" --output-format jsonl "$capture_file" 2>/dev/null | \
            jq -r '.header.status' | grep -c "Error" || echo "0")
        
        if [ "$has_errors" -gt 0 ]; then
            echo "│ ⚠️  Detected $has_errors error responses"
        else
            echo "│ ✅ No error responses detected"
        fi
    else
        echo "│ ❌ No capture file created"
    fi
    
    echo "└─────────────────────────────"
    echo
}

echo "🧪 Running edge case and fuzz tests..."

# Edge Case 1: Invalid preset numbers
fuzz_test "invalid-preset-high" \
    "Test with very high preset number" \
    "$PPA_CLI recall --preset 255 -a \"$DEVICE_IP:$PORT\""

fuzz_test "invalid-preset-zero" \
    "Test with preset number 0 (may be invalid)" \
    "$PPA_CLI recall --preset 0 -a \"$DEVICE_IP:$PORT\""

# Edge Case 2: Extreme volume values
fuzz_test "volume-extreme-high" \
    "Test with maximum volume value" \
    "$PPA_CLI volume -v 1.0 -a \"$DEVICE_IP:$PORT\""

fuzz_test "volume-extreme-low" \
    "Test with minimum volume value" \
    "$PPA_CLI volume -v 0.0 -a \"$DEVICE_IP:$PORT\""

fuzz_test "volume-negative" \
    "Test with negative volume (should be rejected)" \
    "$PPA_CLI volume -v -1.0 -a \"$DEVICE_IP:$PORT\""

fuzz_test "volume-over-max" \
    "Test with volume over maximum" \
    "$PPA_CLI volume -v 2.0 -a \"$DEVICE_IP:$PORT\""

# Edge Case 3: Rapid fire commands
fuzz_test "rapid-ping" \
    "Rapid ping commands to test sequence handling" \
    "for i in {1..10}; do $PPA_CLI ping -a \"$DEVICE_IP:$PORT\" & done; wait" \
    20

fuzz_test "rapid-volume" \
    "Rapid volume changes" \
    "for i in {1..5}; do $PPA_CLI volume -v 0.\$i -a \"$DEVICE_IP:$PORT\" & done; wait" \
    20

# Edge Case 4: Connection stress
fuzz_test "connection-spam" \
    "Multiple simultaneous connections" \
    "for i in {1..5}; do timeout 5 $PPA_CLI ping -a \"$DEVICE_IP:$PORT\" --loop & done; sleep 10; pkill -f ppa-cli" \
    25

# Edge Case 5: Invalid device addresses
fuzz_test "invalid-ip" \
    "Connect to non-existent IP" \
    "timeout 10 $PPA_CLI ping -a \"192.168.255.254:$PORT\"" \
    15

fuzz_test "invalid-port" \
    "Connect to wrong port" \
    "timeout 10 $PPA_CLI ping -a \"$DEVICE_IP:9999\"" \
    15

# Edge Case 6: Discovery edge cases
fuzz_test "discovery-spam" \
    "Rapid discovery requests" \
    "for i in {1..3}; do timeout 5 $PPA_CLI ping --discover & done; wait" \
    20

# Protocol boundary testing
fuzz_test "long-session" \
    "Extended session to test timeouts" \
    "timeout 60 $PPA_CLI ping -a \"$DEVICE_IP:$PORT\" --loop" \
    65

echo "🔍 Analyzing fuzz test results..."

# Create comprehensive fuzz analysis report
cat > "$ANALYSIS_DIR/${FUZZ_SESSION}-fuzz-report.md" << EOF
# PPA Protocol Fuzz Testing Report

**Session:** $FUZZ_SESSION  
**Device:** $DEVICE_IP:$PORT  
**Generated:** $(date)

## Overview

This report summarizes the results of edge case and fuzz testing on the PPA protocol implementation.

## Test Results

EOF

# Add results for each test
for analysis_file in "$ANALYSIS_DIR/${FUZZ_SESSION}"-*.txt; do
    if [ -f "$analysis_file" ]; then
        test_name=$(basename "$analysis_file" | sed 's/.*-\(.*\)\.txt/\1/')
        echo "### $test_name" >> "$ANALYSIS_DIR/${FUZZ_SESSION}-fuzz-report.md"
        echo '```' >> "$ANALYSIS_DIR/${FUZZ_SESSION}-fuzz-report.md"
        cat "$analysis_file" >> "$ANALYSIS_DIR/${FUZZ_SESSION}-fuzz-report.md"
        echo '```' >> "$ANALYSIS_DIR/${FUZZ_SESSION}-fuzz-report.md"
        echo >> "$ANALYSIS_DIR/${FUZZ_SESSION}-fuzz-report.md"
    fi
done

cat >> "$ANALYSIS_DIR/${FUZZ_SESSION}-fuzz-report.md" << EOF

## Protocol Robustness Analysis

### Error Handling
- ✅ Protocol should gracefully handle invalid inputs
- ⚠️  Error status codes should be returned for invalid operations
- ❌ Crashes or hangs indicate protocol implementation issues

### Connection Management
- ✅ Multiple connections should be handled properly
- ⚠️  Connection limits should be enforced
- ❌ Connection leaks or resource exhaustion issues

### Input Validation
- ✅ Invalid parameters should be rejected with proper error codes
- ⚠️  Boundary values should be handled correctly
- ❌ Buffer overflows or memory corruption vulnerabilities

### Timing and Concurrency
- ✅ Rapid commands should be processed in order
- ⚠️  Rate limiting may be necessary for production
- ❌ Race conditions or deadlocks indicate implementation issues

## Recommendations

1. **Improve Error Handling**: Add more specific error codes for different failure modes
2. **Input Validation**: Strengthen parameter validation in client and server
3. **Rate Limiting**: Implement rate limiting for production deployments  
4. **Connection Management**: Add connection pooling and proper cleanup
5. **Monitoring**: Add metrics for error rates and connection counts

## Security Considerations

1. **DoS Protection**: Rapid fire commands could be used for denial of service
2. **Input Sanitization**: Ensure all inputs are properly validated
3. **Resource Limits**: Implement limits on concurrent connections
4. **Authentication**: Consider adding authentication for production use

## Next Steps

1. **Fix Critical Issues**: Address any crashes or hangs discovered
2. **Enhance Validation**: Improve input validation based on edge cases found
3. **Add Tests**: Create unit tests for discovered edge cases
4. **Documentation**: Update protocol docs with error handling details
EOF

echo
echo "✅ Fuzz testing complete!"
echo
echo "📁 Generated files:"
ls -la "$ANALYSIS_DIR/${FUZZ_SESSION}"* 2>/dev/null

echo
echo "📖 Read the fuzz testing report:"
echo "   cat $ANALYSIS_DIR/${FUZZ_SESSION}-fuzz-report.md"

echo
echo "⚠️  Review findings and address any critical issues before production deployment"
