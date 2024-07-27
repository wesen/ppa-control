package main

import (
	"ppa-control/lib/protocol"
	"testing"
)

func TestPacketHandler(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		acceptedTypes  []protocol.MessageType
		rejectedTypes  []protocol.MessageType
		acceptUnknown  bool
		rejectUnknown  bool
		acceptAllOther bool
	}{
		{
			name:           "All packets",
			input:          "all",
			acceptUnknown:  true,
			acceptAllOther: true,
		},
		{
			name:          "Single known type",
			input:         "deviceData",
			acceptedTypes: []protocol.MessageType{protocol.MessageTypeDeviceData},
		},
		{
			name:          "Multiple known types",
			input:         "ping,liveCmd,presetRecall",
			acceptedTypes: []protocol.MessageType{protocol.MessageTypePing, protocol.MessageTypeLiveCmd, protocol.MessageTypePresetRecall},
		},
		{
			name:          "Unknown type",
			input:         "unknown",
			acceptUnknown: true,
		},
		{
			name:           "Blacklisted type",
			input:          "all,-deviceData",
			rejectedTypes:  []protocol.MessageType{protocol.MessageTypeDeviceData},
			acceptAllOther: true,
		},
		{
			name:          "Mixed whitelist and blacklist",
			input:         "ping,liveCmd,-presetRecall",
			acceptedTypes: []protocol.MessageType{protocol.MessageTypePing, protocol.MessageTypeLiveCmd},
			rejectedTypes: []protocol.MessageType{protocol.MessageTypePresetRecall},
		},
		{
			name:           "All with blacklist",
			input:          "all,-deviceData",
			rejectedTypes:  []protocol.MessageType{protocol.MessageTypeDeviceData},
			acceptUnknown:  true,
			acceptAllOther: true,
		},
		{
			name:          "Custom type by hex value",
			input:         "type:AA",
			acceptedTypes: []protocol.MessageType{protocol.MessageType(0xAA)},
		},
		{
			name:          "Mixed known and custom types",
			input:         "ping,type:AA,deviceData",
			acceptedTypes: []protocol.MessageType{protocol.MessageTypePing, protocol.MessageType(0xAA), protocol.MessageTypeDeviceData},
		},
		{
			name:  "Invalid type",
			input: "invalidType",
		},
		{
			name:  "Empty input",
			input: "",
		},
		{
			name:          "Whitespace handling",
			input:         " ping , liveCmd , presetRecall ",
			acceptedTypes: []protocol.MessageType{protocol.MessageTypePing, protocol.MessageTypeLiveCmd, protocol.MessageTypePresetRecall},
		},
		{
			name:          "Multiple custom types",
			input:         "type:A1,type:B2,type:C3",
			acceptedTypes: []protocol.MessageType{protocol.MessageType(0xA1), protocol.MessageType(0xB2), protocol.MessageType(0xC3)},
		},
		{
			name:          "Mixed custom and known types",
			input:         "type:D4,ping,type:E5,deviceData",
			acceptedTypes: []protocol.MessageType{protocol.MessageType(0xD4), protocol.MessageTypePing, protocol.MessageType(0xE5), protocol.MessageTypeDeviceData},
		},
		{
			name:           "Reject unknown, allow specific custom type",
			input:          "all,-unknown,type:F6",
			acceptedTypes:  []protocol.MessageType{protocol.MessageType(0xF6)},
			rejectUnknown:  true,
			acceptAllOther: true,
		},
		{
			name:           "Allow all except specific custom type",
			input:          "all,-type:C7",
			rejectedTypes:  []protocol.MessageType{protocol.MessageType(0xC7)},
			acceptUnknown:  true,
			acceptAllOther: true,
		},
		{
			name:          "Whitelist known types, blacklist custom type",
			input:         "ping,liveCmd,presetRecall,-type:C8",
			acceptedTypes: []protocol.MessageType{protocol.MessageTypePing, protocol.MessageTypeLiveCmd, protocol.MessageTypePresetRecall},
			rejectedTypes: []protocol.MessageType{protocol.MessageType(0xC8)},
		},
		{
			name:          "Blacklist known type, whitelist custom type",
			input:         "-deviceData,type:C9",
			acceptedTypes: []protocol.MessageType{protocol.MessageType(0xC9)},
			rejectedTypes: []protocol.MessageType{protocol.MessageTypeDeviceData},
		},
		{
			name:          "Complex mix of known, unknown, and custom types",
			input:         "ping,unknown,type:C0,-presetRecall,-type:C1",
			acceptedTypes: []protocol.MessageType{protocol.MessageTypePing, protocol.MessageType(0xC0)},
			rejectedTypes: []protocol.MessageType{protocol.MessageTypePresetRecall, protocol.MessageType(0xC1)},
			acceptUnknown: true,
		},
		{
			name:          "Overlapping whitelist and blacklist",
			input:         "type:A2,-type:A2,ping,-ping",
			rejectedTypes: []protocol.MessageType{protocol.MessageType(0xA2), protocol.MessageTypePing},
		},
		{
			name:          "Invalid custom type format",
			input:         "type:XY,ping",
			acceptedTypes: []protocol.MessageType{protocol.MessageTypePing},
		},
		{
			name:  "Case sensitivity in known types",
			input: "PING,DeviceData,LiveCmd",
		},
		{
			name:          "Whitelist unknown, blacklist specific custom type",
			input:         "unknown,-type:B3",
			rejectedTypes: []protocol.MessageType{protocol.MessageType(0xB3)},
			acceptUnknown: true,
		},
		{
			name:          "Complex scenario with multiple custom types and known types",
			input:         "type:C4,ping,-type:D5,deviceData,unknown,-presetRecall,type:E6",
			acceptedTypes: []protocol.MessageType{protocol.MessageType(0xC4), protocol.MessageTypePing, protocol.MessageTypeDeviceData, protocol.MessageType(0xE6)},
			rejectedTypes: []protocol.MessageType{protocol.MessageType(0xD5), protocol.MessageTypePresetRecall},
			acceptUnknown: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ph := NewPacketHandler(tt.input, false, 0)

			// Test accepted types
			for _, msgType := range tt.acceptedTypes {
				if !ph.shouldProcessPacket(msgType) {
					t.Errorf("Expected to accept message type %x, but it was rejected", byte(msgType))
				}
			}

			// Test rejected types
			for _, msgType := range tt.rejectedTypes {
				if ph.shouldProcessPacket(msgType) {
					t.Errorf("Expected to reject message type %x, but it was accepted", byte(msgType))
				}
			}

			// Test unknown type
			if tt.acceptUnknown && !ph.shouldProcessPacket(protocol.MessageTypeUnknown) {
				t.Error("Expected to accept unknown message type, but it was rejected")
			}
			if tt.rejectUnknown && ph.shouldProcessPacket(protocol.MessageTypeUnknown) {
				t.Error("Expected to reject unknown message type, but it was accepted")
			}

			// Test all other types
			for i := 0; i < 255; i++ {
				msgType := protocol.MessageType(i)
				if protocol.IsMessageTypeUnknown(msgType) && (tt.acceptUnknown || tt.rejectUnknown) {
					continue // Skip unknown type as it's handled separately
				}
				if contains(tt.acceptedTypes, msgType) || contains(tt.rejectedTypes, msgType) {
					continue // Skip types that are explicitly tested
				}
				if tt.acceptAllOther && !ph.shouldProcessPacket(msgType) {
					t.Errorf("Expected to accept message type %x, but it was rejected", byte(msgType))
				}
				if !tt.acceptAllOther && ph.shouldProcessPacket(msgType) {
					t.Errorf("Expected to reject message type %x, but it was accepted", byte(msgType))
				}
			}
		})
	}
}

func contains(slice []protocol.MessageType, item protocol.MessageType) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
