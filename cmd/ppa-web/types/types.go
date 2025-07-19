package types

import "time"

// API types for packet analysis

// AppState represents the application state
type AppState struct {
	DestIP            string
	Log               []string
	Status            string
	DiscoveryEnabled  bool
	DiscoveredDevices map[string]DeviceInfo
	ActiveInterfaces  map[string]bool
}

type DeviceInfo struct {
	Address   string
	Interface string
	LastSeen  time.Time
}

// ServerInterface defines what handlers need from the server
type ServerInterface interface {
	GetState() AppState
	SetState(func(*AppState))
	StartDiscovery() error
	StopDiscovery() error
	ConnectToDevice(addr string) error
	IsConnected() bool
	LogPacket(format string, args ...interface{})
	LogPacketDetails(packet PacketInfo)
	AddUpdateListener(ch chan struct{})
	RemoveUpdateListener(ch chan struct{})
}

type PacketInfo struct {
	Timestamp   string
	Direction   string
	Source      string
	Destination string
	Header      map[string]interface{}
	Payload     []byte
	HexDump     string
}

// CaptureFile represents a PCAP file
type CaptureFile struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
	Path         string    `json:"path"`
}

// AnalysisFile represents an analysis output file
type AnalysisFile struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
	Path         string    `json:"path"`
}

// AnalysisResponse represents the response from analysis trigger
type AnalysisResponse struct {
	Session   string    `json:"session"`
	Status    string    `json:"status"`
	Output    string    `json:"output"`
	Timestamp time.Time `json:"timestamp"`
}

// AnalysisResult represents analysis results for a session
type AnalysisResult struct {
	Session string         `json:"session"`
	Files   []AnalysisFile `json:"files"`
}

// SearchMatch represents a search match in a file
type SearchMatch struct {
	LineNumber int      `json:"line_number"`
	Line       string   `json:"line"`
	Context    []string `json:"context"`
}

// SearchResult represents search results for a file
type SearchResult struct {
	File    string        `json:"file"`
	Matches []SearchMatch `json:"matches"`
}

// SearchResponse represents the search API response
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Count   int            `json:"count"`
}
