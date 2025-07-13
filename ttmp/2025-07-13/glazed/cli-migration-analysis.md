# PPA-Control CLI Migration to Glazed Framework - Comprehensive Analysis

## Executive Summary

This document provides a comprehensive technical analysis for migrating ppa-control's CLI commands from the current Cobra-based implementation to the Glazed framework. The migration will provide structured data output, better parameter management, and improved composability while maintaining backward compatibility.

**Key findings:**
- 6 CLI commands need migration: ping, volume, recall, simulate, udp-broadcast, and root
- Current architecture uses custom CommandContext wrapper around Cobra
- High compatibility with Glazed patterns - commands already follow separation of concerns
- Recommended approach: Dual Commands for gradual migration with backward compatibility

## Current CLI Architecture Analysis

### Command Structure Overview

The current ppa-cli implementation consists of:

```
cmd/ppa-cli/
├── main.go                    # Entry point
├── cmds/
│   ├── root.go               # Root command with global flags
│   ├── ping.go               # Network ping functionality  
│   ├── volume.go             # Audio volume control
│   ├── recall.go             # Preset recall functionality
│   ├── simulate.go           # Device simulation
│   └── udpBroadcast.go       # UDP networking utility
```

### Current Command Inventory

| Command | Purpose | Primary Function | Output Type | Complexity |
|---------|---------|------------------|-------------|------------|
| **root** | Base command with global flags | Setup, logging, profiling | None | Low |
| **ping** | Network connectivity testing | Send/receive PPA ping messages | Log messages | Medium |
| **volume** | Audio control | Set master volume on devices | Log messages | Medium |
| **recall** | Preset management | Recall presets by index | Log messages | Medium |
| **simulate** | Device simulation | Run simulated PPA device | Status messages | Medium |
| **udp-broadcast** | Network testing | Raw UDP broadcast utility | Debug output | Low |

### Current Parameter Patterns

All commands follow consistent parameter patterns:

**Common Parameters:**
- `--addresses, -a`: Target device addresses (comma-separated)
- `--discover, -d`: Enable/disable device discovery  
- `--componentId, -c`: Component ID for devices (default: 0xFF)
- `--port, -p`: Network port (default: 5001)
- `--interfaces`: Network interfaces for discovery

**Global Parameters:**
- `--log-level`: Logging verbosity (debug, info, warn, error, fatal)
- `--log-format`: Output format (json, text)
- `--with-caller`: Include caller information in logs
- `--dump-mem-profile`: Memory profiling file path
- `--track-leaks`: Enable memory/goroutine leak tracking

**Command-Specific Parameters:**
- `volume`: `--volume, -v` (float32), `--loop, -l` (bool)
- `recall`: `--preset` (int), `--loop, -l` (bool)
- `simulate`: `--interface, -i`, `--address, -a`, `--port, -p`
- `udp-broadcast`: `--server, -s` (bool), `--interface, -i`

### Current Architecture Patterns

#### CommandContext Abstraction
The implementation uses a sophisticated `CommandContext` wrapper that provides:

```go
type CommandContext struct {
    Config      *CommandConfig     // Parameter storage
    Channels    *CommandChannels   // Message channels
    ctx         context.Context    // Cancellation context
    cancelFunc  context.CancelFunc // Cancel function
    group       *errgroup.Group    // Goroutine management
    multiClient *client.MultiClient // PPA client management
}
```

**Strengths:**
- Consistent parameter extraction across commands
- Built-in context cancellation and signal handling
- Structured goroutine management with errgroup
- Centralized MultiClient lifecycle management
- Separation of concerns between parameter handling and business logic

#### Command Execution Pattern
All network commands follow this pattern:

