package cmds

import (
	"ppa-control/lib"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Set the volume of one or more clients",
	Run: func(cmd *cobra.Command, args []string) {
		// Get command-specific flags
		volume, _ := cmd.PersistentFlags().GetFloat32("volume")
		loop, _ := cmd.PersistentFlags().GetBool("loop")

		// Validate volume range
		if volume < 0 || volume > 1 {
			log.Fatal().Msg("Volume must be between 0 and 1")
		}

		// Setup command context
		cmdCtx := lib.SetupCommand(cmd)
		defer cmdCtx.Cancel()

		// Setup multiclient
		if err := cmdCtx.SetupMultiClient("volume"); err != nil {
			log.Fatal().Err(err).Msg("Failed to setup multiclient")
			return
		}

		// Setup discovery if enabled
		cmdCtx.SetupDiscovery()

		// Start multiclient
		cmdCtx.StartMultiClient()

		// Main command loop
		cmdCtx.RunInGroup(func() error {
			// Send initial volume
			cmdCtx.GetMultiClient().SendMasterVolume(volume)

			// If not looping, just wait for context cancellation
			if !loop {
				<-cmdCtx.Context().Done()
				return cmdCtx.Context().Err()
			}

			for {
				t := time.NewTimer(5 * time.Second)

				select {
				case <-cmdCtx.Context().Done():
					t.Stop()
					return cmdCtx.Context().Err()

				case <-t.C:
					cmdCtx.GetMultiClient().SendMasterVolume(volume)

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
						// Send volume immediately to newly discovered client
						newClient.SendMasterVolume(volume)
					}
				}
			}
		})

		// Wait for completion
		cmdCtx.Wait()
	},
}

func init() {
	rootCmd.AddCommand(volumeCmd)

	volumeCmd.PersistentFlags().StringP(
		"addresses", "a", "",
		"Addresses to control, comma separated",
	)
	volumeCmd.PersistentFlags().BoolP(
		"discover", "d", true,
		"Send broadcast discovery messages",
	)
	volumeCmd.PersistentFlags().StringArray(
		"interfaces", []string{},
		"Interfaces to use for discovery",
	)
	volumeCmd.PersistentFlags().UintP(
		"componentId", "c", 0xFF,
		"Component ID to use for devices",
	)
	volumeCmd.PersistentFlags().Float32P(
		"volume", "v", 0.5,
		"Volume level (0.0-1.0)",
	)
	volumeCmd.PersistentFlags().BoolP(
		"loop", "l", true,
		"Send volume commands in a loop",
	)
	volumeCmd.PersistentFlags().UintP(
		"port", "p", 5001,
		"Port to use",
	)
}
