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

	// Add ping command
	pingCmd, err := NewPingCommand()
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

	rootCmd.AddCommand(pingCobraCmd)

	// Add volume command
	volumeCmd, err := NewVolumeCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create volume command")
	}

	// Build Cobra command with dual mode support
	volumeCobraCmd, err := cli.BuildCobraCommandDualMode(
		volumeCmd,
		cli.WithGlazeToggleFlag("structured-output"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build volume Cobra command")
	}

	rootCmd.AddCommand(volumeCobraCmd)

	// Add recall command
	recallCmd, err := NewRecallCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create recall command")
	}

	// Build Cobra command with dual mode support
	recallCobraCmd, err := cli.BuildCobraCommandDualMode(
		recallCmd,
		cli.WithGlazeToggleFlag("structured-output"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build recall Cobra command")
	}

	rootCmd.AddCommand(recallCobraCmd)

	// Add UDP broadcast command
	udpBroadcastCmd, err := NewUDPBroadcastCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create UDP broadcast command")
	}

	// Build Cobra command with dual mode support
	udpBroadcastCobraCmd, err := cli.BuildCobraCommandDualMode(
		udpBroadcastCmd,
		cli.WithGlazeToggleFlag("structured-output"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build UDP broadcast Cobra command")
	}

	rootCmd.AddCommand(udpBroadcastCobraCmd)

}
