#!/bin/bash
# Specific PPA Operations Recording Script
# Records individual operation types for detailed protocol analysis

set -e

# Configuration
DEVICE_IP="${1:-192.168.0.200}"
PORT="5001"
PCAP_DIR="$(dirname "$0")/captures"
SESSION_PREFIX="ppa-ops-$(date +%Y%m%d-%H%M%S)"

mkdir -p "$PCAP_DIR"

echo "=== PPA Protocol Operation Recording ==="
echo "Device: $DEVICE_IP:$PORT"
echo "Session: $SESSION_PREFIX"
echo

# Function to record specific operation
record_operation() {
    local op_name="$1"
    local description="$2"
    local duration="${3:-30}"
    local extra_notes="$4"
    
    echo "┌─ Recording: $op_name ─"
    echo "│ $description"
    if [ -n "$extra_notes" ]; then
        echo "│ $extra_notes"
    fi
    echo "│ Duration: ${duration}s"
    echo "│"
    
    read -p "│ Press ENTER when ready to start recording $op_name..."
    echo "│ ⏺️  RECORDING - Perform the operation now!"
    
    # Start capture
    timeout "$duration" sudo tcpdump -i any -w "$PCAP_DIR/${SESSION_PREFIX}-${op_name}.pcap" \
        "host $DEVICE_IP and port $PORT" 2>/dev/null || true
    
    if [ -f "$PCAP_DIR/${SESSION_PREFIX}-${op_name}.pcap" ]; then
        local packet_count=$(tcpdump -r "$PCAP_DIR/${SESSION_PREFIX}-${op_name}.pcap" 2>/dev/null | wc -l)
        echo "│ ✅ Captured $packet_count packets"
    else
        echo "│ ❌ Capture failed"
    fi
    echo "└─────────────────────────────"
    echo
}

# Build pcap tool if needed
if [ ! -f "$(dirname "$0")/../../../pcap" ]; then
    echo "Building pcap tool..."
    cd "$(dirname "$0")/../../.."
    go build ./cmd/pcap
    cd - > /dev/null
fi

echo "🎯 This script will record specific PPA operations"
echo "   Follow the prompts to capture different protocol interactions"
echo

# Connection and Discovery
# record_operation "01-connection" \
    # "Initial connection and handshake" \
    # 45 \
    # "Start your PPA client and connect to the device"
# 
# record_operation "02-device-discovery" \
    # "Device discovery and enumeration" \
    # 30 \
    # "Trigger device scan/discovery in your PPA software"
# 
# record_operation "03-ping-keepalive" \
    # "Ping and keepalive messages" \
    # 60 \
    # "Let the connection idle to capture ping/keepalive traffic"
# 
# # Control Operations  
# record_operation "04-volume-control" \
    # "Volume/gain adjustments" \
    # 45 \
    # "Adjust input/output volumes and gains"
# 
# record_operation "05-mute-operations" \
    # "Mute and unmute operations" \
    # 30 \
    # "Toggle mute on various channels"

record_operation "06-eq-adjustments" \
    "EQ parameter changes" \
    60 \
    "Modify EQ settings: frequency, gain, Q factor"

record_operation "07-preset-recall" \
    "Preset recall operations" \
    30 \
    "Load different presets (try presets 1, 2, 3)"

record_operation "08-preset-save" \
    "Preset save operations" \
    30 \
    "Save current settings to a preset slot"

record_operation "09-delay-phase" \
    "Delay and phase adjustments" \
    45 \
    "Adjust delay times and phase inversion settings"

record_operation "10-realtime-monitoring" \
    "Real-time parameter monitoring" \
    120 \
    "Enable real-time monitoring/metering in your software"

record_operation "11-error-conditions" \
    "Error conditions and recovery" \
    60 \
    "Try invalid operations or disconnect/reconnect"

record_operation "12-rapid-commands" \
    "Rapid command sequence" \
    30 \
    "Send multiple commands quickly to test sequence handling"

echo "🎉 Recording session complete!"
echo
echo "📁 Captured files:"
ls -la "$PCAP_DIR/${SESSION_PREFIX}"*.pcap 2>/dev/null || echo "No captures found"

echo
echo "🔍 Quick analysis commands:"
echo
echo "# Message type summary for all captures"
echo "for f in $PCAP_DIR/${SESSION_PREFIX}*.pcap; do"
echo "  echo \"=== \$(basename \$f) ===\""
echo "  ../../../pcap --output-format jsonl \"\$f\" | jq -r '.header.message_type' | sort | uniq -c"
echo "done"
echo
echo "# Detailed analysis of specific operation"
echo "../../../pcap --print-hexdump --output-format text $PCAP_DIR/${SESSION_PREFIX}-04-volume-control.pcap"
echo
echo "# Export all to JSON for analysis"
echo "for f in $PCAP_DIR/${SESSION_PREFIX}*.pcap; do"
echo "  ../../../pcap --output-format json \"\$f\" > \"\${f%.pcap}.json\""
echo "done"
echo
echo "Run the analysis script:"
echo "./analyze-ppa-captures.sh ${SESSION_PREFIX}"