```go
func (cmd *Command) Run(cmd *cobra.Command, args []string) {
    // 1. Setup command context
    cmdCtx := lib.SetupCommand(cmd)
    defer cmdCtx.Cancel()

    // 2. Setup multiclient  
    cmdCtx.SetupMultiClient("command-name")
    
    // 3. Setup discovery (if enabled)
    cmdCtx.SetupDiscovery()
    
    // 4. Start multiclient
    cmdCtx.StartMultiClient()
    
    // 5. Main command loop with select/channels
    cmdCtx.RunInGroup(func() error {
        for {
            select {
            case <-cmdCtx.Context().Done():
                return cmdCtx.Context().Err()
            case msg := <-cmdCtx.Channels.ReceivedCh:
                // Handle received messages
            case msg := <-cmdCtx.Channels.DiscoveryCh:
                // Handle discovery messages  
            case <-timer.C:
                // Periodic actions
            }
        }
    })
    
    // 6. Wait for completion
    cmdCtx.Wait()
}
```

### Data Structures and Output

#### Current Output Patterns
Commands primarily output structured log messages using zerolog:

```go
log.Info().Str("from", msg.RemoteAddress.String()).
    Str("pkg", msg.Client.Name()).
    Str("type", msg.Header.MessageType.String()).
    Str("status", msg.Header.Status.String()).
    Msg("received message")
```

#### Data Structures Available for Glazed Migration
The codebase has rich data structures that can be exposed via Glazed:

**Message Headers:**
```go
type Header struct {
    MessageType MessageType
    Status      Status
    // ... additional fields
}
```

**Device Information:**
```go
type Client interface {
    Name() string
    Address() string
    // ... methods
}
```

**Discovery Information:**
```go
type PeerInformation interface {
    GetAddress() string
    GetInterface() string
}
```

## Glazed Framework Integration Analysis

### Glazed Command Types Suitable for ppa-control

Based on the analysis of current commands and Glazed capabilities:

| Command | Recommended Type | Rationale |
|---------|------------------|-----------|
| **ping** | Dual Command | Needs both human-readable status and structured data for monitoring |
| **volume** | Dual Command | Interactive use requires feedback, automation needs structured output |
| **recall** | Dual Command | Same dual requirements as volume |
| **simulate** | Bare Command | Primarily status/debug output, less need for structured data |
| **udp-broadcast** | Bare Command | Low-level debugging tool, structured output not critical |

### Parameter Migration Strategy

#### Global Parameters → Glazed Layers
Current global parameters map well to Glazed parameter layers:

```go
// Current: Root command persistent flags
rootCmd.PersistentFlags().String("log-level", "debug", "Log level")
rootCmd.PersistentFlags().String("log-format", "text", "Log format")

// Glazed: Logging layer
loggingLayer := layers.NewParameterLayer(
    "logging",
    "Logging configuration",
    layers.WithParameterDefinitions(
        parameters.NewParameterDefinition(
            "log-level",
            parameters.ParameterTypeChoice,
            parameters.WithChoices("debug", "info", "warn", "error", "fatal"),
            parameters.WithDefault("debug"),
            parameters.WithHelp("Logging verbosity level"),
        ),
        parameters.NewParameterDefinition(
            "log-format", 
            parameters.ParameterTypeChoice,
            parameters.WithChoices("json", "text"),
            parameters.WithDefault("text"),
            parameters.WithHelp("Log output format"),
        ),
    ),
)
```

#### Common PPA Parameters → Custom Layer
PPA-specific parameters can be grouped into a reusable layer:

```go
ppaLayer := layers.NewParameterLayer(
    "ppa",
    "PPA protocol configuration",
    layers.WithParameterDefinitions(
        parameters.NewParameterDefinition(
            "addresses",
            parameters.ParameterTypeStringList,
            parameters.WithHelp("Target device addresses"),
            parameters.WithShortFlag("a"),
        ),
        parameters.NewParameterDefinition(
            "discover",
            parameters.ParameterTypeBool,
            parameters.WithDefault(true),
            parameters.WithHelp("Enable device discovery"),
            parameters.WithShortFlag("d"),
        ),
        parameters.NewParameterDefinition(
            "component-id",
            parameters.ParameterTypeInteger,
            parameters.WithDefault(0xFF),
            parameters.WithHelp("Component ID for devices"),
            parameters.WithShortFlag("c"),
        ),
        parameters.NewParameterDefinition(
            "port",
            parameters.ParameterTypeInteger,
            parameters.WithDefault(5001),
            parameters.WithHelp("Network port"),
            parameters.WithShortFlag("p"),
        ),
        parameters.NewParameterDefinition(
            "interfaces",
            parameters.ParameterTypeStringList,
            parameters.WithHelp("Network interfaces for discovery"),
        ),
    ),
)
```

