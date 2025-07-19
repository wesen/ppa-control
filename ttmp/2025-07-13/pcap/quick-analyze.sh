#!/bin/bash
PCAP_FILE="${1:-captures/latest-session.pcapng}"
PCAP_TOOL="../../../pcap"

echo "=== Quick Analysis: $(basename $PCAP_FILE) ==="
echo "Message types:"
"$PCAP_TOOL" --output-format jsonl "$PCAP_FILE" 2>/dev/null | \
    jq -r '.header.message_type // "unknown"' | \
    sort | uniq -c | sort -nr

echo -e "\nStatus codes:"
"$PCAP_TOOL" --output-format jsonl "$PCAP_FILE" 2>/dev/null | \
    jq -r '.header.status // "unknown"' | \
    sort | uniq -c | sort -nr

echo -e "\nType 6/9 activity (config sync):"
"$PCAP_TOOL" --output-format jsonl "$PCAP_FILE" 2>/dev/null | \
    jq -r 'select(.header.message_type == 6 or .header.message_type == 9) | [.timestamp, .header.message_type, .header.status]' | \
    head -20