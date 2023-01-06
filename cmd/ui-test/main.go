package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"ppa-control/lib/utils"
	"time"
)

const DEFAULT_COMPONENT_ID = 0xFF

var app = &App{}

var rootCmd = &cobra.Command{
	Use:   "ui",
	Short: "main ppa-control UI",
	Run: func(cmd *cobra.Command, args []string) {
		appConfig := NewAppConfigFromCommand(cmd)

		app = &App{
			Config: appConfig,
		}
		err := app.initLogger()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize logger")
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
		app.Run()
	},
}

func init() {
	rootCmd.PersistentFlags().String("dump-mem-profile", "", "Dump memory profile to file")
	rootCmd.PersistentFlags().Bool("track-leaks", false, "Track memory and goroutine leaks")

	AddAppConfigFlags(rootCmd)

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