## Detailed Migration Plan

### Phase 1: Foundation Setup (Estimated: 1-2 days)

#### 1.1 Create Glazed Integration Package
Create `ppa-control/lib/glazed/` to house Glazed integration:

```
lib/glazed/
├── layers.go          # Custom parameter layers
├── settings.go        # Settings structs  
├── context.go         # Glazed-CommandContext bridge
└── commands/          # Migrated commands
    ├── ping.go
    ├── volume.go
    ├── recall.go
    ├── simulate.go
    └── udp_broadcast.go
```

#### 1.2 Parameter Layer Definitions
**File: `lib/glazed/layers.go`**

```go
package glazed

import (
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/settings"
)

// CreatePPAParameterLayer creates reusable PPA-specific parameters
func CreatePPAParameterLayer() (*layers.ParameterLayerImpl, error) {
    return layers.NewParameterLayer(
        "ppa",
        "PPA protocol configuration",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "addresses",
                parameters.ParameterTypeStringList,
                parameters.WithHelp("Target device addresses (comma-separated)"),
                parameters.WithShortFlag("a"),
            ),
            parameters.NewParameterDefinition(
                "discover",
                parameters.ParameterTypeBool,
                parameters.WithDefault(true),
                parameters.WithHelp("Enable broadcast discovery"),
                parameters.WithShortFlag("d"),
            ),
            parameters.NewParameterDefinition(
                "component-id",
                parameters.ParameterTypeInteger,
                parameters.WithDefault(0xFF),
                parameters.WithHelp("Component ID for devices"),
                parameters.WithShortFlag("c"),
            ),
            parameters.NewParameterDefinition(
                "port",
                parameters.ParameterTypeInteger,
                parameters.WithDefault(5001),
                parameters.WithHelp("Network port"),
                parameters.WithShortFlag("p"),
            ),
            parameters.NewParameterDefinition(
                "interfaces",
                parameters.ParameterTypeStringList,
                parameters.WithHelp("Network interfaces for discovery"),
            ),
        ),
    ), nil
}

// CreateLoggingParameterLayer creates logging-specific parameters
func CreateLoggingParameterLayer() (*layers.ParameterLayerImpl, error) {
    return layers.NewParameterLayer(
        "logging",
        "Logging configuration",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "log-level",
                parameters.ParameterTypeChoice,
                parameters.WithChoices("debug", "info", "warn", "error", "fatal"),
                parameters.WithDefault("debug"),
                parameters.WithHelp("Logging verbosity level"),
            ),
            parameters.NewParameterDefinition(
                "log-format",
                parameters.ParameterTypeChoice,
                parameters.WithChoices("json", "text"),
                parameters.WithDefault("text"),
                parameters.WithHelp("Log output format"),
            ),
            parameters.NewParameterDefinition(
                "with-caller",
                parameters.ParameterTypeBool,
                parameters.WithDefault(false),
                parameters.WithHelp("Include caller information in logs"),
            ),
            parameters.NewParameterDefinition(
                "dump-mem-profile",
                parameters.ParameterTypeString,
                parameters.WithHelp("File path for memory profile dump"),
            ),
            parameters.NewParameterDefinition(
                "track-leaks",
                parameters.ParameterTypeBool,
                parameters.WithDefault(false),
                parameters.WithHelp("Track memory and goroutine leaks"),
            ),
        ),
    ), nil
}
```

#### 1.3 Settings Structs
**File: `lib/glazed/settings.go`**

