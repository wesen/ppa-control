package commands

import (
	"context"
	"time"

	"ppa-control/lib/glazed"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
)

// PingCommand handles PPA ping messages.
// It supports classic (human-readable) and structured Glazed output.
type PingCommand struct {
	*cmds.CommandDescription
}

// Run implements the BareCommand interface for classic text output
func (c *PingCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Initialize logging
	if err := glazed.InitLogging(parsedLayers); err != nil {
		return err
	}

	// Create command context from parsed layers
	cmdCtx, err := glazed.CreateCommandContextFromParsedLayers(ctx, parsedLayers)
	if err != nil {
		return err
	}
	defer cmdCtx.Cancel()

	// Setup multiclient
	if err := cmdCtx.SetupMultiClient("ping"); err != nil {
		log.Fatal().Err(err).Msg("Failed to setup multiclient")
		return err
	}

	// Setup discovery if enabled
	cmdCtx.SetupDiscovery()

	// Start multiclient
	cmdCtx.StartMultiClient()

	// Main command loop
	cmdCtx.RunInGroup(func() error {
		// Send initial ping
		cmdCtx.GetMultiClient().SendPing()

		for {
			t := time.NewTimer(5 * time.Second)

			select {
			case <-cmdCtx.Context().Done():
				t.Stop()
				return cmdCtx.Context().Err()

			case <-t.C:
				cmdCtx.GetMultiClient().SendPing()

			case msg := <-cmdCtx.Channels.ReceivedCh:
				t.Stop()
				if msg.Header != nil {
					log.Info().Str("from", msg.RemoteAddress.String()).
						Str("pkg", msg.Client.Name()).
						Str("type", msg.Header.MessageType.String()).
						Str("status", msg.Header.Status.String()).
						Msg("received message")
				} else {
					log.Debug().Str("from", msg.RemoteAddress.String()).
						Str("pkg", msg.Client.Name()).
						Msg("received unknown message")
				}

			case msg := <-cmdCtx.Channels.DiscoveryCh:
				t.Stop()
				log.Debug().Str("addr", msg.GetAddress()).Msg("discovery message")
				if newClient, err := cmdCtx.HandleDiscoveryMessage(msg); err != nil {
					return err
				} else if newClient != nil {
					// Send ping immediately to newly discovered client
					newClient.SendPing()
				}
			}
		}
	})

	// Wait for completion
	return cmdCtx.Wait()
}

// RunIntoGlazeProcessor implements the GlazeCommand interface for structured output
func (c *PingCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Initialize logging
	if err := glazed.InitLogging(parsedLayers); err != nil {
		return err
	}

	// Create command context from parsed layers
	cmdCtx, err := glazed.CreateCommandContextFromParsedLayers(ctx, parsedLayers)
	if err != nil {
		return err
	}
	defer cmdCtx.Cancel()

	// Setup multiclient
	if err := cmdCtx.SetupMultiClient("ping"); err != nil {
		log.Error().Err(err).Msg("Failed to setup multiclient")
		return err
	}

	// Setup discovery if enabled
	cmdCtx.SetupDiscovery()

	// Start multiclient
	cmdCtx.StartMultiClient()

	// Main command loop with structured output
	cmdCtx.RunInGroup(func() error {
		// Send initial ping
		cmdCtx.GetMultiClient().SendPing()

		for {
			t := time.NewTimer(5 * time.Second)

			select {
			case <-cmdCtx.Context().Done():
				t.Stop()
				return cmdCtx.Context().Err()

			case <-t.C:
				cmdCtx.GetMultiClient().SendPing()

			case msg := <-cmdCtx.Channels.ReceivedCh:
				t.Stop()
				if msg.Header != nil {
					// Emit structured data row
					row := types.NewRow(
						types.MRP("timestamp", time.Now()),
						types.MRP("from", msg.RemoteAddress.String()),
						types.MRP("client", msg.Client.Name()),
						types.MRP("type", msg.Header.MessageType.String()),
						types.MRP("status", msg.Header.Status.String()),
						types.MRP("event", "message_received"),
					)
					if err := gp.AddRow(ctx, row); err != nil {
						return err
					}

					// Also log for debugging
					log.Debug().Str("from", msg.RemoteAddress.String()).
						Str("pkg", msg.Client.Name()).
						Str("type", msg.Header.MessageType.String()).
						Str("status", msg.Header.Status.String()).
						Msg("received message")
				} else {
					// Emit structured data for unknown messages
					row := types.NewRow(
						types.MRP("timestamp", time.Now()),
						types.MRP("from", msg.RemoteAddress.String()),
						types.MRP("client", msg.Client.Name()),
						types.MRP("event", "unknown_message"),
					)
					if err := gp.AddRow(ctx, row); err != nil {
						return err
					}

					log.Debug().Str("from", msg.RemoteAddress.String()).
						Str("pkg", msg.Client.Name()).
						Msg("received unknown message")
				}

			case msg := <-cmdCtx.Channels.DiscoveryCh:
				t.Stop()
				// Emit structured data for discovery
				row := types.NewRow(
					types.MRP("timestamp", time.Now()),
					types.MRP("address", msg.GetAddress()),
					types.MRP("event", "discovery_message"),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}

				log.Debug().Str("addr", msg.GetAddress()).Msg("discovery message")
				if newClient, err := cmdCtx.HandleDiscoveryMessage(msg); err != nil {
					return err
				} else if newClient != nil {
					// Send ping immediately to newly discovered client
					newClient.SendPing()
				}
			}
		}
	})

	// Wait for completion
	return cmdCtx.Wait()
}

// NewPingCommand creates a new ping command with glazed support
func NewPingCommand() (*PingCommand, error) {
	// Get standard layers
	layers, err := glazed.NewStandardLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"ping",
		cmds.WithShort("Send ping messages to one or multiple PPA servers"),
		cmds.WithLong("Send periodic ping messages to PPA servers and display responses. "+
			"Supports both human-readable logging output and structured data output."),
		cmds.WithLayersList(layers...),
	)

	return &PingCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure interface compliance
var _ cmds.BareCommand = &PingCommand{}
var _ cmds.GlazeCommand = &PingCommand{}
