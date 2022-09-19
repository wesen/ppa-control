package cmds

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"ppa-control/lib/client"
	"strings"
	"time"
)

var recallCmd = &cobra.Command{
	Use:   "recall",
	Short: "Recall a preset by index",
	Run: func(cmd *cobra.Command, args []string) {
		addresses, _ := cmd.PersistentFlags().GetString("addresses")
		discovery, _ := cmd.PersistentFlags().GetBool("discover")
		componentId, _ := cmd.PersistentFlags().GetUint("componentId")
		preset, _ := cmd.PersistentFlags().GetInt("preset")

		port, _ := cmd.PersistentFlags().GetUint("port")

		var clients []client.Client
		for _, addr := range strings.Split(addresses, ",") {
			clients = append(clients, client.NewClient(fmt.Sprintf("%s:%d", addr, port), int(componentId)))
		}
		multiClient := client.NewMultiClient(clients)
		ctx := context.Background()
		grp, ctx := errgroup.WithContext(ctx)

		if discovery {
			grp.Go(func() error {
				return multiClient.Discover(ctx, uint16(port))
			})
		}

		receivedCh := make(chan client.ReceivedMessage)

		grp.Go(func() error {
			// TODO print out received messages
			return multiClient.Run(ctx, &receivedCh)
		})

		grp.Go(func() error {

			multiClient.SendPresetRecallByPresetIndex(preset)
			for {
				preset = (preset + 1) % 5
				t := time.NewTicker(5 * time.Second)
				select {
				case <-t.C:
					multiClient.SendPresetRecallByPresetIndex(preset)
				case msg := <-receivedCh:
					if msg.Header != nil {
						log.Info().Str("from", msg.Address.String()).
							Str("type", msg.Header.MessageType.String()).
							Str("status", msg.Header.Status.String()).
							Msg("received message")
					} else {
						log.Debug().Str("from", msg.Address.String()).Msg("received unknown message")
					}
				}
			}

			return nil
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
	recallCmd.PersistentFlags().UintP(
		"componentId", "c", 0xFF,
		"Component ID to use for devices")
	recallCmd.PersistentFlags().IntP(
		"preset", "", 0,
		"Preset to recall")

	recallCmd.PersistentFlags().UintP("port", "p", 5001, "Port to ping on")
}