```go
package glazed

import (
    "time"
    "ppa-control/lib"
)

// PPASettings maps to PPA parameter layer
type PPASettings struct {
    Addresses   []string `glazed.parameter:"addresses"`
    Discover    bool     `glazed.parameter:"discover"`
    ComponentID int      `glazed.parameter:"component-id"`
    Port        int      `glazed.parameter:"port"`
    Interfaces  []string `glazed.parameter:"interfaces"`
}

// LoggingSettings maps to logging parameter layer
type LoggingSettings struct {
    LogLevel        string `glazed.parameter:"log-level"`
    LogFormat       string `glazed.parameter:"log-format"`
    WithCaller      bool   `glazed.parameter:"with-caller"`
    DumpMemProfile  string `glazed.parameter:"dump-mem-profile"`
    TrackLeaks      bool   `glazed.parameter:"track-leaks"`
}

// PingSettings for ping command
type PingSettings struct {
    PPASettings     `glazed.parameter:",squash"`
    LoggingSettings `glazed.parameter:",squash"`
}

// VolumeSettings for volume command
type VolumeSettings struct {
    PPASettings     `glazed.parameter:",squash"`
    LoggingSettings `glazed.parameter:",squash"`
    Volume          float32 `glazed.parameter:"volume"`
    Loop            bool    `glazed.parameter:"loop"`
}

// RecallSettings for recall command
type RecallSettings struct {
    PPASettings     `glazed.parameter:",squash"`
    LoggingSettings `glazed.parameter:",squash"`
    Preset          int  `glazed.parameter:"preset"`
    Loop            bool `glazed.parameter:"loop"`
}

// ToCommandConfig converts settings to current CommandConfig
func (s *PPASettings) ToCommandConfig() *lib.CommandConfig {
    return &lib.CommandConfig{
        Addresses:   strings.Join(s.Addresses, ","),
        Discovery:   s.Discover,
        ComponentID: uint(s.ComponentID),
        Port:        uint(s.Port),
        Interfaces:  s.Interfaces,
    }
}
```

#### 1.4 Context Bridge
**File: `lib/glazed/context.go`**

```go
package glazed

import (
    "context"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "ppa-control/lib"
)

// GlazedCommandContext bridges Glazed with existing CommandContext
type GlazedCommandContext struct {
    *lib.CommandContext
    parsedLayers *layers.ParsedLayers
}

// NewGlazedCommandContext creates a bridge between Glazed and existing context
func NewGlazedCommandContext(parsedLayers *layers.ParsedLayers) (*GlazedCommandContext, error) {
    // Extract PPA settings
    ppaSettings := &PPASettings{}
    if err := parsedLayers.InitializeStruct("ppa", ppaSettings); err != nil {
        return nil, err
    }

    // Extract logging settings  
    loggingSettings := &LoggingSettings{}
    if err := parsedLayers.InitializeStruct("logging", loggingSettings); err != nil {
        return nil, err
    }

    // Initialize logging based on settings
    setupLoggingFromSettings(loggingSettings)

    // Create traditional CommandContext with converted config
    cmdCtx := lib.SetupCommandWithConfig(ppaSettings.ToCommandConfig())

    return &GlazedCommandContext{
        CommandContext: cmdCtx,
        parsedLayers:   parsedLayers,
    }, nil
}

// GetSettings returns typed settings for a specific layer
func (gc *GlazedCommandContext) GetSettings(layer string, target interface{}) error {
    return gc.parsedLayers.InitializeStruct(layer, target)
}
```

### Phase 2: Dual Command Implementation (Estimated: 2-3 days)

#### 2.1 Ping Command Migration
**File: `lib/glazed/commands/ping.go`**

