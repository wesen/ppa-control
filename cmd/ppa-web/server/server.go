package server

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"ppa-control/lib/client"
	"sync"
	"time"
)

// Server represents the web server and manages the application state
type Server struct {
	state     AppState
	mu        sync.RWMutex
	client    *client.MultiClient
	ctx       context.Context
	cancel    context.CancelFunc
	receiveCh chan client.ReceivedMessage
}

// NewServer creates a new server instance
func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		state: AppState{
			DestIP: "",
			Log:    make([]string, 0),
			Status: "Disconnected",
		},
		ctx:       ctx,
		cancel:    cancel,
		receiveCh: make(chan client.ReceivedMessage),
	}
}

// LogPacket logs a formatted message to the state log
func (s *Server) LogPacket(format string, args ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	s.state.Log = append(s.state.Log, msg)
	if len(s.state.Log) > 100 {
		s.state.Log = s.state.Log[1:]
	}
}

// LogPacketDetails logs detailed packet information to the state log
func (s *Server) LogPacketDetails(packet PacketInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Convert to JSON for console logging
	jsonData, err := json.Marshal(packet)
	if err != nil {
		return
	}

	// Add a special marker that the frontend will recognize for console logging
	s.state.Log = append(s.state.Log, fmt.Sprintf("__PACKET__%s", string(jsonData)))
	if len(s.state.Log) > 100 {
		s.state.Log = s.state.Log[1:]
	}
}

// GetState returns a copy of the current application state
func (s *Server) GetState() AppState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// SetState updates the application state using the provided function
func (s *Server) SetState(fn func(*AppState)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.state)
}

// ConnectToDevice establishes a connection to a PPA device
func (s *Server) ConnectToDevice(addr string) error {
	if s.client != nil {
		s.cancel()
		s.client = nil
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.client = client.NewMultiClient("web")

	c, err := s.client.AddClient(s.ctx, fmt.Sprintf("%s:%d", addr, 5001), "", 0xFF)
	if err != nil {
		return fmt.Errorf("failed to add client: %w", err)
	}

	// Start the client run loop
	go func() {
		if err := s.client.Run(s.ctx, &s.receiveCh); err != nil {
			s.LogPacket("Client error: %v", err)
		}
	}()

	// Start the ping loop
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				c.SendPing()
				s.LogPacketDetails(PacketInfo{
					Timestamp:   time.Now().Format(time.RFC3339Nano),
					Direction:   "Client → Device",
					Source:      "Web Client",
					Destination: addr,
					Header: map[string]interface{}{
						"MessageType": "Ping",
						"Status":      "CommandClient",
					},
				})
			case msg := <-s.receiveCh:
				if msg.Header != nil {
					s.SetState(func(state *AppState) {
						state.Status = "Connected"
					})

					// Log received packet details
					packet := PacketInfo{
						Timestamp:   time.Now().Format(time.RFC3339Nano),
						Direction:   "Device → Client",
						Source:      msg.RemoteAddress.String(),
						Destination: "Web Client",
						Header:      msg.Header,
					}

					if msg.Data != nil {
						packet.Payload = msg.Data
						packet.HexDump = hex.Dump(msg.Data)
					}

					s.LogPacketDetails(packet)
				}
			}
		}
	}()

	return nil
}

// IsConnected returns true if the server is connected to a device
func (s *Server) IsConnected() bool {
	return s.client != nil
}
