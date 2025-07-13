#!/bin/bash
# Quick PPA Protocol Capture Script
# For rapid protocol inspection and debugging

DEVICE_IP="${1:-192.168.0.200}"
PORT="5001"
DURATION="${2:-30}"
OUTPUT_FILE="${3:-ppa-quick-$(date +%H%M%S).pcap}"

PCAP_DIR="$(dirname "$0")/captures"
PCAP_TOOL="$(dirname "$0")/../../../pcap"

mkdir -p "$PCAP_DIR"

echo "🎯 Quick PPA Protocol Capture"
echo "Device: $DEVICE_IP:$PORT"
echo "Duration: ${DURATION}s"
echo "Output: $PCAP_DIR/$OUTPUT_FILE"
echo

# Build pcap tool if needed
if [ ! -f "$PCAP_TOOL" ]; then
    echo "Building pcap tool..."
    cd "$(dirname "$0")/../../.."
    go build ./cmd/pcap
    cd - > /dev/null
fi

echo "⏺️  Starting capture... (Press Ctrl+C to stop early)"
echo "Perform your PPA operations now!"

# Get the default network interface
DEFAULT_INTERFACE=$(ip route | grep default | head -1 | awk '{print $5}')

# Start capture
timeout "$DURATION" sudo tcpdump -i "$DEFAULT_INTERFACE" -w "$PCAP_DIR/$OUTPUT_FILE" \
    "host $DEVICE_IP and port $PORT" 2>/dev/null || true

if [ -f "$PCAP_DIR/$OUTPUT_FILE" ]; then
    packet_count=$(tcpdump -r "$PCAP_DIR/$OUTPUT_FILE" 2>/dev/null | wc -l)
    echo
    echo "✅ Captured $packet_count packets to $PCAP_DIR/$OUTPUT_FILE"
    
    if [ "$packet_count" -gt 0 ]; then
        echo
        echo "📊 Quick Analysis:"
        echo "=================="
        
        # Message type summary
        echo "Message Types:"
        "$PCAP_TOOL" --output-format jsonl "$PCAP_DIR/$OUTPUT_FILE" 2>/dev/null | \
            jq -r '.header.message_type // "unknown"' | \
            sort | uniq -c | sed 's/^/  /'
        
        echo
        echo "Status Codes:"
        "$PCAP_TOOL" --output-format jsonl "$PCAP_DIR/$OUTPUT_FILE" 2>/dev/null | \
            jq -r '.header.status // "unknown"' | \
            sort | uniq -c | sed 's/^/  /'
        
        echo
        echo "🔍 Commands for detailed analysis:"
        echo "# View all packets"
        echo "$PCAP_TOOL --output-format text $PCAP_DIR/$OUTPUT_FILE"
        echo
        echo "# View with hex dumps"
        echo "$PCAP_TOOL --print-hexdump $PCAP_DIR/$OUTPUT_FILE"
        echo
        echo "# Export to JSON"
        echo "$PCAP_TOOL --output-format json $PCAP_DIR/$OUTPUT_FILE > ${OUTPUT_FILE%.pcap}.json"
        echo
        echo "# Filter specific message types"
        echo "$PCAP_TOOL --print-packets liveCmd $PCAP_DIR/$OUTPUT_FILE"
        echo "$PCAP_TOOL --print-packets ping $PCAP_DIR/$OUTPUT_FILE"
        echo "$PCAP_TOOL --print-packets deviceData $PCAP_DIR/$OUTPUT_FILE"
    else
        echo "❌ No packets captured - check device connectivity and network interface"
    fi
else
    echo "❌ Capture failed - check permissions and network interface"
fi
