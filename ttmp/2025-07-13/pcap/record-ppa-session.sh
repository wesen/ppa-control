#!/bin/bash
# PPA Protocol PCAP Recording Script
# Usage: ./record-ppa-session.sh [session-name] [duration-seconds] [device-ip]

set -e

# Configuration
DEFAULT_DEVICE_IP="192.168.0.200"
DEFAULT_PORT="5001"
DEFAULT_DURATION="300"  # 5 minutes
PCAP_DIR="$(dirname "$0")/captures"
ANALYSIS_DIR="$(dirname "$0")/analysis"

# Parse arguments
SESSION_NAME="${1:-ppa-session-$(date +%Y%m%d-%H%M%S)}"
DURATION="${2:-$DEFAULT_DURATION}"
DEVICE_IP="${3:-$DEFAULT_DEVICE_IP}"

# Ensure directories exist
mkdir -p "$PCAP_DIR" "$ANALYSIS_DIR"

echo "PPA Protocol Recording Session: $SESSION_NAME"
echo "Device IP: $DEVICE_IP:$DEFAULT_PORT"
echo "Duration: $DURATION seconds"
echo "Capture directory: $PCAP_DIR"
echo "Analysis directory: $ANALYSIS_DIR"
echo

# Build pcap tool if not exists
if [ ! -f "$(dirname "$0")/../../../pcap" ]; then
    echo "Building pcap tool..."
    cd "$(dirname "$0")/../../.."
    go build ./cmd/pcap
    cd - > /dev/null
fi

# Function to capture packets
capture_session() {
    local name="$1"
    local description="$2"
    local filter="$3"
    local duration="$4"
    
    echo "Recording: $description"
    echo "Filter: $filter"
    echo "Duration: ${duration}s"
    
    # Use timeout to limit capture duration
    timeout "$duration" sudo tcpdump -i any -w "$PCAP_DIR/${SESSION_NAME}-${name}.pcap" "$filter" 2>/dev/null || true
    
    if [ -f "$PCAP_DIR/${SESSION_NAME}-${name}.pcap" ]; then
        echo "✓ Captured: $PCAP_DIR/${SESSION_NAME}-${name}.pcap"
        # Quick analysis
        echo "  Packet count: $(tcpdump -r "$PCAP_DIR/${SESSION_NAME}-${name}.pcap" 2>/dev/null | wc -l)"
    else
        echo "✗ Failed to capture: ${name}"
    fi
    echo
}

# Main capture filter for PPA protocol
PPA_FILTER="host $DEVICE_IP and port $DEFAULT_PORT"

echo "Starting PCAP recording session..."
echo "Press Ctrl+C to stop early"
echo

# Record different scenarios based on user interaction
echo "=== Recording Full Session ==="
echo "Perform your PPA operations now..."
echo "This will capture all traffic for $DURATION seconds"
echo

# Full session capture
capture_session "full" "Complete PPA session" "$PPA_FILTER" "$DURATION"

echo "=== Recording Complete ==="
echo
echo "Captured files:"
ls -la "$PCAP_DIR/${SESSION_NAME}"*.pcap 2>/dev/null || echo "No captures found"

echo
echo "To analyze captures, run:"
echo "  ./analyze-ppa-captures.sh $SESSION_NAME"
echo
echo "Individual capture commands:"
echo "  # Connection handshake analysis"
echo "  ../../../pcap --output-format text $PCAP_DIR/${SESSION_NAME}-full.pcap | head -20"
echo
echo "  # Message type summary"
echo "  ../../../pcap --output-format jsonl $PCAP_DIR/${SESSION_NAME}-full.pcap | jq -r '.header.message_type' | sort | uniq -c"
echo
echo "  # Real-time monitoring"
echo "  ../../../pcap --print-packets liveCmd --print-hexdump $PCAP_DIR/${SESSION_NAME}-full.pcap"
