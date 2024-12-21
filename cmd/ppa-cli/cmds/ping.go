package cmds

import (
	"ppa-control/lib/client"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "SendPing one or multiple PPA servers",
	Run: func(cmd *cobra.Command, args []string) {
		// Setup command context
		cmdCtx := client.SetupCommand(cmd)
		defer cmdCtx.Cancel()

		// Setup multiclient
		if err := cmdCtx.SetupMultiClient("ping"); err != nil {
			log.Fatal().Err(err).Msg("Failed to setup multiclient")
			return
		}

		// Setup discovery if enabled
		cmdCtx.SetupDiscovery()

		// Start multiclient
		cmdCtx.StartMultiClient()

		// Main command loop
		cmdCtx.RunInGroup(func() error {
			// Send initial ping
			cmdCtx.GetMultiClient().SendPing()

			for {
				t := time.NewTimer(5 * time.Second)

				select {
				case <-cmdCtx.Context().Done():
					t.Stop()
					return cmdCtx.Context().Err()

				case <-t.C:
					cmdCtx.GetMultiClient().SendPing()

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
						// Send ping immediately to newly discovered client
						newClient.SendPing()
					}
				}
			}
		})

		// Wait for completion
		cmdCtx.Wait()
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
	pingCmd.PersistentFlags().StringP(
		"addresses", "a", "",
		"Addresses to ping, comma separated",
	)
	// disable discovery by default when pinging
	pingCmd.PersistentFlags().BoolP(
		"discover", "d", false,
		"Send broadcast discovery messages",
	)
	pingCmd.PersistentFlags().StringArray(
		"interfaces", []string{},
		"Interfaces to use for discovery",
	)
	pingCmd.PersistentFlags().UintP(
		"componentId", "c", 0xFF,
		"Component ID to use for devices",
	)
	pingCmd.PersistentFlags().UintP(
		"port", "p", 5001,
		"Port to ping on",
	)
}