```go
package commands

import (
    "context"
    "fmt"
    "time"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/types"
    "github.com/rs/zerolog/log"
    glazedLib "ppa-control/lib/glazed"
)

type PingCommand struct {
    *cmds.CommandDescription
}

// Implement BareCommand for classic text output
func (c *PingCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    cmdCtx, err := glazedLib.NewGlazedCommandContext(parsedLayers)
    if err != nil {
        return err
    }
    defer cmdCtx.Cancel()

    // Setup multiclient
    if err := cmdCtx.SetupMultiClient("ping"); err != nil {
        log.Fatal().Err(err).Msg("Failed to setup multiclient")
        return err
    }

    cmdCtx.SetupDiscovery()
    cmdCtx.StartMultiClient()

    // Classic output: logs and status messages
    return c.runClassicMode(cmdCtx)
}

// Implement GlazeCommand for structured output
func (c *PingCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    cmdCtx, err := glazedLib.NewGlazedCommandContext(parsedLayers)
    if err != nil {
        return err
    }
    defer cmdCtx.Cancel()

    if err := cmdCtx.SetupMultiClient("ping"); err != nil {
        return fmt.Errorf("failed to setup multiclient: %w", err)
    }

    cmdCtx.SetupDiscovery()
    cmdCtx.StartMultiClient()

    // Structured output: data rows
    return c.runStructuredMode(cmdCtx, gp)
}

func (c *PingCommand) runClassicMode(cmdCtx *glazedLib.GlazedCommandContext) error {
    return cmdCtx.RunInGroup(func() error {
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
                    newClient.SendPing()
                }
            }
        }
    })
}

func (c *PingCommand) runStructuredMode(
    cmdCtx *glazedLib.GlazedCommandContext,
    gp middlewares.Processor,
) error {
    return cmdCtx.RunInGroup(func() error {
        cmdCtx.GetMultiClient().SendPing()

        for {
            t := time.NewTimer(5 * time.Second)

            select {
            case <-cmdCtx.Context().Done():
                t.Stop()
                return cmdCtx.Context().Err()

            case <-t.C:
                cmdCtx.GetMultiClient().SendPing()
                // Emit ping event
                row := types.NewRow(
                    types.MRP("timestamp", time.Now()),
                    types.MRP("event_type", "ping_sent"),
                    types.MRP("targets", cmdCtx.GetMultiClient().GetTargetAddresses()),
                )
                if err := gp.AddRow(cmdCtx.Context(), row); err != nil {
                    return err
                }

            case msg := <-cmdCtx.Channels.ReceivedCh:
                t.Stop()
                // Emit structured message data
                row := types.NewRow(
                    types.MRP("timestamp", time.Now()),
                    types.MRP("event_type", "message_received"),
                    types.MRP("from_address", msg.RemoteAddress.String()),
                    types.MRP("client_name", msg.Client.Name()),
                )

                if msg.Header != nil {
                    row.Set("message_type", msg.Header.MessageType.String())
                    row.Set("status", msg.Header.Status.String())
                } else {
                    row.Set("message_type", "unknown")
                    row.Set("status", "unknown")
                }

                if err := gp.AddRow(cmdCtx.Context(), row); err != nil {
                    return err
                }

            case msg := <-cmdCtx.Channels.DiscoveryCh:
                t.Stop()
                // Emit discovery data
                row := types.NewRow(
                    types.MRP("timestamp", time.Now()),
                    types.MRP("event_type", "device_discovered"),
                    types.MRP("address", msg.GetAddress()),
                    types.MRP("interface", msg.GetInterface()),
                )
                if err := gp.AddRow(cmdCtx.Context(), row); err != nil {
                    return err
                }

                if newClient, err := cmdCtx.HandleDiscoveryMessage(msg); err != nil {
                    return err
                } else if newClient != nil {
                    newClient.SendPing()
                }
            }
        }
    })
}

// NewPingCommand creates the glazed ping command
func NewPingCommand() (*PingCommand, error) {
    ppaLayer, err := glazedLib.CreatePPAParameterLayer()
    if err != nil {
        return nil, err
    }

    loggingLayer, err := glazedLib.CreateLoggingParameterLayer()
    if err != nil {
        return nil, err
    }

    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, err
    }

    cmdDesc := cmds.NewCommandDescription(
        "ping",
        cmds.WithShort("Send ping to one or multiple PPA servers"),
        cmds.WithLong("Continuously sends ping messages to PPA devices and reports responses. "+
            "Supports device discovery and provides both human-readable and structured output."),
        cmds.WithLayersList(ppaLayer, loggingLayer, glazedLayer),
    )

    return &PingCommand{
        CommandDescription: cmdDesc,
    }, nil
}

// Ensure interface compliance
var _ cmds.BareCommand = &PingCommand{}
var _ cmds.GlazeCommand = &PingCommand{}
```

