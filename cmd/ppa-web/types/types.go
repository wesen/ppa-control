package types

import "time"

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
