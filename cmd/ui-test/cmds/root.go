package cmds

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"ppa-control/cmd/ui-test/ui"
	"ppa-control/lib/client"
	"ppa-control/lib/client/discovery"
	logger "ppa-control/lib/log"
	"ppa-control/lib/utils"
	"strings"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "ui",
	Short: "main ppa-control UI",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		withCaller, _ := cmd.Flags().GetBool("with-caller")
		fmt.Println("withCaller", withCaller)
		logger.InitializeLogger(withCaller)

		logFormat, _ := cmd.Flags().GetString("log-format")
		if logFormat == "text" {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		} else {
			// json is the default
		}

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
		ctx := context.Background()
		ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
		defer func() {
			log.Debug().Msg("Cancelling context")
			cancel()
		}()

		grp, ctx2 := errgroup.WithContext(ctx)

		receivedCh := make(chan client.ReceivedMessage)
		discoveryCh := make(chan discovery.PeerInformation)

		addresses, _ := cmd.Flags().GetString("addresses")
		componentId, _ := cmd.Flags().GetUint("componentId")
		discover, _ := cmd.Flags().GetBool("discover")
		port, _ := cmd.Flags().GetUint("port")
		interfaces, _ := cmd.Flags().GetStringArray("interfaces")

		multiClient := client.NewMultiClient()
		for _, addr := range strings.Split(addresses, ",") {
			if addr == "" {
				continue
			}
			log.Info().Msgf("adding client %s", addr)
			_, err := multiClient.AddClient(ctx2, addr, "", componentId)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to add client")
			}
		}

		if discover {
			grp.Go(func() error {
				return discovery.Discover(ctx, discoveryCh, interfaces, uint16(port))
			})
		}

		zerolog.SetGlobalLevel(zerolog.DebugLevel)

		ui_ := ui.BuildUI(multiClient, cancel)
		ui_.Log("Startup")

		grp.Go(func() error {
			return multiClient.Run(ctx2, &receivedCh)
		})

		grp.Go(func() error {
			for {
				select {
				case <-ctx2.Done():
					return ctx2.Err()
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
				case msg := <-discoveryCh:
					log.Debug().Str("addr", msg.GetAddress()).Msg("discovery message")
					switch msg.(type) {
					case discovery.PeerDiscovered:
						log.Info().
							Str("addr", msg.GetAddress()).
							Str("iface", msg.GetInterface()).
							Msg("peer discovered")
						ui_.Log("Peer discovered: " + msg.GetAddress())
						c, err := multiClient.AddClient(ctx, msg.GetAddress(), msg.GetInterface(), componentId)
						if err != nil {
							log.Error().Err(err).Msg("failed to add client")
							return err
						}
						// send immediate ping
						c.SendPing()
					case discovery.PeerLost:
						log.Info().
							Str("addr", msg.GetAddress()).
							Str("iface", msg.GetInterface()).
							Msg("peer lost")
						ui_.Log("Peer lost: " + msg.GetAddress())
						err := multiClient.CancelClient(msg.GetAddress())
						if err != nil {
							log.Error().Err(err).Msg("failed to remove client")
							return err
						}
					}
				}
			}
		})

		// TODO this feels quite odd, let's learn more about fyne next
		go func() {
			log.Debug().Msg("Waiting for main loop")
			err := grp.Wait()
			log.Debug().Msg("Waited for main loop")

			if err != nil {
				log.Error().Err(err).Msg("Error in main loop")
			}
		}()

		ui_.Run()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("log-level", "debug", "Log level")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format (json, text)")
	rootCmd.PersistentFlags().Bool("with-caller", false, "Log caller")
	rootCmd.PersistentFlags().String("dump-mem-profile", "", "Dump memory profile to file")
	rootCmd.PersistentFlags().Bool("track-leaks", false, "Track memory and goroutine leaks")

	rootCmd.PersistentFlags().StringP(
		"addresses", "a", "",
		"Addresses to ping, comma separated",
	)
	// disable discovery by default when pinging
	rootCmd.PersistentFlags().BoolP(
		"discover", "d", true,
		"Send broadcast discovery messages",
	)

	rootCmd.PersistentFlags().StringArray("interfaces", []string{}, "Interfaces to use for discovery")

	rootCmd.PersistentFlags().UintP(
		"componentId", "c", 0xFF,
		"Component ID to use for devices")

	rootCmd.PersistentFlags().UintP("port", "p", 5001, "Port to ping on")
}
