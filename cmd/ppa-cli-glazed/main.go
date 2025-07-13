package main

import (
	"os"

	"ppa-control/lib/glazed/commands"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "ppa-cli-glazed",
		Short: "PPA Control CLI with Glazed framework support",
		Long:  "A CLI tool for managing PPA (Programmable Power Array) devices with structured output support.",
	}

	// Create ping command
	pingCmd, err := commands.NewPingCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create ping command")
	}

	// Build Cobra command with dual mode support
	pingCobraCmd, err := cli.BuildCobraCommandDualMode(
		pingCmd,
		cli.WithGlazeToggleFlag("structured-output"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build ping Cobra command")
	}

	// Add commands to root
	rootCmd.AddCommand(pingCobraCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
