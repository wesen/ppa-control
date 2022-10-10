package cmds

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"os"
	logger "ppa-control/lib/log"
)

var rootCmd = &cobra.Command{
	Use:   "ppa-cli",
	Short: "ppa-cli is a command line interface for the PPA protocol",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		withCaller, _ := cmd.Flags().GetBool("with-caller")
		fmt.Println("withCaller", withCaller)
		logger.InitializeLogger(withCaller)

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
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("log-level", "l", "debug", "Log level")
	rootCmd.PersistentFlags().Bool("with-caller", false, "Log caller")
}
