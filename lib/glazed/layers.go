package glazed

import (
	"context"
	"ppa-control/lib"
	"ppa-control/lib/client"
	"ppa-control/lib/client/discovery"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/settings"
)

const (
	// PPALayerSlug is the identifier for the PPA parameter layer
	PPALayerSlug = "ppa"
)

// PPASettings represents the common PPA configuration parameters
type PPASettings struct {
	Addresses         string   `glazed.parameter:"addresses"`
	Discover          bool     `glazed.parameter:"discover"`
	ComponentID       uint     `glazed.parameter:"component-id"`
	ComponentIDLegacy uint     `glazed.parameter:"componentId"` // Legacy compatibility
	Port              uint     `glazed.parameter:"port"`
	Interfaces        []string `glazed.parameter:"interfaces"`
}

// NewPPAParameterLayer creates a parameter layer for common PPA configuration
func NewPPAParameterLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		PPALayerSlug,
		"PPA Connection Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"addresses",
				parameters.ParameterTypeString,
				parameters.WithHelp("Addresses to connect to, comma separated"),
				parameters.WithShortFlag("a"),
			),
			parameters.NewParameterDefinition(
				"discover",
				parameters.ParameterTypeBool,
				parameters.WithDefault(true), // Changed to match legacy behavior
				parameters.WithHelp("Send broadcast discovery messages"),
				parameters.WithShortFlag("d"),
			),
			parameters.NewParameterDefinition(
				"component-id",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(0xFF),
				parameters.WithHelp("Component ID to use for devices"),
				parameters.WithShortFlag("c"),
			),
			// Legacy compatibility for camelCase componentId
			parameters.NewParameterDefinition(
				"componentId",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(0xFF),
				parameters.WithHelp("Component ID to use for devices (legacy compatibility)"),
			),
			parameters.NewParameterDefinition(
				"port",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(5001),
				parameters.WithHelp("Port to connect to"),
				parameters.WithShortFlag("p"),
			),
			parameters.NewParameterDefinition(
				"interfaces",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Interfaces to use for discovery"),
			),
		),
	)
}

// CreateCommandContextFromParsedLayers creates a lib.CommandContext from parsed layers
func CreateCommandContextFromParsedLayers(ctx context.Context, parsedLayers *layers.ParsedLayers) (*lib.CommandContext, error) {
	// Extract PPA settings
	ppaSettings := &PPASettings{}
	if err := parsedLayers.InitializeStruct(PPALayerSlug, ppaSettings); err != nil {
		return nil, err
	}

	// Determine ComponentID: use legacy if main is default and legacy is set
	componentID := ppaSettings.ComponentID
	if componentID == 0xFF && ppaSettings.ComponentIDLegacy != 0xFF {
		componentID = ppaSettings.ComponentIDLegacy
	}

	// Convert to lib.CommandConfig
	config := &lib.CommandConfig{
		Addresses:   ppaSettings.Addresses,
		Discovery:   ppaSettings.Discover,
		ComponentID: componentID,
		Port:        ppaSettings.Port,
		Interfaces:  ppaSettings.Interfaces,
	}

	// Create channels
	channels := &lib.CommandChannels{
		DiscoveryCh: make(chan discovery.PeerInformation),
		ReceivedCh:  make(chan client.ReceivedMessage),
	}

	// Create command context
	cmdCtx := &lib.CommandContext{
		Config:   config,
		Channels: channels,
	}

	// Setup context with cancellation
	cmdCtx.SetupContext(ctx, nil)

	return cmdCtx, nil
}

// InitLogging initializes logging from parsed layers using glazed logging
func InitLogging(parsedLayers *layers.ParsedLayers) error {
	// Extract logging settings
	var loggingSettings logging.LoggingSettings
	if err := parsedLayers.InitializeStruct(logging.LoggingLayerSlug, &loggingSettings); err != nil {
		return err
	}

	// Initialize logger from settings
	return logging.InitLoggerFromSettings(&loggingSettings)
}

// NewStandardLayers creates the standard layer set for PPA commands
func NewStandardLayers() ([]layers.ParameterLayer, error) {
	// Create PPA layer
	ppaLayer, err := NewPPAParameterLayer()
	if err != nil {
		return nil, err
	}

	// Create logging layer
	loggingLayer, err := logging.NewLoggingLayer()
	if err != nil {
		return nil, err
	}

	// Create glazed layer for structured output
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return []layers.ParameterLayer{ppaLayer, loggingLayer, glazedLayer}, nil
}
