package cmds

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"ppa-control/lib/client"
	"ppa-control/lib/client/discovery"
	"strings"
	"time"
)

var recallCmd = &cobra.Command{
	Use:   "recall",
	Short: "Recall a preset by index",
	Run: func(cmd *cobra.Command, args []string) {
		addresses, _ := cmd.PersistentFlags().GetString("addresses")
		discovery_, _ := cmd.PersistentFlags().GetBool("discover")
		//loop, _ := cmd.PersistentFlags().GetBool("loop")
		componentId, _ := cmd.PersistentFlags().GetUint("componentId")
		preset, _ := cmd.PersistentFlags().GetInt("preset")

		port, _ := cmd.PersistentFlags().GetUint("port")

		ctx := context.Background()
		grp, ctx := errgroup.WithContext(ctx)

		discoveryCh := make(chan discovery.PeerInformation)
		receivedCh := make(chan client.ReceivedMessage)

		multiClient := client.NewMultiClient()
		for _, addr := range strings.Split(addresses, ",") {
			if addr == "" {
				continue
			}
			// TODO allow passing in the interface name for a client, here
			_, err := multiClient.AddClient(ctx, addr, "", componentId)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to add client")
			}
		}

		if discovery_ {
			grp.Go(func() error {
				return discovery.Discover(ctx, discoveryCh, nil, uint16(port))
			})
		}

		grp.Go(func() error {
			return multiClient.Run(ctx, &receivedCh)
		})

		grp.Go(func() error {
			// TODO we do need to wait for this to have made it at least out of the socket
			multiClient.SendPresetRecallByPresetIndex(preset)

			preset = (preset + 1) % 5

			for {
				t := time.NewTicker(5 * time.Second)
				select {
				case <-ctx.Done():
					return ctx.Err()

				case msg := <-discoveryCh:
					log.Debug().Str("addr", msg.GetAddress()).Msg("discovery message")
					switch msg.(type) {
					case discovery.PeerDiscovered:
						log.Info().
							Str("addr", msg.GetAddress()).
							Str("iface", msg.GetInterface()).
							Msg("peer discovered")
						c, err := multiClient.AddClient(ctx, msg.GetAddress(), msg.GetInterface(), componentId)
						if err != nil {
							log.Error().Err(err).Msg("failed to add client")
							return err
						}
						// immediately send preset recall on discovery
						c.SendPresetRecallByPresetIndex(preset)
					case discovery.PeerLost:
						log.Info().
							Str("addr", msg.GetAddress()).
							Str("iface", msg.GetInterface()).
							Msg("peer lost")
						err := multiClient.CancelClient(msg.GetAddress())
						if err != nil {
							log.Error().Err(err).Msg("failed to remove client")
							return err
						}
					}

				case msg := <-receivedCh:
					if msg.Header != nil {
						log.Info().Str("from", msg.RemoteAddress.String()).
							Str("type", msg.Header.MessageType.String()).
							Str("client", msg.Client.Name()).
							Str("status", msg.Header.Status.String()).
							Msg("received message")
					} else {
						log.Debug().
							Str("from", msg.RemoteAddress.String()).
							Str("client", msg.Client.Name()).
							Msg("received unknown message")
					}

				case <-t.C:
					multiClient.SendPresetRecallByPresetIndex(preset)
				}
			}
		})

		err := grp.Wait()
		if err != nil {
			log.Error().Err(err).Msg("Error running multiclient")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(recallCmd)

	recallCmd.PersistentFlags().StringP(
		"addresses", "a", "",
		"Addresses to ping, comma separated",
	)
	recallCmd.PersistentFlags().BoolP(
		"discover", "d", true,
		"Send broadcast discovery messages",
	)
	recallCmd.PersistentFlags().BoolP(
		"loop", "l", true,
		"Send recalls in a loop",
	)
	recallCmd.PersistentFlags().UintP(
		"componentId", "c", 0xFF,
		"Component ID to use for devices")
	recallCmd.PersistentFlags().IntP(
		"preset", "", 0,
		"Preset to recall")

	recallCmd.PersistentFlags().UintP("port", "p", 5001, "Port to ping on")
}
