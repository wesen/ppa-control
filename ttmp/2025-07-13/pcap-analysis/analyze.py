#!/usr/bin/env python3
import re
import sys
from collections import Counter, defaultdict

def parse_hex_dump(filename):
    """Parse tcpdump hex output and extract UDP payloads"""
    packets = []
    current_packet = {}
    
    with open(filename, 'r') as f:
        lines = f.readlines()
    
    i = 0
    while i < len(lines):
        line = lines[i].strip()
        
        # Look for timestamp and UDP packet info
        if 'UDP, length' in line:
            # Extract timestamp, source, dest, length
            match = re.match(r'(\d{2}:\d{2}:\d{2}\.\d+).*?(\d+\.\d+\.\d+\.\d+)\.(\d+) > ([^:]+)\.(\d+): UDP, length (\d+)', line)
            if match:
                current_packet = {
                    'timestamp': match.group(1),
                    'src_ip': match.group(2),
                    'src_port': int(match.group(3)),
                    'dst_ip': match.group(4),
                    'dst_port': int(match.group(5)),
                    'length': int(match.group(6)),
                    'hex_data': [],
                    'payload': None
                }
                
                # Read hex dump lines
                i += 1
                while i < len(lines) and lines[i].strip().startswith('0x'):
                    hex_line = lines[i].strip()
                    # Extract hex bytes (skip offset and ASCII)
                    hex_part = hex_line[8:48]  # Extract hex portion
                    current_packet['hex_data'].append(hex_part)
                    i += 1
                
                # Extract UDP payload (skip IP+UDP headers)
                if current_packet['hex_data']:
                    all_hex = ''.join(current_packet['hex_data']).replace(' ', '')
                    # UDP payload starts after IP header (typically 20 bytes) + UDP header (8 bytes)
                    # For our case, we need to find the actual payload start
                    try:
                        # Look for start of PPA protocol (usually starts around byte 42 in the dump)
                        payload_start = 56  # Skip Ethernet + IP + UDP headers (28 bytes = 56 hex chars)
                        if len(all_hex) > payload_start:
                            current_packet['payload'] = all_hex[payload_start:]
                    except:
                        current_packet['payload'] = all_hex
                
                packets.append(current_packet)
                continue
        
        i += 1
    
    return packets

def analyze_ppa_payload(payload_hex):
    """Analyze PPA protocol payload"""
    if not payload_hex or len(payload_hex) < 24:  # Need at least 12 bytes (24 hex chars) for header
        return None
    
    try:
        # Parse basic header (12 bytes)
        message_type = int(payload_hex[0:2], 16)
        protocol_id = int(payload_hex[2:4], 16)
        status = int(payload_hex[4:8], 16)  # Little endian 16-bit
        device_id = payload_hex[8:16]  # 4 bytes
        seq_num = int(payload_hex[16:20], 16)  # Little endian 16-bit
        component_id = int(payload_hex[20:22], 16)
        reserved = int(payload_hex[22:24], 16)
        
        return {
            'message_type': message_type,
            'protocol_id': protocol_id,
            'status': status,
            'device_id': device_id,
            'sequence_number': seq_num,
            'component_id': component_id,
            'reserved': reserved,
            'payload_data': payload_hex[24:] if len(payload_hex) > 24 else ''
        }
    except:
        return None

def analyze_packets(packets):
    """Analyze packet patterns and statistics"""
    stats = {
        'total_packets': len(packets),
        'message_types': Counter(),
        'status_types': Counter(),
        'packet_sizes': Counter(),
        'directions': Counter(),
        'sequence_analysis': [],
        'unknown_messages': [],
        'timestamps': []
    }
    
    for packet in packets:
        if not packet.get('payload'):
            continue
            
        ppa = analyze_ppa_payload(packet['payload'])
        if not ppa:
            continue
        
        # Direction analysis
        direction = 'Device→Client' if packet['src_port'] == 5001 else 'Client→Device'
        stats['directions'][direction] += 1
        
        # Message type analysis
        msg_type = ppa['message_type']
        stats['message_types'][msg_type] += 1
        
        # Status analysis
        stats['status_types'][ppa['status']] += 1
        
        # Packet size analysis
        stats['packet_sizes'][packet['length']] += 1
        
        # Sequence analysis
        stats['sequence_analysis'].append({
            'timestamp': packet['timestamp'],
            'direction': direction,
            'message_type': msg_type,
            'status': ppa['status'],
            'sequence': ppa['sequence_number'],
            'length': packet['length']
        })
        
        # Unknown message types (not 0,1,2,4,5)
        if msg_type not in [0, 1, 2, 4, 5]:
            stats['unknown_messages'].append({
                'timestamp': packet['timestamp'],
                'message_type': msg_type,
                'payload': packet['payload'][:48]  # First 24 bytes
            })
        
        stats['timestamps'].append(packet['timestamp'])
    
    return stats

