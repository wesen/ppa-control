package cmds

import (
	"context"
	"fmt"
	"time"

	"ppa-control/lib/glazed"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
)

// VolumeCommand handles PPA volume commands with glazed framework support.
// It supports structured output only (implements GlazeCommand interface).
type VolumeCommand struct {
	*cmds.CommandDescription
}

// VolumeSettings represents volume-specific parameters
type VolumeSettings struct {
	Volume float32 `glazed.parameter:"volume"`
	Loop   bool    `glazed.parameter:"loop"`
}

// Run implements the BareCommand interface for classic text output
func (c *VolumeCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Initialize logging
	if err := glazed.InitLogging(parsedLayers); err != nil {
		return err
	}

	// Extract volume-specific settings
	volumeSettings := &VolumeSettings{}
	if err := parsedLayers.InitializeStruct("volume", volumeSettings); err != nil {
		return err
	}

	// Validate volume range
	if volumeSettings.Volume < 0 || volumeSettings.Volume > 1 {
		return fmt.Errorf("volume must be between 0 and 1")
	}

	// Create command context from parsed layers
	cmdCtx, err := glazed.CreateCommandContextFromParsedLayers(ctx, parsedLayers)
	if err != nil {
		return err
	}
	defer cmdCtx.Cancel()

	// Setup multiclient
	if err := cmdCtx.SetupMultiClient("volume"); err != nil {
		log.Fatal().Err(err).Msg("Failed to setup multiclient")
		return err
	}

	// Setup discovery if enabled
	cmdCtx.SetupDiscovery()

	// Start multiclient
	cmdCtx.StartMultiClient()

	// Main command loop
	cmdCtx.RunInGroup(func() error {
		// Send initial volume
		cmdCtx.GetMultiClient().SendMasterVolume(volumeSettings.Volume)

		// If not looping, just wait for context cancellation
		if !volumeSettings.Loop {
			<-cmdCtx.Context().Done()
			return cmdCtx.Context().Err()
		}

		for {
			t := time.NewTimer(5 * time.Second)

			select {
			case <-cmdCtx.Context().Done():
				t.Stop()
				return cmdCtx.Context().Err()

			case <-t.C:
				cmdCtx.GetMultiClient().SendMasterVolume(volumeSettings.Volume)

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
					// Send volume immediately to newly discovered client
					newClient.SendMasterVolume(volumeSettings.Volume)
				}
			}
		}
	})

	// Wait for completion
	return cmdCtx.Wait()
}

