package cmds

import (
	"context"
	"time"

	"ppa-control/lib/glazed"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
)

// RecallGlazedCommand handles PPA recall messages with glazed framework support.
// It supports classic (human-readable) and structured Glazed output.
type RecallGlazedCommand struct {
	*cmds.CommandDescription
}

// RecallSettings represents recall-specific parameters
type RecallSettings struct {
	Preset int  `glazed.parameter:"preset"`
	Loop   bool `glazed.parameter:"loop"`
}

// Run implements the BareCommand interface for classic text output
func (c *RecallGlazedCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Initialize logging
	if err := glazed.InitLogging(parsedLayers); err != nil {
		return err
	}

	// Extract recall settings
	recallSettings := &RecallSettings{}
	if err := parsedLayers.InitializeStruct("recall", recallSettings); err != nil {
		return err
	}

	// Create command context from parsed layers
	cmdCtx, err := glazed.CreateCommandContextFromParsedLayers(ctx, parsedLayers)
	if err != nil {
		return err
	}
	defer cmdCtx.Cancel()

	// Setup multiclient
	if err := cmdCtx.SetupMultiClient("recall"); err != nil {
		log.Fatal().Err(err).Msg("Failed to setup multiclient")
		return err
	}

	// Setup discovery if enabled
	cmdCtx.SetupDiscovery()

	// Start multiclient
	cmdCtx.StartMultiClient()

	// Main command loop
	cmdCtx.RunInGroup(func() error {
		// Send initial recall
		cmdCtx.GetMultiClient().SendPresetRecallByPresetIndex(recallSettings.Preset)

		// If not looping, just wait for context cancellation
		if !recallSettings.Loop {
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
				cmdCtx.GetMultiClient().SendPresetRecallByPresetIndex(recallSettings.Preset)

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
					// Send recall immediately to newly discovered client
					newClient.SendPresetRecallByPresetIndex(recallSettings.Preset)
				}
			}
		}
	})

	// Wait for completion
	return cmdCtx.Wait()
}

// RunIntoGlazeProcessor implements the GlazeCommand interface for structured output
func (c *RecallGlazedCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Initialize logging
	if err := glazed.InitLogging(parsedLayers); err != nil {
		return err
	}

	// Extract recall settings
	recallSettings := &RecallSettings{}
	if err := parsedLayers.InitializeStruct("recall", recallSettings); err != nil {
		return err
	}

	// Create command context from parsed layers
	cmdCtx, err := glazed.CreateCommandContextFromParsedLayers(ctx, parsedLayers)
	if err != nil {
		return err
	}
	defer cmdCtx.Cancel()

	// Setup multiclient
	if err := cmdCtx.SetupMultiClient("recall"); err != nil {
		log.Error().Err(err).Msg("Failed to setup multiclient")
		return err
	}

	// Setup discovery if enabled
	cmdCtx.SetupDiscovery()

	// Start multiclient
	cmdCtx.StartMultiClient()

	// Main command loop with structured output
	cmdCtx.RunInGroup(func() error {
		// Send initial recall and emit structured data
		cmdCtx.GetMultiClient().SendPresetRecallByPresetIndex(recallSettings.Preset)

		// Emit recall command initiation
		row := types.NewRow(
			types.MRP("timestamp", time.Now()),
			types.MRP("preset_index", recallSettings.Preset),
			types.MRP("loop_enabled", recallSettings.Loop),
			types.MRP("event", "recall_initiated"),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}

		// If not looping, just wait for context cancellation
		if !recallSettings.Loop {
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
				cmdCtx.GetMultiClient().SendPresetRecallByPresetIndex(recallSettings.Preset)

				// Emit recall send event
				row := types.NewRow(
					types.MRP("timestamp", time.Now()),
					types.MRP("preset_index", recallSettings.Preset),
					types.MRP("event", "recall_sent"),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}

			case msg := <-cmdCtx.Channels.ReceivedCh:
				t.Stop()
				if msg.Header != nil {
					// Emit structured data row for recall response
					row := types.NewRow(
						types.MRP("timestamp", time.Now()),
						types.MRP("from", msg.RemoteAddress.String()),
						types.MRP("client", msg.Client.Name()),
						types.MRP("type", msg.Header.MessageType.String()),
						types.MRP("status", msg.Header.Status.String()),
						types.MRP("preset_index", recallSettings.Preset),
						types.MRP("event", "recall_response"),
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
						types.MRP("preset_index", recallSettings.Preset),
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
					types.MRP("preset_index", recallSettings.Preset),
					types.MRP("event", "discovery_message"),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}

				log.Debug().Str("addr", msg.GetAddress()).Msg("discovery message")
				if newClient, err := cmdCtx.HandleDiscoveryMessage(msg); err != nil {
					return err
				} else if newClient != nil {
					// Send recall immediately to newly discovered client
					newClient.SendPresetRecallByPresetIndex(recallSettings.Preset)

					// Emit new client recall event
					row := types.NewRow(
						types.MRP("timestamp", time.Now()),
						types.MRP("address", msg.GetAddress()),
						types.MRP("preset_index", recallSettings.Preset),
						types.MRP("event", "new_client_recall"),
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

// NewRecallParameterLayer creates a parameter layer for recall-specific parameters
func NewRecallParameterLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"recall",
		"Recall Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"preset",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(0),
				parameters.WithHelp("Preset index to recall"),
			),
			parameters.NewParameterDefinition(
				"loop",
				parameters.ParameterTypeBool,
				parameters.WithDefault(true),
				parameters.WithHelp("Send recalls in a loop"),
				parameters.WithShortFlag("l"),
			),
		),
	)
}

// NewRecallGlazedCommand creates a new recall command with glazed support
func NewRecallGlazedCommand() (*RecallGlazedCommand, error) {
	// Get standard layers
	standardLayers, err := glazed.NewStandardLayers()
	if err != nil {
		return nil, err
	}

	// Add recall-specific layer
	recallLayer, err := NewRecallParameterLayer()
	if err != nil {
		return nil, err
	}

	layers := append(standardLayers, recallLayer)

	cmdDesc := cmds.NewCommandDescription(
		"recall-glazed",
		cmds.WithShort("Recall presets on PPA servers with structured output support"),
		cmds.WithLong("Send preset recall messages to PPA servers by preset index. "+
			"Supports both human-readable logging output and structured data output."),
		cmds.WithLayersList(layers...),
	)

	return &RecallGlazedCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure interface compliance
var _ cmds.BareCommand = &RecallGlazedCommand{}
var _ cmds.GlazeCommand = &RecallGlazedCommand{}
