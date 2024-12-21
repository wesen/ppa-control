package client

import "context"

// Commander defines the command-sending capabilities of a client
type Commander interface {
	SendPing()
	SendPresetRecallByPresetIndex(index int)
	SendMasterVolume(volume float32)
}

// Client extends Commander with lifecycle management
type Client interface {
	Commander
	Run(ctx context.Context, receivedCh chan<- ReceivedMessage) error
	Name() string
} 