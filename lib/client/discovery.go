package client

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"time"
)

func (mc *MultiClient) Discover(ctx context.Context, port uint16) error {
	broadcastAddr := fmt.Sprintf("255.255.255.255:%d", port)
	c := NewClient(broadcastAddr, 0xfe)

	receivedCh := make(chan ReceivedMessage)

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return c.Run(ctx, &receivedCh)
	})
	grp.Go(func() error {
		log.Debug().Msg("Starting discovery loop")
		c.SendPing()

		for {
			t := time.NewTimer(5 * time.Second)

			select {
			case <-t.C:
				c.SendPing()

			case msg := <-receivedCh:
				if msg.Header != nil {
					log.Info().Str("from", msg.Address.String()).
						Str("type", msg.Header.MessageType.String()).
						Str("status", msg.Header.Status.String()).
						Msg("received message")
				} else {
					log.Debug().Str("from", msg.Address.String()).Msg("received unknown message")
				}

			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	return grp.Wait()
}

func (mc *MultiClient) RunPingLoop(ctx context.Context) error {
	grp, ctx2 := errgroup.WithContext(ctx)

	// TODO we should do one goroutine per SingleDevice here, that reads and sends pings
	// although the reading will interact with the main receive loop maybe?
	// so the main loop for each SingleDevice will emit the parsed responses

	grp.Go(func() error {
		log.Info().Msg("ping-loop started")
		for {
			log.Debug().Msg("pinging clients")
			for _, c := range mc.clients {
				log.Debug().Str("SingleDevice", c.Name()).Msg("pinging SingleDevice")
				c.SendPing()
			}

			select {
			case <-ctx2.Done():
				log.Info().Msg("stopping ping loop")
				return ctx2.Err()

			case <-time.After(5 * time.Second):
				continue
			}

		}
	})

	return grp.Wait()
}
