package cmds

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"ppa-control/lib/client"
	"ppa-control/lib/client/discovery"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// CommandConfig holds common configuration for commands
type CommandConfig struct {
	Addresses   string
	Discovery   bool
	ComponentID uint
	Port        uint
	Interfaces  []string
}

// CommandChannels holds common channels used across commands
type CommandChannels struct {
	DiscoveryCh chan discovery.PeerInformation
	ReceivedCh  chan client.ReceivedMessage
}

// CommandContext encapsulates all command execution context and resources
type CommandContext struct {
	Config      *CommandConfig
	Channels    *CommandChannels
	ctx         context.Context
	cancelFunc  context.CancelFunc
	group       *errgroup.Group
	multiClient *client.MultiClient
}

// Context returns the context.Context for this command
func (cc *CommandContext) Context() context.Context {
	return cc.ctx
}

// Cancel cancels the command's context
func (cc *CommandContext) Cancel() {
	if cc.cancelFunc != nil {
		log.Debug().Msg("Cancelling context")
		cc.cancelFunc()
	}
}

// Wait waits for all goroutines to complete and returns any error
func (cc *CommandContext) Wait() error {
	err := cc.group.Wait()
	log.Debug().Err(err).Msg("finished command execution")
	HandleCommandError(err)
	return err
}

// RunInGroup runs the given function in the error group
func (cc *CommandContext) RunInGroup(f func() error) {
	cc.group.Go(f)
}

// SetupCommand initializes common command configuration and context
func SetupCommand(cmd *cobra.Command) *CommandContext {
	cfg := &CommandConfig{
		Addresses:   cmd.Flag("addresses").Value.String(),
		Discovery:   cmd.Flag("discover").Value.String() == "true",
		ComponentID: uint(mustUint(cmd.Flag("componentId").Value.String())),
		Port:        uint(mustUint(cmd.Flag("port").Value.String())),
	}

	if cfg.Discovery {
		if interfaces, err := cmd.Flags().GetStringArray("interfaces"); err == nil {
			cfg.Interfaces = interfaces
		}
	}

	channels := &CommandChannels{
		DiscoveryCh: make(chan discovery.PeerInformation),
		ReceivedCh:  make(chan client.ReceivedMessage),
	}

	// Setup context with cancellation
	ctx := context.Background()
	ctx, cancelFunc := signal.NotifyContext(ctx, os.Interrupt)

	// Create error group
	grp, ctx := errgroup.WithContext(ctx)

	return &CommandContext{
		Config:     cfg,
		Channels:   channels,
		ctx:        ctx,
		cancelFunc: cancelFunc,
		group:      grp,
	}
}

// SetupMultiClient creates and configures a MultiClient with the given configuration
func (cc *CommandContext) SetupMultiClient(name string) error {
	multiClient := client.NewMultiClient(name)

	// Add clients for specified addresses
	for _, addr := range strings.Split(cc.Config.Addresses, ",") {
		if addr == "" {
			continue
		}
		_, err := multiClient.AddClient(cc.ctx, fmt.Sprintf("%s:%d", addr, cc.Config.Port), "", cc.Config.ComponentID)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to add client")
			return err
		}
	}

	cc.multiClient = multiClient
	return nil
}

// GetMultiClient returns the configured MultiClient
func (cc *CommandContext) GetMultiClient() *client.MultiClient {
	return cc.multiClient
}

// SetupDiscovery starts the discovery process if enabled
func (cc *CommandContext) SetupDiscovery() {
	if cc.Config.Discovery {
		cc.group.Go(func() error {
			return discovery.Discover(cc.ctx, cc.Channels.DiscoveryCh, cc.Config.Interfaces, uint16(cc.Config.Port))
		})
	}
}

// StartMultiClient starts the MultiClient in the error group
func (cc *CommandContext) StartMultiClient() {
	cc.group.Go(func() error {
		return cc.multiClient.Run(cc.ctx, cc.Channels.ReceivedCh)
	})
}

// HandleDiscoveryMessage processes discovery messages and updates the MultiClient accordingly
func (cc *CommandContext) HandleDiscoveryMessage(msg discovery.PeerInformation) (client.Client, error) {
	switch msg.(type) {
	case discovery.PeerDiscovered:
		log.Info().
			Str("addr", msg.GetAddress()).
			Str("iface", msg.GetInterface()).
			Msg("peer discovered")
		return cc.multiClient.AddClient(cc.ctx, msg.GetAddress(), msg.GetInterface(), cc.Config.ComponentID)
	case discovery.PeerLost:
		log.Info().
			Str("addr", msg.GetAddress()).
			Str("iface", msg.GetInterface()).
			Msg("peer lost")
		err := cc.multiClient.CancelClient(msg.GetAddress())
		if err != nil {
			log.Error().Err(err).Msg("failed to remove client")
			return nil, err
		}
	}
	return nil, nil
}

// HandleCommandError processes command errors consistently
func HandleCommandError(err error) {
	if err != nil && err.Error() != "context canceled" {
		log.Error().Err(err).Msg("Error running multiclient")
	}
}

// mustUint converts a string to uint, panicking if it fails
func mustUint(s string) uint64 {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return v
}
