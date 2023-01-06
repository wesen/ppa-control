package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	bucheron "github.com/wesen/bucheron/pkg"
	"golang.org/x/sync/errgroup"
	"os"
	app2 "ppa-control/cmd/ui-test/app"
	"ppa-control/lib/utils"
	"time"
)

var app = &app2.App{}

var rootCmd = &cobra.Command{
	Use:   "ui",
	Short: "main ppa-control UI",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		appConfig := app2.NewAppConfigFromCommand(cmd)

		app = &app2.App{
			Config: appConfig,
		}
		err := app.InitLogger()
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
	},
	Run: func(cmd *cobra.Command, args []string) {
		app.Run()
	},
}

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file to the PPA",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		progressChannel := make(chan bucheron.ProgressEvent)
		errGroup, ctx2 := errgroup.WithContext(ctx)

		errGroup.Go(func() error {
			for {
				select {
				case <-ctx2.Done():
					return ctx2.Err()
				case progress, ok := <-progressChannel:
					if !ok {
						return nil
					}
					fmt.Printf("%v: %v\n", progress, ok)
					fmt.Printf("Progress: %s %f\n", progress.Step, progress.StepProgress)
				}
			}
		})

		errGroup.Go(func() error {
			return app.UploadLogs(ctx2, progressChannel)
		})

		err := errGroup.Wait()
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.PersistentFlags().String("dump-mem-profile", "", "Dump memory profile to file")
	rootCmd.PersistentFlags().Bool("track-leaks", false, "Track memory and goroutine leaks")

	rootCmd.AddCommand(uploadCmd)

	app2.AddAppConfigFlags(rootCmd)

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