// RunIntoGlazeProcessor implements the GlazeCommand interface for structured output
func (c *VolumeCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Initialize logging
	if err := glazed.InitLogging(parsedLayers); err != nil {
		return err
	}

	// Extract volume-specific settings
	volumeSettings := &VolumeSettings{}
	if err := parsedLayers.InitializeStruct("volume", volumeSettings); err != nil {
		return err
	}

	// Validate volume range
	if volumeSettings.Volume < 0 || volumeSettings.Volume > 1 {
		return fmt.Errorf("volume must be between 0 and 1")
	}

	// Create command context from parsed layers
	cmdCtx, err := glazed.CreateCommandContextFromParsedLayers(ctx, parsedLayers)
	if err != nil {
		return err
	}
	defer cmdCtx.Cancel()

	// Setup multiclient
	if err := cmdCtx.SetupMultiClient("volume"); err != nil {
		log.Error().Err(err).Msg("Failed to setup multiclient")
		return err
	}

	// Setup discovery if enabled
	cmdCtx.SetupDiscovery()

	// Start multiclient
	cmdCtx.StartMultiClient()

	// Main command loop with structured output
	cmdCtx.RunInGroup(func() error {
		// Send initial volume and emit structured data
		cmdCtx.GetMultiClient().SendMasterVolume(volumeSettings.Volume)

		// Emit volume command sent event
		row := types.NewRow(
			types.MRP("timestamp", time.Now()),
			types.MRP("event", "volume_command_sent"),
			types.MRP("volume", volumeSettings.Volume),
			types.MRP("loop", volumeSettings.Loop),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}

		// If not looping, just wait for context cancellation
		if !volumeSettings.Loop {
			<-cmdCtx.Context().Done()
			return cmdCtx.Context().Err()
		}

		for {
			t := time.NewTimer(5 * time.Second)

			select {
			case <-cmdCtx.Context().Done():
				t.Stop()
				return cmdCtx.Context().Err()

			case <-t.C:
				cmdCtx.GetMultiClient().SendMasterVolume(volumeSettings.Volume)

				// Emit periodic volume command event
				row := types.NewRow(
					types.MRP("timestamp", time.Now()),
					types.MRP("event", "volume_command_sent"),
					types.MRP("volume", volumeSettings.Volume),
					types.MRP("loop", volumeSettings.Loop),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}

			case msg := <-cmdCtx.Channels.ReceivedCh:
				t.Stop()
				if msg.Header != nil {
					// Emit structured data row for received messages
					row := types.NewRow(
						types.MRP("timestamp", time.Now()),
						types.MRP("from", msg.RemoteAddress.String()),
						types.MRP("client", msg.Client.Name()),
						types.MRP("type", msg.Header.MessageType.String()),
						types.MRP("status", msg.Header.Status.String()),
						types.MRP("event", "message_received"),
						types.MRP("volume", volumeSettings.Volume),
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
						types.MRP("volume", volumeSettings.Volume),
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
					types.MRP("volume", volumeSettings.Volume),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}

				log.Debug().Str("addr", msg.GetAddress()).Msg("discovery message")
				if newClient, err := cmdCtx.HandleDiscoveryMessage(msg); err != nil {
					return err
				} else if newClient != nil {
					// Send volume immediately to newly discovered client
					newClient.SendMasterVolume(volumeSettings.Volume)

					// Emit event for new client volume command
					row := types.NewRow(
						types.MRP("timestamp", time.Now()),
						types.MRP("event", "new_client_volume_sent"),
						types.MRP("client_address", msg.GetAddress()),
						types.MRP("volume", volumeSettings.Volume),
					)
					if err := gp.AddRow(ctx, row); err != nil {
						return err
					}
				}
			}
		}
	})

	// Wait for completion
	return cmdCtx.Wait()
}

// NewVolumeParameterLayer creates a parameter layer for volume-specific configuration
func NewVolumeParameterLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"volume",
		"Volume Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"volume",
				parameters.ParameterTypeFloat,
				parameters.WithDefault(0.5),
				parameters.WithHelp("Volume level (0.0-1.0)"),
				parameters.WithShortFlag("v"),
			),
			parameters.NewParameterDefinition(
				"loop",
				parameters.ParameterTypeBool,
				parameters.WithDefault(true),
				parameters.WithHelp("Send volume commands in a loop"),
				parameters.WithShortFlag("l"),
			),
		),
	)
}

// NewVolumeCommand creates a new volume command with glazed support
func NewVolumeCommand() (*VolumeCommand, error) {
	// Get standard layers
	standardLayers, err := glazed.NewStandardLayers()
	if err != nil {
		return nil, err
	}

	// Add volume-specific layer
	volumeLayer, err := NewVolumeParameterLayer()
	if err != nil {
		return nil, err
	}

	layers := append(standardLayers, volumeLayer)

	cmdDesc := cmds.NewCommandDescription(
		"volume",
		cmds.WithShort("Set the volume of one or more clients with structured output support"),
		cmds.WithLong("Set the volume of one or more PPA clients. "+
			"Outputs structured data about volume operations including target addresses, "+
			"volume levels, success/failure status, and discovery events."),
		cmds.WithLayersList(layers...),
	)

	return &VolumeCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure interface compliance
var _ cmds.BareCommand = &VolumeCommand{}
var _ cmds.GlazeCommand = &VolumeCommand{}
