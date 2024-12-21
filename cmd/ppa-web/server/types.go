package server

// AppState represents the current state of the application
type AppState struct {
	DestIP string
	Log    []string
	Status string
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
