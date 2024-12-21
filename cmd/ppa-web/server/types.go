package server

import "time"

// DeviceInfo holds information about a discovered device
type DeviceInfo struct {
	Address   string
	Interface string
	LastSeen  time.Time
}

// AppState represents the current state of the application
type AppState struct {
	DestIP            string
	Log               []string
	Status            string
	DiscoveryEnabled  bool
	DiscoveredDevices map[string]DeviceInfo // addr -> info
	ActiveInterfaces  map[string]bool       // iface -> active
}

// PacketInfo represents a network packet with all its details
type PacketInfo struct {
	Timestamp   string      `json:"timestamp"`
	Direction   string      `json:"direction"`
	Source      string      `json:"source"`
	Destination string      `json:"destination"`
	Header      interface{} `json:"header,omitempty"`
	Payload     interface{} `json:"payload,omitempty"`
	HexDump     string      `json:"hexDump,omitempty"`
}
