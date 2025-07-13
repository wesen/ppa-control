package cmds

import (
	"context"
	"fmt"
	"time"

	"ppa-control/lib/glazed"
	"ppa-control/lib/simulation"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

// SimulateGlazedCommand handles PPA simulation with glazed framework support.
// It supports classic (logging) and structured Glazed output.
type SimulateGlazedCommand struct {
	*cmds.CommandDescription
}

// SimulateSettings represents simulation-specific parameters
type SimulateSettings struct {
	Interface    string `glazed.parameter:"interface"`
	ListenAddr   string `glazed.parameter:"listen-address"`
	ListenPort   uint   `glazed.parameter:"listen-port"`
}

// Run implements the BareCommand interface for classic text output
func (c *SimulateGlazedCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Initialize logging
	if err := glazed.InitLogging(parsedLayers); err != nil {
		return err
	}

	// Get simulation settings
	settings, err := c.getSimulationSettings(parsedLayers)
	if err != nil {
		return err
	}

	// Display startup information
	serverString := fmt.Sprintf("%s:%d", settings.Address, settings.Port)
	log.Info().
		Str("address", serverString).
		Str("interface", settings.Interface).
		Str("name", settings.Name).
		Msg("Starting simulated PPA device")

	// Run the simulation
	return c.runSimulation(ctx, settings)
}

// RunIntoGlazeProcessor implements the GlazeCommand interface for structured output
func (c *SimulateGlazedCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Initialize logging
	if err := glazed.InitLogging(parsedLayers); err != nil {
		return err
	}

	// Get simulation settings
	settings, err := c.getSimulationSettings(parsedLayers)
	if err != nil {
		return err
	}

	// Emit device creation event
	serverString := fmt.Sprintf("%s:%d", settings.Address, settings.Port)
	row := types.NewRow(
		types.MRP("timestamp", time.Now()),
		types.MRP("event", "device_created"),
		types.MRP("name", settings.Name),
		types.MRP("address", serverString),
		types.MRP("interface", settings.Interface),
		types.MRP("unique_id", fmt.Sprintf("%02x%02x%02x%02x", 
			settings.UniqueId[0], settings.UniqueId[1], 
			settings.UniqueId[2], settings.UniqueId[3])),
		types.MRP("component_id", settings.ComponentId),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}

	// Run simulation with structured output capture
	return c.runSimulationWithStructuredOutput(ctx, settings, gp)
}

// getSimulationSettings extracts simulation parameters from parsed layers
func (c *SimulateGlazedCommand) getSimulationSettings(parsedLayers *layers.ParsedLayers) (simulation.SimulatedDeviceSettings, error) {
	// Extract simulation-specific settings
	simulateSettings := &SimulateSettings{}
	if err := parsedLayers.InitializeStruct("simulate", simulateSettings); err != nil {
		return simulation.SimulatedDeviceSettings{}, err
	}

	return simulation.SimulatedDeviceSettings{
		UniqueId:    [4]byte{0, 1, 2, 3},
		ComponentId: 0xff,
		Name:        "simulated",
		Address:     simulateSettings.ListenAddr,
		Port:        uint16(simulateSettings.ListenPort),
		Interface:   simulateSettings.Interface,
	}, nil
}

// runSimulation runs the simulation in classic mode
func (c *SimulateGlazedCommand) runSimulation(ctx context.Context, settings simulation.SimulatedDeviceSettings) error {
	grp, ctx := errgroup.WithContext(ctx)

	client := simulation.NewSimulatedDevice(settings)
	grp.Go(func() error {
		return client.Run(ctx)
	})

	return grp.Wait()
}

// runSimulationWithStructuredOutput runs the simulation with structured output capture
func (c *SimulateGlazedCommand) runSimulationWithStructuredOutput(
	ctx context.Context, 
	settings simulation.SimulatedDeviceSettings, 
	gp middlewares.Processor,
) error {
	grp, ctx := errgroup.WithContext(ctx)

	// Create a custom simulation device that emits structured events
	client := simulation.NewSimulatedDevice(settings)
	
	// Emit startup event
	row := types.NewRow(
		types.MRP("timestamp", time.Now()),
		types.MRP("event", "device_started"),
		types.MRP("name", settings.Name),
		types.MRP("address", fmt.Sprintf("%s:%d", settings.Address, settings.Port)),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}

	// Run the device
	grp.Go(func() error {
		// Note: For structured output, we're currently limited by the simulation.SimulatedDevice
		// not having hooks for emitting structured events. This would require modifying
		// the simulation package to support callbacks or event channels.
		// For now, we run the simulation and emit basic lifecycle events.
		
		err := client.Run(ctx)
		
		// Emit shutdown event when device stops
		shutdownRow := types.NewRow(
			types.MRP("timestamp", time.Now()),
			types.MRP("event", "device_stopped"),
			types.MRP("name", settings.Name),
		)
		// Ignore error on shutdown as context might be cancelled
		_ = gp.AddRow(context.Background(), shutdownRow)
		
		return err
	})

	return grp.Wait()
}

// NewSimulateParameterLayer creates a parameter layer for simulation-specific configuration
func NewSimulateParameterLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"simulate",
		"Simulation Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"interface",
				parameters.ParameterTypeString,
				parameters.WithHelp("Bind listener to interface"),
				parameters.WithShortFlag("i"),
			),
			parameters.NewParameterDefinition(
				"listen-address",
				parameters.ParameterTypeString,
				parameters.WithDefault("localhost"),
				parameters.WithHelp("Address to listen on"),
			),
			parameters.NewParameterDefinition(
				"listen-port",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(uint(5001)),
				parameters.WithHelp("Port to listen on"),
			),
		),
	)
}

// NewSimulateGlazedCommand creates a new simulate command with glazed support
func NewSimulateGlazedCommand() (*SimulateGlazedCommand, error) {
	// Get standard layers
	standardLayers, err := glazed.NewStandardLayers()
	if err != nil {
		return nil, err
	}

	// Add simulation-specific layer
	simulateLayer, err := NewSimulateParameterLayer()
	if err != nil {
		return nil, err
	}

	layers := append(standardLayers, simulateLayer)

	cmdDesc := cmds.NewCommandDescription(
		"simulate-glazed",
		cmds.WithShort("Start a simulated PPA device with structured output support"),
		cmds.WithLong("Start a simulated PPA device that responds to PPA protocol messages. "+
			"Supports both human-readable logging output and structured data output for simulation events."),
		cmds.WithLayersList(layers...),
	)

	return &SimulateGlazedCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure interface compliance
var _ cmds.BareCommand = &SimulateGlazedCommand{}
var _ cmds.GlazeCommand = &SimulateGlazedCommand{}
