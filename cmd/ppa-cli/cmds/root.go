package cmds

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	logger "ppa-control/lib/log"
	"ppa-control/lib/utils"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
)

var rootCmd = &cobra.Command{
	Use:   "ppa-cli",
	Short: "ppa-cli is a command line interface for the PPA protocol",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		withCaller, _ := cmd.Flags().GetBool("with-caller")
		fmt.Println("withCaller", withCaller)
		logger.InitializeLogger(withCaller)

		logFormat, _ := cmd.Flags().GetString("log-format")
		if logFormat == "text" {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		}

		level, _ := cmd.Flags().GetString("log-level")
		switch level {
		case "debug":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "info":
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case "warn":
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case "error":
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		case "fatal":
			zerolog.SetGlobalLevel(zerolog.FatalLevel)
		}

		// TODO this will not compile on windows
		memProfile, _ := cmd.Flags().GetString("dump-mem-profile")
		log.Info().Str("memProfile", memProfile).Msg("Listening for SIGPOLL to dump stacktrace and memory profile")
		utils.StartSIGPOLLStacktraceDumper(memProfile)

		trackLeaks, _ := cmd.Flags().GetBool("track-leaks")
		if trackLeaks {
			log.Info().Msg("tracking memory and goroutine leaks")
			utils.StartBackgroundLeakTracker(5 * time.Second)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("log-level", "debug", "Log level")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format (json, text)")
	rootCmd.PersistentFlags().Bool("with-caller", false, "Log caller")
	rootCmd.PersistentFlags().String("dump-mem-profile", "", "Dump memory profile to file")
	rootCmd.PersistentFlags().Bool("track-leaks", false, "Track memory and goroutine leaks")

	// Add glazed ping command
	pingGlazedCmd, err := NewPingGlazedCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create glazed ping command")
	}

	// Build Cobra command with dual mode support
	pingGlazedCobraCmd, err := cli.BuildCobraCommandDualMode(
		pingGlazedCmd,
		cli.WithGlazeToggleFlag("structured-output"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build glazed ping Cobra command")
	}

	rootCmd.AddCommand(pingGlazedCobraCmd)

	// Add glazed volume command
	volumeGlazedCmd, err := NewVolumeGlazedCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create glazed volume command")
	}

	// Build Cobra command with dual mode support
	volumeGlazedCobraCmd, err := cli.BuildCobraCommandDualMode(
		volumeGlazedCmd,
		cli.WithGlazeToggleFlag("structured-output"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build glazed volume Cobra command")
	}

	rootCmd.AddCommand(volumeGlazedCobraCmd)

	// Add glazed recall command
	recallGlazedCmd, err := NewRecallGlazedCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create glazed recall command")
	}

	// Build Cobra command with dual mode support
	recallGlazedCobraCmd, err := cli.BuildCobraCommandDualMode(
		recallGlazedCmd,
		cli.WithGlazeToggleFlag("structured-output"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build glazed recall Cobra command")
	}

	rootCmd.AddCommand(recallGlazedCobraCmd)

	// Add glazed UDP broadcast command
	udpBroadcastGlazedCmd, err := NewUDPBroadcastGlazedCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create glazed UDP broadcast command")
	}

	// Build Cobra command with dual mode support
	udpBroadcastGlazedCobraCmd, err := cli.BuildCobraCommandDualMode(
		udpBroadcastGlazedCmd,
		cli.WithGlazeToggleFlag("structured-output"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build glazed UDP broadcast Cobra command")
	}

	rootCmd.AddCommand(udpBroadcastGlazedCobraCmd)

	// Add glazed simulate command
	simulateGlazedCmd, err := NewSimulateGlazedCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create glazed simulate command")
	}

	// Build Cobra command with dual mode support
	simulateGlazedCobraCmd, err := cli.BuildCobraCommandDualMode(
		simulateGlazedCmd,
		cli.WithGlazeToggleFlag("structured-output"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build glazed simulate Cobra command")
	}

	rootCmd.AddCommand(simulateGlazedCobraCmd)
}
