package cmds

import (
	"context"
	"fmt"
	"net"
	"time"

	"ppa-control/lib/glazed"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"

)

const (
	// UDPBroadcastLayerSlug is the identifier for the UDP broadcast parameter layer
	UDPBroadcastLayerSlug = "udp-broadcast"
)

// UDPBroadcastSettings represents the UDP broadcast specific parameters
type UDPBroadcastSettings struct {
	Address   string `glazed.parameter:"address"`
	Server    bool   `glazed.parameter:"server"`
	UDPPort   uint   `glazed.parameter:"udp-port"`
	Interface string `glazed.parameter:"interface"`
}

// NewUDPBroadcastParameterLayer creates a parameter layer for UDP broadcast configuration
func NewUDPBroadcastParameterLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		UDPBroadcastLayerSlug,
		"UDP Broadcast Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"address",
				parameters.ParameterTypeString,
				parameters.WithDefault("localhost"),
				parameters.WithHelp("Address to listen on or send to"),
				parameters.WithShortFlag("u"),
			),
			parameters.NewParameterDefinition(
				"server",
				parameters.ParameterTypeBool,
				parameters.WithDefault(false),
				parameters.WithHelp("Run as server (listen mode)"),
				parameters.WithShortFlag("s"),
			),
			parameters.NewParameterDefinition(
				"udp-port",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(5001),
				parameters.WithHelp("UDP port to use"),
				parameters.WithShortFlag("P"),
			),
			parameters.NewParameterDefinition(
				"interface",
				parameters.ParameterTypeString,
				parameters.WithHelp("Interface to bind to"),
				parameters.WithShortFlag("i"),
			),
		),
	)
}

// UDPBroadcastGlazedCommand handles UDP broadcast operations with glazed framework support.
// It supports classic (human-readable) and structured Glazed output.
type UDPBroadcastGlazedCommand struct {
	*cmds.CommandDescription
}

// Run implements the BareCommand interface for classic text output
func (c *UDPBroadcastGlazedCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Initialize logging
	if err := glazed.InitLogging(parsedLayers); err != nil {
		return err
	}

	// Extract UDP broadcast settings
	settings := &UDPBroadcastSettings{}
	if err := parsedLayers.InitializeStruct(UDPBroadcastLayerSlug, settings); err != nil {
		return err
	}

	return c.runUDPBroadcast(ctx, settings, nil)
}

// RunIntoGlazeProcessor implements the GlazeCommand interface for structured output
func (c *UDPBroadcastGlazedCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Initialize logging
	if err := glazed.InitLogging(parsedLayers); err != nil {
		return err
	}

	// Extract UDP broadcast settings
	settings := &UDPBroadcastSettings{}
	if err := parsedLayers.InitializeStruct(UDPBroadcastLayerSlug, settings); err != nil {
		return err
	}

	return c.runUDPBroadcast(ctx, settings, gp)
}

