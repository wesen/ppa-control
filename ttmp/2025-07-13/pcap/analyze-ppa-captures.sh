#!/bin/bash
# PPA Protocol PCAP Analysis Script
# Analyzes captured PPA protocol traffic for reverse engineering

set -e

# Configuration
SESSION_PREFIX="${1:-}"
PCAP_DIR="$(dirname "$0")/captures"
ANALYSIS_DIR="$(dirname "$0")/analysis"
PCAP_TOOL="$(dirname "$0")/../../../pcap"

if [ -z "$SESSION_PREFIX" ]; then
    echo "Usage: $0 <session-prefix>"
    echo
    echo "Available sessions:"
    ls -1 "$PCAP_DIR"/*.pcap 2>/dev/null | sed 's/.*\///; s/\.pcap$//' | cut -d'-' -f1-3 | sort -u || echo "No captures found"
    exit 1
fi

mkdir -p "$ANALYSIS_DIR"

# Build pcap tool if needed
if [ ! -f "$PCAP_TOOL" ]; then
    echo "Building pcap tool..."
    cd "$(dirname "$0")/../../.."
    go build ./cmd/pcap
    cd - > /dev/null
fi

echo "🔍 PPA Protocol Analysis"
echo "Session: $SESSION_PREFIX"
echo "Input: $PCAP_DIR"
echo "Output: $ANALYSIS_DIR"
echo

# Find all PCAP files for this session
PCAP_FILES=($(ls "$PCAP_DIR/${SESSION_PREFIX}"*.pcap 2>/dev/null || true))

if [ ${#PCAP_FILES[@]} -eq 0 ]; then
    echo "❌ No PCAP files found for session: $SESSION_PREFIX"
    exit 1
fi

echo "📁 Found ${#PCAP_FILES[@]} capture files:"
for file in "${PCAP_FILES[@]}"; do
    echo "  $(basename "$file")"
done
echo

# Analysis functions
analyze_message_types() {
    echo "=== Message Type Analysis ===" > "$ANALYSIS_DIR/${SESSION_PREFIX}-message-types.txt"
    echo >> "$ANALYSIS_DIR/${SESSION_PREFIX}-message-types.txt"
    
    for file in "${PCAP_FILES[@]}"; do
        echo "📊 Analyzing message types in $(basename "$file")..."
        echo "File: $(basename "$file")" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-message-types.txt"
        
        if [ -s "$file" ]; then
            # Message type counts
            "$PCAP_TOOL" --output-format jsonl "$file" 2>/dev/null | \
                jq -r '.header.message_type // "unknown"' | \
                sort | uniq -c | \
                sed 's/^/  /' >> "$ANALYSIS_DIR/${SESSION_PREFIX}-message-types.txt"
        else
            echo "  (empty file)" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-message-types.txt"
        fi
        echo >> "$ANALYSIS_DIR/${SESSION_PREFIX}-message-types.txt"
    done
}

analyze_status_codes() {
    echo "=== Status Code Analysis ===" > "$ANALYSIS_DIR/${SESSION_PREFIX}-status-codes.txt"
    echo >> "$ANALYSIS_DIR/${SESSION_PREFIX}-status-codes.txt"
    
    for file in "${PCAP_FILES[@]}"; do
        echo "📊 Analyzing status codes in $(basename "$file")..."
        echo "File: $(basename "$file")" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-status-codes.txt"
        
        if [ -s "$file" ]; then
            # Status code analysis with hex values
            "$PCAP_TOOL" --output-format jsonl "$file" 2>/dev/null | \
                jq -r '.header.status // "unknown"' | \
                sort | uniq -c | \
                sed 's/^/  /' >> "$ANALYSIS_DIR/${SESSION_PREFIX}-status-codes.txt"
        else
            echo "  (empty file)" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-status-codes.txt"
        fi
        echo >> "$ANALYSIS_DIR/${SESSION_PREFIX}-status-codes.txt"
    done
}

analyze_sequence_numbers() {
    echo "=== Sequence Number Analysis ===" > "$ANALYSIS_DIR/${SESSION_PREFIX}-sequences.txt"
    echo >> "$ANALYSIS_DIR/${SESSION_PREFIX}-sequences.txt"
    
    for file in "${PCAP_FILES[@]}"; do
        echo "📊 Analyzing sequence numbers in $(basename "$file")..."
        echo "File: $(basename "$file")" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-sequences.txt"
        
        if [ -s "$file" ]; then
            # Sequence number ranges and patterns
            "$PCAP_TOOL" --output-format jsonl "$file" 2>/dev/null | \
                jq -r '.header.sequence_number // 0' | \
                sort -n | \
                awk 'BEGIN{print "  Range: [min-max], Count: total"} 
                     {if(NR==1) min=max=$1; if($1<min) min=$1; if($1>max) max=$1; count++} 
                     END{print "  Range: ["min"-"max"], Count: "count}' >> "$ANALYSIS_DIR/${SESSION_PREFIX}-sequences.txt"
        else
            echo "  (empty file)" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-sequences.txt"
        fi
        echo >> "$ANALYSIS_DIR/${SESSION_PREFIX}-sequences.txt"
    done
}

analyze_payload_patterns() {
    echo "=== Payload Pattern Analysis ===" > "$ANALYSIS_DIR/${SESSION_PREFIX}-payloads.txt"
    echo >> "$ANALYSIS_DIR/${SESSION_PREFIX}-payloads.txt"
    
    for file in "${PCAP_FILES[@]}"; do
        echo "📊 Analyzing payload patterns in $(basename "$file")..."
        echo "File: $(basename "$file")" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-payloads.txt"
        
        if [ -s "$file" ]; then
            # Payload size distribution
            echo "  Payload sizes:" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-payloads.txt"
            "$PCAP_TOOL" --output-format jsonl "$file" 2>/dev/null | \
                jq -r '.payload | if . == null then 0 else (. | tostring | length) end' | \
                sort -n | uniq -c | \
                sed 's/^/    /' >> "$ANALYSIS_DIR/${SESSION_PREFIX}-payloads.txt"
        else
            echo "  (empty file)" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-payloads.txt"
        fi
        echo >> "$ANALYSIS_DIR/${SESSION_PREFIX}-payloads.txt"
    done
}

extract_unknown_messages() {
    echo "📊 Extracting unknown message types..."
    echo "=== Unknown Message Analysis ===" > "$ANALYSIS_DIR/${SESSION_PREFIX}-unknown.txt"
    echo >> "$ANALYSIS_DIR/${SESSION_PREFIX}-unknown.txt"
    
    for file in "${PCAP_FILES[@]}"; do
        echo "File: $(basename "$file")" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-unknown.txt"
        
        if [ -s "$file" ]; then
            # Extract unknown messages with hex dumps
            unknown_count=$("$PCAP_TOOL" --print-packets unknown "$file" 2>/dev/null | wc -l)
            if [ "$unknown_count" -gt 0 ]; then
                echo "  Found $unknown_count unknown messages:" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-unknown.txt"
                "$PCAP_TOOL" --print-packets unknown --print-hexdump "$file" 2>/dev/null | \
                    head -50 >> "$ANALYSIS_DIR/${SESSION_PREFIX}-unknown.txt"
            else
                echo "  No unknown messages found" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-unknown.txt"
            fi
        else
            echo "  (empty file)" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-unknown.txt"
        fi
        echo >> "$ANALYSIS_DIR/${SESSION_PREFIX}-unknown.txt"
    done
}

generate_json_exports() {
    echo "📊 Generating JSON exports..."
    for file in "${PCAP_FILES[@]}"; do
        if [ -s "$file" ]; then
            basename_no_ext=$(basename "$file" .pcap)
            echo "  Exporting $(basename "$file") to JSON..."
            "$PCAP_TOOL" --output-format json "$file" > "$ANALYSIS_DIR/${basename_no_ext}.json" 2>/dev/null || true
        fi
    done
}

analyze_livecmd_patterns() {
    echo "📊 Analyzing LiveCmd patterns..."
    echo "=== LiveCmd Analysis ===" > "$ANALYSIS_DIR/${SESSION_PREFIX}-livecmd.txt"
    echo >> "$ANALYSIS_DIR/${SESSION_PREFIX}-livecmd.txt"
    
    for file in "${PCAP_FILES[@]}"; do
        echo "File: $(basename "$file")" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-livecmd.txt"
        
        if [ -s "$file" ]; then
            # Extract LiveCmd messages
            livecmd_count=$("$PCAP_TOOL" --print-packets liveCmd "$file" 2>/dev/null | wc -l)
            if [ "$livecmd_count" -gt 0 ]; then
                echo "  Found $livecmd_count LiveCmd messages" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-livecmd.txt"
                echo "  Sample LiveCmd messages:" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-livecmd.txt"
                "$PCAP_TOOL" --print-packets liveCmd --output-format text "$file" 2>/dev/null | \
                    head -20 | sed 's/^/    /' >> "$ANALYSIS_DIR/${SESSION_PREFIX}-livecmd.txt"
                
                # Extract path patterns
                echo "  Path patterns:" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-livecmd.txt"
                "$PCAP_TOOL" --print-packets liveCmd --output-format jsonl "$file" 2>/dev/null | \
                    jq -r '.payload.path // empty' | \
                    sort | uniq -c | \
                    sed 's/^/    /' >> "$ANALYSIS_DIR/${SESSION_PREFIX}-livecmd.txt"
            else
                echo "  No LiveCmd messages found" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-livecmd.txt"
            fi
        else
            echo "  (empty file)" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-livecmd.txt"
        fi
        echo >> "$ANALYSIS_DIR/${SESSION_PREFIX}-livecmd.txt"
    done
}

create_summary_report() {
    echo "📊 Creating summary report..."
    
    cat > "$ANALYSIS_DIR/${SESSION_PREFIX}-summary-report.md" << EOF
# PPA Protocol Analysis Report

**Session:** $SESSION_PREFIX  
**Generated:** $(date)  
**Files analyzed:** ${#PCAP_FILES[@]}

## Files Analyzed

EOF
    
    for file in "${PCAP_FILES[@]}"; do
        packet_count=$(tcpdump -r "$file" 2>/dev/null | wc -l || echo 0)
        file_size=$(ls -lh "$file" | awk '{print $5}')
        echo "- \`$(basename "$file")\` - $packet_count packets, $file_size" >> "$ANALYSIS_DIR/${SESSION_PREFIX}-summary-report.md"
    done
    
    cat >> "$ANALYSIS_DIR/${SESSION_PREFIX}-summary-report.md" << EOF

## Key Findings

### Message Type Distribution
\`\`\`
$(cat "$ANALYSIS_DIR/${SESSION_PREFIX}-message-types.txt" | grep -A 20 "=== Message Type Analysis ===" | tail -n +3)
\`\`\`

### Protocol Validation

Compare these findings with the protocol specification:

- **Expected Message Types**: Ping(0), LiveCmd(1), DeviceData(2), PresetRecall(4), PresetSave(5)
- **Expected Status Codes**: 
  - Client: 0x0102 (Command), 0x0106 (Request), 0x0101 (Response), 0x0109 (Error), 0x0141 (Wait)
  - Server: 0x0002 (Command), 0x0006 (Request), 0x0001 (Response), 0x0009 (Error), 0x0041 (Wait)

### Analysis Files Generated

- \`${SESSION_PREFIX}-message-types.txt\` - Message type distribution
- \`${SESSION_PREFIX}-status-codes.txt\` - Status code analysis  
- \`${SESSION_PREFIX}-sequences.txt\` - Sequence number patterns
- \`${SESSION_PREFIX}-payloads.txt\` - Payload pattern analysis
- \`${SESSION_PREFIX}-unknown.txt\` - Unknown message investigation
- \`${SESSION_PREFIX}-livecmd.txt\` - LiveCmd pattern analysis
- \`*.json\` - Full packet data in JSON format

### Next Steps

1. **Validate Protocol Implementation**: Compare findings with \`lib/protocol/ppa-protocol.go\`
2. **Investigate Unknown Messages**: Review any unknown message types found
3. **Test Edge Cases**: Create test cases for discovered patterns
4. **Update Documentation**: Document any new protocol insights

### Commands for Further Analysis

\`\`\`bash
# View specific message type details
../../../pcap --print-packets <messageType> --print-hexdump file.pcap

# Export specific data for processing
../../../pcap --output-format jsonl file.pcap | jq '.header.message_type' | sort | uniq -c

# Analyze timing patterns  
../../../pcap --output-format jsonl file.pcap | jq -r '[.timestamp, .header.message_type] | @csv'
\`\`\`
EOF
}

# Run all analyses
echo "🚀 Starting comprehensive analysis..."

analyze_message_types
analyze_status_codes  
analyze_sequence_numbers
analyze_payload_patterns
extract_unknown_messages
generate_json_exports
analyze_livecmd_patterns
create_summary_report

echo
echo "✅ Analysis complete!"
echo
echo "📁 Generated files in $ANALYSIS_DIR:"
ls -la "$ANALYSIS_DIR/${SESSION_PREFIX}"* 2>/dev/null || echo "No analysis files generated"

echo
echo "📖 Read the summary report:"
echo "   cat $ANALYSIS_DIR/${SESSION_PREFIX}-summary-report.md"
echo
echo "🔍 Key analysis files:"
echo "   $ANALYSIS_DIR/${SESSION_PREFIX}-message-types.txt"
echo "   $ANALYSIS_DIR/${SESSION_PREFIX}-unknown.txt" 
echo "   $ANALYSIS_DIR/${SESSION_PREFIX}-livecmd.txt"
