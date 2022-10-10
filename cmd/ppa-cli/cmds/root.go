package cmds

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	logger "ppa-control/lib/log"
	"ppa-control/lib/utils"
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
		} else {
			// json is the default
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
			utils.StartBackgroundLeakTracker()
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
}
