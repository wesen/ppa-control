package server

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"ppa-control/lib"
	"ppa-control/lib/client"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// Server represents the web server and manages the application state
type Server struct {
	state     AppState
	mu        sync.RWMutex
	cmdCtx    *lib.CommandContext
	receiveCh chan client.ReceivedMessage
}

// NewServer creates a new server instance for direct use
func NewServer() *Server {
	cmdCtx := &lib.CommandContext{
		Config: &lib.CommandConfig{
			ComponentID: 0xFF,
			Port:        5001,
		},
		Channels: &lib.CommandChannels{
			ReceivedCh: make(chan client.ReceivedMessage),
		},
	}

	// Setup context with cancellation
	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(context.Background())
	cmdCtx.SetupContext(ctx, cancelFunc)

	return &Server{
		state: AppState{
			DestIP: "",
			Log:    make([]string, 0),
			Status: "Disconnected",
		},
		cmdCtx:    cmdCtx,
		receiveCh: cmdCtx.Channels.ReceivedCh,
	}
}

// FromCobraCommand creates a new server instance from a cobra command
func FromCobraCommand(cmd *cobra.Command) *Server {
	cmdCtx := lib.SetupCommand(cmd)

	return &Server{
		state: AppState{
			DestIP: "",
			Log:    make([]string, 0),
			Status: "Disconnected",
		},
		cmdCtx:    cmdCtx,
		receiveCh: cmdCtx.Channels.ReceivedCh,
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
	if s.cmdCtx.GetMultiClient() != nil {
		s.cmdCtx.Cancel()
	}

	// Create new context and setup multiclient
	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(context.Background())
	s.cmdCtx.SetupContext(ctx, cancelFunc)

	if err := s.cmdCtx.SetupMultiClient("web"); err != nil {
		return fmt.Errorf("failed to setup client: %w", err)
	}

	c, err := s.cmdCtx.GetMultiClient().AddClient(ctx, fmt.Sprintf("%s:%d", addr, s.cmdCtx.Config.Port), "", s.cmdCtx.Config.ComponentID)
	if err != nil {
		return fmt.Errorf("failed to add client: %w", err)
	}

	// Start the client run loop
	s.cmdCtx.StartMultiClient()

	// Start the ping loop
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.cmdCtx.Context().Done():
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
	return s.cmdCtx.GetMultiClient() != nil
}
