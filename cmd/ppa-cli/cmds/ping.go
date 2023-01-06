package cmds

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"ppa-control/lib/client"
	"ppa-control/lib/client/discovery"
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
		discovery_, _ := cmd.PersistentFlags().GetBool("discover")
		componentId, _ := cmd.PersistentFlags().GetUint("componentId")

		port, _ := cmd.PersistentFlags().GetUint("port")

		ctx := context.Background()

		// TODO(@manuel) you are supposed to defer the cancelFunc when using a cancellable context
		ctx, cancelFunc := signal.NotifyContext(ctx, os.Interrupt)
		defer func() {
			log.Debug().Msg("Cancelling context")
			cancelFunc()
		}()

		grp, ctx := errgroup.WithContext(ctx)

		discoveryCh := make(chan discovery.PeerInformation)
		receivedCh := make(chan client.ReceivedMessage)

		multiClient := client.NewMultiClient("ping")
		for _, addr := range strings.Split(addresses, ",") {
			if addr == "" {
				continue
			}
			// although that's weird, in the case I want to specify over which iface to reach things.
			// say i have a simulated pkg bound to en0, then this will resolve the iface to be lo0.
			// this might mostly just be an issue with local testing.
			_, err := multiClient.AddClient(ctx, fmt.Sprintf("%s:%d", addr, port), "", componentId)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to add pkg")
			}
		}

		if discovery_ {
			interfaces, _ := cmd.PersistentFlags().GetStringArray("interfaces")
			grp.Go(func() error {
				return discovery.Discover(ctx, discoveryCh, interfaces, uint16(port))
			})
		}

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
							Str("pkg", msg.Client.Name()).
							Str("type", msg.Header.MessageType.String()).
							Str("status", msg.Header.Status.String()).
							Msg("received message")
					} else {
						log.Debug().Str("from", msg.RemoteAddress.String()).
							Str("pkg", msg.Client.Name()).
							Msg("received unknown message")
					}

				// this won't trigger if the discovery loop is not running
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
							log.Error().Err(err).Msg("failed to add pkg")
							return err
						}
						// send immediate ping
						c.SendPing()
					case discovery.PeerLost:
						log.Info().
							Str("addr", msg.GetAddress()).
							Str("iface", msg.GetInterface()).
							Msg("peer lost")
						err := multiClient.CancelClient(msg.GetAddress())
						if err != nil {
							log.Error().Err(err).Msg("failed to remove pkg")
							return err
						}
					}
				}
			}
		})

		err := grp.Wait()

		log.Debug().Err(err).Msg("finished ping loop")
		if err != nil && err.Error() != "context canceled" {
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

	pingCmd.PersistentFlags().StringArray("interfaces", []string{}, "Interfaces to use for discovery")

	pingCmd.PersistentFlags().UintP(
		"componentId", "c", 0xFF,
		"Component ID to use for devices")

	pingCmd.PersistentFlags().UintP("port", "p", 5001, "Port to ping on")
}