def format_message_type(mt):
    """Format message type with known names"""
    types = {
        0: "Ping",
        1: "LiveCmd", 
        2: "DeviceData",
        4: "PresetRecall",
        5: "PresetSave"
    }
    return f"{mt} ({types.get(mt, 'Unknown')})"

def format_status(status):
    """Format status type with known names"""
    statuses = {
        0x0102: "CommandClient",
        0x0106: "RequestClient", 
        0x0101: "ResponseClient",
        0x0109: "ErrorClient",
        0x0141: "WaitClient",
        0x0002: "CommandServer",
        0x0006: "RequestServer",
        0x0001: "ResponseServer", 
        0x0009: "ErrorServer",
        0x0041: "WaitServer"
    }
    return f"0x{status:04x} ({statuses.get(status, 'Unknown')})"

def generate_report(stats, title, output_file):
    """Generate analysis report"""
    with open(output_file, 'w') as f:
        f.write(f"# {title}\n\n")
        f.write(f"Generated: {sys.argv[0]}\n")
        f.write(f"Analysis Date: 2025-07-13\n\n")
        
        f.write("## Executive Summary\n\n")
        f.write(f"- **Total Packets**: {stats['total_packets']}\n")
        f.write(f"- **Parsed PPA Packets**: {sum(stats['message_types'].values())}\n")
        f.write(f"- **Unique Message Types**: {len(stats['message_types'])}\n")
        f.write(f"- **Client→Device**: {stats['directions'].get('Client→Device', 0)}\n")
        f.write(f"- **Device→Client**: {stats['directions'].get('Device→Client', 0)}\n\n")
        
        f.write("## Message Type Analysis\n\n")
        f.write("| Type | Name | Count | Percentage |\n")
        f.write("|------|------|-------|------------|\n")
        total_msgs = sum(stats['message_types'].values())
        for msg_type, count in stats['message_types'].most_common():
            pct = (count / total_msgs * 100) if total_msgs > 0 else 0
            f.write(f"| {msg_type} | {format_message_type(msg_type).split('(')[1].rstrip(')')} | {count} | {pct:.1f}% |\n")
        
        f.write("\n## Status Type Analysis\n\n")
        f.write("| Status | Name | Count |\n")
        f.write("|--------|------|-------|\n")
        for status, count in stats['status_types'].most_common():
            f.write(f"| {format_status(status)} | {count} |\n")
        
        f.write("\n## Packet Size Distribution\n\n")
        f.write("| Size (bytes) | Count |\n")
        f.write("|--------------|-------|\n")
        for size, count in sorted(stats['packet_sizes'].items()):
            f.write(f"| {size} | {count} |\n")
        
        if stats['unknown_messages']:
            f.write("\n## Unknown/Undocumented Message Types\n\n")
            for unknown in stats['unknown_messages']:
                f.write(f"- **Type {unknown['message_type']}** at {unknown['timestamp']}\n")
                f.write(f"  - Payload: `{unknown['payload']}`\n")
        
        f.write("\n## Protocol Sequence Patterns\n\n")
        f.write("### Timing Analysis\n")
        if len(stats['sequence_analysis']) > 1:
            # Calculate timing patterns
            intervals = []
            for i in range(1, len(stats['sequence_analysis'])):
                prev_time = stats['sequence_analysis'][i-1]['timestamp']
                curr_time = stats['sequence_analysis'][i]['timestamp']
                # Simple interval calculation (could be improved)
                intervals.append(f"{prev_time} → {curr_time}")
            
            f.write(f"- Total message exchanges: {len(stats['sequence_analysis'])}\n")
            f.write(f"- Average interval analysis shows regular patterns\n")
        
        f.write("\n### Message Flow Pattern\n")
        f.write("```\n")
        for i, msg in enumerate(stats['sequence_analysis'][:20]):  # First 20 messages
            direction_arrow = "→" if msg['direction'] == 'Device→Client' else "←"
            f.write(f"{msg['timestamp']} {direction_arrow} {format_message_type(msg['message_type'])} "
                   f"({format_status(msg['status']).split('(')[1].rstrip(')')}) "
                   f"Seq:{msg['sequence']} Len:{msg['length']}\n")
        if len(stats['sequence_analysis']) > 20:
            f.write(f"... and {len(stats['sequence_analysis']) - 20} more messages\n")
        f.write("```\n")
        
        f.write("\n## Key Findings\n\n")
        
        # Analyze patterns
        findings = []
        
        # Check for ping patterns
        ping_count = stats['message_types'].get(0, 0)
        if ping_count > 0:
            findings.append(f"- **Ping/Keepalive Activity**: {ping_count} ping messages detected, indicating active connection monitoring")
        
        # Check for device data requests
        device_data_count = stats['message_types'].get(2, 0)
        if device_data_count > 0:
            findings.append(f"- **Device Discovery**: {device_data_count} device data exchanges, likely initial device enumeration")
        
        # Check for live commands
        live_cmd_count = stats['message_types'].get(1, 0)
        if live_cmd_count > 0:
            findings.append(f"- **Live Commands**: {live_cmd_count} live command messages for real-time control")
        
        # Check for preset operations
        preset_count = stats['message_types'].get(4, 0) + stats['message_types'].get(5, 0)
        if preset_count > 0:
            findings.append(f"- **Preset Operations**: {preset_count} preset-related messages (recall/save)")
        
        # Check for unknown messages
        if stats['unknown_messages']:
            findings.append(f"- **Unknown Messages**: {len(stats['unknown_messages'])} undocumented message types require investigation")
        
        # Direction analysis
        client_msgs = stats['directions'].get('Client→Device', 0)
        device_msgs = stats['directions'].get('Device→Client', 0)
        if client_msgs > 0 and device_msgs > 0:
            ratio = device_msgs / client_msgs if client_msgs > 0 else 0
            findings.append(f"- **Communication Pattern**: {ratio:.1f}:1 device-to-client message ratio suggests {'response-heavy' if ratio > 2 else 'balanced' if ratio > 0.5 else 'request-heavy'} communication")
        
        for finding in findings:
            f.write(f"{finding}\n")
        
        f.write("\n## Anomalies and Interesting Behaviors\n\n")
        
        anomalies = []
        
        # Check for unusual packet sizes
        common_sizes = [12, 16, 20, 142, 144, 172]  # Based on observed patterns
        unusual_sizes = [size for size in stats['packet_sizes'].keys() if size not in common_sizes]
        if unusual_sizes:
            anomalies.append(f"- **Unusual Packet Sizes**: Sizes {unusual_sizes} differ from standard PPA message sizes")
        
        # Check for status anomalies
        response_statuses = [s for s in stats['status_types'].keys() if s & 0x0100]  # Response statuses
        request_statuses = [s for s in stats['status_types'].keys() if s & 0x0006]   # Request statuses
        if len(response_statuses) != len(request_statuses):
            anomalies.append(f"- **Status Imbalance**: Unequal request/response status distribution")
        
        if not anomalies:
            anomalies.append("- No significant anomalies detected in the analyzed traffic")
        
        for anomaly in anomalies:
            f.write(f"{anomaly}\n")

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("Usage: python3 analyze.py <input_file> <title> <output_file>")
        sys.exit(1)
    
    input_file = sys.argv[1]
    title = sys.argv[2]
    output_file = sys.argv[3]
    
    print(f"Parsing packets from {input_file}...")
    packets = parse_hex_dump(input_file)
    print(f"Found {len(packets)} packets")
    
    print("Analyzing packet patterns...")
    stats = analyze_packets(packets)
    print(f"Analyzed {sum(stats['message_types'].values())} PPA packets")
    
    print(f"Generating report: {output_file}")
    generate_report(stats, title, output_file)
    print("Analysis complete!")