#### 2.2 Volume Command Migration
Similar pattern to ping, with volume-specific parameters and business logic.

#### 2.3 Recall Command Migration  
Similar pattern with preset-specific parameters.

### Phase 3: CLI Integration (Estimated: 1 day)

#### 3.1 New Main Entry Point
**File: `cmd/ppa-cli-glazed/main.go`**

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/spf13/cobra"
    
    glazedCommands "ppa-control/lib/glazed/commands"
)

func main() {
    var rootCmd = &cobra.Command{
        Use:   "ppa-cli",
        Short: "ppa-cli is a command line interface for the PPA protocol",
    }

    // Create glazed commands
    pingCmd, err := glazedCommands.NewPingCommand()
    if err != nil {
        log.Fatalf("Error creating ping command: %v", err)
    }

    volumeCmd, err := glazedCommands.NewVolumeCommand()
    if err != nil {
        log.Fatalf("Error creating volume command: %v", err)
    }

    recallCmd, err := glazedCommands.NewRecallCommand()
    if err != nil {
        log.Fatalf("Error creating recall command: %v", err)
    }

    // Convert to Cobra commands with dual mode support
    pingCobraCmd, err := cli.BuildCobraCommandDualMode(
        pingCmd,
        cli.WithGlazeToggleFlag("structured-output"),
    )
    if err != nil {
        log.Fatalf("Error building ping command: %v", err)
    }

    volumeCobraCmd, err := cli.BuildCobraCommandDualMode(
        volumeCmd,
        cli.WithGlazeToggleFlag("structured-output"),
    )
    if err != nil {
        log.Fatalf("Error building volume command: %v", err)
    }

    recallCobraCmd, err := cli.BuildCobraCommandDualMode(
        recallCmd,
        cli.WithGlazeToggleFlag("structured-output"),
    )
    if err != nil {
        log.Fatalf("Error building recall command: %v", err)
    }

    // Add commands to root
    rootCmd.AddCommand(pingCobraCmd)
    rootCmd.AddCommand(volumeCobraCmd)
    rootCmd.AddCommand(recallCobraCmd)

    // Execute
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

## Technical Challenges and Solutions

### Challenge 1: CommandContext Integration

**Problem:** The existing `CommandContext` is tightly coupled to Cobra and may not integrate smoothly with Glazed's parameter system.

**Solution:** Create a bridge adapter (`GlazedCommandContext`) that:
- Converts Glazed parameters to `CommandConfig`
- Maintains backward compatibility with existing `CommandContext` API
- Preserves all existing functionality (signal handling, goroutine management, etc.)

### Challenge 2: Complex Channel-Based Logic

**Problem:** Commands use complex `select` statements with multiple channels that need to work in both classic and structured modes.

**Solution:** Extract common channel logic into shared methods:
```go
func (c *PingCommand) handleChannelMessages(
    cmdCtx *GlazedCommandContext,
    outputHandler func(event Event) error,
) error {
    // Shared select logic
    // outputHandler differs between classic (log) and structured (AddRow)
}
```

### Challenge 3: Backward Compatibility

**Problem:** Existing scripts and users depend on current CLI behavior.

**Solution:** Implement dual commands with careful mode selection:
- Default to classic mode for backward compatibility
- Use `--structured-output` flag to enable Glazed mode
- Maintain identical parameter names and behavior

### Challenge 4: Performance Impact

**Problem:** Glazed's structured output processing might add overhead to real-time commands.

**Solution:** 
- Benchmark both modes during testing
- Use buffered channels and batch processing where appropriate
- Implement streaming output for long-running commands

## Implementation Effort Estimation

### Development Phases

| Phase | Tasks | Estimated Time | Dependencies |
|-------|-------|----------------|--------------|
| **Phase 1: Foundation** | Parameter layers, settings, context bridge | 1-2 days | None |
| **Phase 2: Command Migration** | Ping, volume, recall dual commands | 2-3 days | Phase 1 |
| **Phase 3: CLI Integration** | New main entry, cobra integration | 1 day | Phase 2 |
| **Phase 4: Testing** | Compatibility tests, integration tests | 1-2 days | Phase 3 |
| **Phase 5: Documentation** | Updated docs, examples, migration guide | 0.5-1 day | Phase 4 |

**Total Estimated Time: 5.5-9 days**

### Risk Factors

**Low Risk:**
- Parameter migration (straightforward mapping)
- Basic command structure (well-understood patterns)
- Documentation generation (automatic)

**Medium Risk:**
- Channel handling complexity (requires careful testing)
- Performance considerations (needs benchmarking)
- Integration edge cases (thorough testing needed)

**High Risk:**
- User adoption (change management)
- Backward compatibility edge cases (extensive testing)

## Migration Strategy Options

### Option A: Big Bang Migration (Not Recommended)
Replace entire CLI at once.

**Pros:** Clean break, immediate benefits
**Cons:** High risk, potential disruption, difficult rollback

### Option B: Gradual Migration (Recommended)
Implement dual commands alongside existing CLI.

**Pros:** Low risk, gradual adoption, easy rollback
**Cons:** Maintenance of two systems temporarily

### Option C: Side-by-Side CLI (Alternative)
Create separate `ppa-cli-glazed` binary.

**Pros:** Zero disruption, experimentation friendly
**Cons:** User confusion, duplicate maintenance

## Recommended Implementation Approach

### Phase-by-Phase Migration (Option B)

#### Week 1: Foundation + Core Commands
1. Implement foundation (layers, settings, context bridge)
2. Migrate `ping` command as dual command
3. Test compatibility and functionality

#### Week 2: Complete Command Set
1. Migrate `volume` and `recall` commands
2. Implement comprehensive testing
3. Performance benchmarking

#### Week 3: Integration + Documentation
1. Update CLI integration
2. Create migration documentation
3. User acceptance testing

### Success Criteria

#### Technical Criteria
- [ ] All existing functionality preserved in classic mode
- [ ] Structured output provides richer data than current logs
- [ ] Performance overhead < 10% in structured mode
- [ ] 100% backward compatibility for existing scripts

#### User Experience Criteria
- [ ] No learning curve for existing users (classic mode default)
- [ ] Clear migration path to structured output
- [ ] Comprehensive help and documentation
- [ ] Intuitive parameter names and behavior

#### Integration Criteria
- [ ] Easy integration with monitoring systems
- [ ] Standard data formats (JSON, CSV, YAML)
- [ ] Composable with standard CLI tools (jq, grep, etc.)

## Next Steps

### Immediate Actions (Next 1-2 days)
1. **Create foundation package** - Implement `lib/glazed/` with parameter layers
2. **Prototype ping command** - Single command implementation as proof of concept
3. **Validate approach** - Ensure technical feasibility and performance

### Short-term Actions (Next 1-2 weeks)
1. **Complete core command migration** - Implement ping, volume, recall
2. **Integration testing** - Ensure compatibility with existing usage patterns
3. **Performance optimization** - Address any bottlenecks in structured mode

### Long-term Actions (Next 1-2 months)
1. **User migration support** - Documentation, examples, training
2. **Advanced features** - Custom output templates, additional filters
3. **Ecosystem integration** - Integration with monitoring and automation tools

## Conclusion

The migration to Glazed framework provides significant benefits for ppa-control CLI:

**Technical Benefits:**
- Structured data output enables integration with modern tooling
- Parameter validation and help generation improve user experience  
- Composable command architecture enables programmatic usage

**Strategic Benefits:**
- Future-proofs CLI architecture
- Enables data-driven automation and monitoring
- Maintains backward compatibility during transition

**Recommended Approach:**
- Implement dual commands for gradual migration
- Start with ping command as proof of concept
- Maintain classic mode as default for backward compatibility

The estimated 5.5-9 days of development time provides significant long-term value through improved CLI capabilities and integration potential.