// runUDPBroadcast contains the main UDP broadcast logic
func (c *UDPBroadcastGlazedCommand) runUDPBroadcast(ctx context.Context, settings *UDPBroadcastSettings, gp middlewares.Processor) error {
	listenAddr := ":0"
	if settings.Server {
		listenAddr = fmt.Sprintf("%s:%d", settings.Address, settings.UDPPort)
	}

	pc, err := net.ListenPacket("udp4", listenAddr)
	if err != nil {
		return err
	}
	defer pc.Close()

	// Emit initial setup information
	if gp != nil {
		row := types.NewRow(
			types.MRP("timestamp", time.Now()),
			types.MRP("event", "socket_created"),
			types.MRP("listen_addr", listenAddr),
			types.MRP("actual_addr", pc.LocalAddr().String()),
			types.MRP("server_mode", settings.Server),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	} else {
		if settings.Server {
			fmt.Printf("Listening on %s\n", pc.LocalAddr().String())
		}
	}

	// Handle interface binding if specified
	if settings.Interface != "" {
		if udpConn, ok := pc.(*net.UDPConn); ok {
			c, err := udpConn.SyscallConn()
			if err != nil {
				return err
			}
			err = c.Control(func(fd uintptr) {
				if gp != nil {
					row := types.NewRow(
						types.MRP("timestamp", time.Now()),
						types.MRP("event", "interface_bind_attempt"),
						types.MRP("interface", settings.Interface),
						types.MRP("socket_fd", int(fd)),
					)
					gp.AddRow(ctx, row)
				} else {
					fmt.Printf("Binding socket %d to interface %s\n", fd, settings.Interface)
				}
			})
			if err != nil {
				return err
			}
		}
	}

	if settings.Server {
		// Server mode: listen for messages
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				buf := make([]byte, 1024)
				n, addr, err := pc.ReadFrom(buf)
				if err != nil {
					if gp != nil {
						row := types.NewRow(
							types.MRP("timestamp", time.Now()),
							types.MRP("event", "read_error"),
							types.MRP("error", err.Error()),
						)
						gp.AddRow(ctx, row)
					}
					return err
				}

				message := string(buf[:n])
				if gp != nil {
					row := types.NewRow(
						types.MRP("timestamp", time.Now()),
						types.MRP("event", "message_received"),
						types.MRP("from", addr.String()),
						types.MRP("message", message),
						types.MRP("size", n),
					)
					if err := gp.AddRow(ctx, row); err != nil {
						return err
					}
				} else {
					fmt.Printf("%s sent this: %s\n", addr, message)
				}
			}
		}
	} else {
		// Client mode: send message
		dstAddr := fmt.Sprintf("%s:%d", settings.Address, settings.UDPPort)
		addr, err := net.ResolveUDPAddr("udp4", dstAddr)
		if err != nil {
			return err
		}

		message := "data to transmit"
		if gp != nil {
			row := types.NewRow(
				types.MRP("timestamp", time.Now()),
				types.MRP("event", "sending_message"),
				types.MRP("to", dstAddr),
				types.MRP("message", message),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		} else {
			fmt.Printf("Sending to %s\n", dstAddr)
		}

		n, err := pc.WriteTo([]byte(message), addr)
		if err != nil {
			if gp != nil {
				row := types.NewRow(
					types.MRP("timestamp", time.Now()),
					types.MRP("event", "send_error"),
					types.MRP("error", err.Error()),
				)
				gp.AddRow(ctx, row)
			}
			return err
		}

		if gp != nil {
			row := types.NewRow(
				types.MRP("timestamp", time.Now()),
				types.MRP("event", "message_sent"),
				types.MRP("to", dstAddr),
				types.MRP("bytes_sent", n),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	return nil
}

// NewUDPBroadcastGlazedCommand creates a new UDP broadcast command with glazed support
func NewUDPBroadcastGlazedCommand() (*UDPBroadcastGlazedCommand, error) {
	// Get standard layers
	standardLayers, err := glazed.NewStandardLayers()
	if err != nil {
		return nil, err
	}

	// Create UDP broadcast specific layer
	udpBroadcastLayer, err := NewUDPBroadcastParameterLayer()
	if err != nil {
		return nil, err
	}

	// Combine layers: UDP broadcast layer first, then standard layers
	allLayers := []layers.ParameterLayer{udpBroadcastLayer}
	allLayers = append(allLayers, standardLayers...)

	cmdDesc := cmds.NewCommandDescription(
		"udp-broadcast-glazed",
		cmds.WithShort("Send and receive UDP broadcast messages with structured output support"),
		cmds.WithLong("Send or receive UDP broadcast messages. "+
			"Supports both human-readable logging output and structured data output. "+
			"Use --server to run in listen mode, or specify --address to send messages."),
		cmds.WithLayersList(allLayers...),
	)

	return &UDPBroadcastGlazedCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure interface compliance
var _ cmds.BareCommand = &UDPBroadcastGlazedCommand{}
var _ cmds.GlazeCommand = &UDPBroadcastGlazedCommand{}
