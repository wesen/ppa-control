# PPA-Control CLI Migration to Glazed Framework - Technical Sketch

## Objective
Migrate the existing Cobra-based CLI commands to the Glazed framework in order to gain structured data output, unified parameter handling, and future extensibility.

## Commands In Scope
- ping
- volume
- recall
- simulate
- udp-broadcast
- root

## Architectural Sketch
1. **Glazed integration package**
   - `lib/glazed/layers.go` – define shared parameter layers ( ppa).
      - Parse layer into PPASetitngs struct
      - Use that struct to:
         - Add a method CreateCommandContextFromParsedLayers()
   - Use initLogging from glazed for setting up the logging (see glazed/cmd/glaze/main.go)
   - Try to keep things as close as possible to the original source, no new files if possible, etc...


2. **Command skeletons** (example shown for `ping`):

```go
package commands

import (
    "context"

    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/middlewares"
)

// PingCommand handles PPA ping messages.
// It supports classic (human-readable) and structured Glazed output.
type PingCommand struct {
    *cmds.CommandDescription
}

func (c *PingCommand) Run(ctx context.Context, pl *layers.ParsedLayers) error {
    // setup context → multiclient → discovery → main loop (logs only)
    return nil
}

func (c *PingCommand) RunIntoGlazeProcessor(
    ctx context.Context, pl *layers.ParsedLayers, gp middlewares.Processor,
) error {
    // same setup as Run(), but emit gp.AddRow(...) for structured data
    return nil
}

// Ensure interface compliance.
var _ cmds.BareCommand  = &PingCommand{}
var _ cmds.GlazeCommand = &PingCommand{}
```

3. **CLI entrypoint** (`cmd/ppa-cli-glazed/main.go`):
   - Instantiate Glazed commands (`NewPingCommand`, `NewVolumeCommand`, …).
   - Wrap each with `cli.BuildCobraCommandDualMode` and attach to the root `cobra.Command`.

## High-Level Implementation Phases
1. Foundation (layers + adapter).
2. Port `ping`.
3. Port `volume` and `recall`.
4. Integrate CLI and iterate.

## Notable Risks
- Channel handling complexity when sharing logic between classic and structured modes.

## Output Mode Guidelines

| Command | Nature | Output Mode |
|---------|--------|-------------|
| ping | Continuous monitoring / polling | **Dual** (structured rows + optional log streaming) |
| simulate | Long-running simulation | **Dual** |
| udp-broadcast (listen/server modes) | May run indefinitely | **Dual** |
| volume | Single, quick query/set operation | **Structured-only** |
| recall | Single, quick query/set operation | **Structured-only** |
| root (help/version etc.) | Instant | **Structured-only** |

Structured-only verbs implement only the `cmds.GlazeCommand` interface, whereas dual-mode verbs satisfy both `cmds.BareCommand` and `cmds.GlazeCommand`.

---
