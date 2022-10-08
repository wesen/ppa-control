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

// You can create virtual ethernet interfaces on linux using the `dummy` driver
// ip link add eth10 type dummy
// ip link name eth10 dev dummy0
// ip addr add 192.168.100.199/24 brd + dev eth10 label eth10:0
// ip link set eth10 up
//
// ip link delete eth10 type dummy

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "SendPing one or multiple PPA servers",
	Run: func(cmd *cobra.Command, args []string) {
		addresses, _ := cmd.PersistentFlags().GetString("addresses")
		discovery, _ := cmd.PersistentFlags().GetBool("discover")
		componentId, _ := cmd.PersistentFlags().GetUint("componentId")

		port, _ := cmd.PersistentFlags().GetUint("port")

		var clients []client.Client
		for _, addr := range strings.Split(addresses, ",") {
			if addr != "" {
				clients = append(clients, client.NewClient(fmt.Sprintf("%s:%d", addr, port), int(componentId)))
			}
		}
		multiClient := client.NewMultiClient(clients)
		ctx := context.Background()
		grp, ctx := errgroup.WithContext(ctx)

		if discovery {
			discoveryCh := make(chan client.PeerInformation)
			grp.Go(func() error {
				for {
					select {
					case <-ctx.Done():
						return nil
					case msg := <-discoveryCh:
						log.Debug().Str("addr", msg.GetAddress()).Msg("discovery message")
						switch msg.(type) {
						case client.PeerDiscovered:
							log.Info().Str("addr", msg.GetAddress()).Msg("peer discovered")
						case client.PeerLost:
							log.Info().Str("addr", msg.GetAddress()).Msg("peer lost")
						}
					}
				}
			})
			grp.Go(func() error {
				return client.Discover(ctx, discoveryCh, uint16(port))
			})
		}

		receivedCh := make(chan client.ReceivedMessage)

		grp.Go(func() error {
			// runs both the send and read loops
			return multiClient.Run(ctx, &receivedCh)
		})

		grp.Go(func() error {
			multiClient.SendPing()

			for {
				t := time.NewTimer(5 * time.Second)

				select {
				case <-ctx.Done():
					return ctx.Err()

				case <-t.C:
					multiClient.SendPing()

				case msg := <-receivedCh:
					if msg.Header != nil {
						log.Info().Str("from", msg.RemoteAddress.String()).
							Str("client", msg.Client.Name()).
							Str("type", msg.Header.MessageType.String()).
							Str("status", msg.Header.Status.String()).
							Msg("received message")
					} else {
						log.Debug().Str("from", msg.RemoteAddress.String()).
							Str("client", msg.Client.Name()).
							Msg("received unknown message")
					}
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
	pingCmd.PersistentFlags().UintP(
		"componentId", "c", 0xFF,
		"Component ID to use for devices")

	pingCmd.PersistentFlags().UintP("port", "p", 5001, "Port to ping on")
}
