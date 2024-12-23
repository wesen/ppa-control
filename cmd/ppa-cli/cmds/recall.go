package cmds

import (
	"ppa-control/lib"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var recallCmd = &cobra.Command{
	Use:   "recall",
	Short: "Recall a preset by index",
	Run: func(cmd *cobra.Command, args []string) {
		// Get command-specific flags
		preset, _ := cmd.PersistentFlags().GetInt("preset")
		loop, _ := cmd.PersistentFlags().GetBool("loop")

		// Setup command context
		cmdCtx := lib.SetupCommand(cmd)
		defer cmdCtx.Cancel()

		// Setup multiclient
		if err := cmdCtx.SetupMultiClient("recall"); err != nil {
			log.Fatal().Err(err).Msg("Failed to setup multiclient")
			return
		}

		// Setup discovery if enabled
		cmdCtx.SetupDiscovery()

		// Start multiclient
		cmdCtx.StartMultiClient()

		// Main command loop
		cmdCtx.RunInGroup(func() error {
			// Send initial recall
			cmdCtx.GetMultiClient().SendPresetRecallByPresetIndex(preset)

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
					cmdCtx.GetMultiClient().SendPresetRecallByPresetIndex(preset)

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
						// Send recall immediately to newly discovered client
						newClient.SendPresetRecallByPresetIndex(preset)
					}
				}
			}
		})

		// Wait for completion
		cmdCtx.Wait()
	},
}

func init() {
	rootCmd.AddCommand(recallCmd)

	recallCmd.PersistentFlags().StringP(
		"addresses", "a", "",
		"Addresses to recall on, comma separated",
	)
	recallCmd.PersistentFlags().BoolP(
		"discover", "d", true,
		"Send broadcast discovery messages",
	)
	recallCmd.PersistentFlags().BoolP(
		"loop", "l", true,
		"Send recalls in a loop",
	)
	recallCmd.PersistentFlags().StringArray(
		"interfaces", []string{},
		"Interfaces to use for discovery",
	)
	recallCmd.PersistentFlags().UintP(
		"componentId", "c", 0xFF,
		"Component ID to use for devices",
	)
	recallCmd.PersistentFlags().IntP(
		"preset", "", 0,
		"Preset to recall",
	)
	recallCmd.PersistentFlags().UintP(
		"port", "p", 5001,
		"Port to use",
	)
}
