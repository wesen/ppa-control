package client

import (
	"context"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"time"
)

func (mc *multiClient) Discover(addresses []string) error {
	return nil
}

func (mc *multiClient) RunPingLoop(ctx context.Context) error {
	grp, ctx2 := errgroup.WithContext(ctx)

	grp.Go(func() error {
		log.Info().Msg("ping-loop started")
		for {
			log.Debug().Msg("pinging clients")
			for _, c := range mc.clients {
				log.Debug().Str("client", c.Name()).Msg("pinging client")
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
