package client

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"time"
)

type PeerInformation interface {
	GetAddress() string
}

type PeerDiscovered struct {
	addr string
}

type PeerLost struct {
	addr string
}

func (c PeerDiscovered) GetAddress() string {
	return c.addr
}

func (c PeerLost) GetAddress() string {
	return c.addr
}

func Discover(ctx context.Context, msgCh chan PeerInformation, port uint16) error {
	broadcastAddr := fmt.Sprintf("255.255.255.255:%d", port)

	// From: https://stackoverflow.com/questions/683624/udp-broadcast-on-all-interfaces
	//
	// With a single sendto(), only a single packet is generated, and the outgoing interface is
	// determined by the operating system's routing table (ip route on linux). You can't have
	// a single sendto() generate more than one packet, you would have to iterate over all
	// interfaces, and either use raw sockets or bind the socket to a device using
	// setsockopt(..., SOL_SOCKET, SO_BINDTODEVICE, "ethX") to send each packet bypassing the
	// OS routing table (this requires root privileges). Not a good solution.
	//
	// TODO(manuel) We should bind each client to an interface in order to send out multiple broadcast
	// packets
	c := NewClient(broadcastAddr, 0xfe)

	receivedCh := make(chan ReceivedMessage)

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return c.Run(ctx, &receivedCh)
	})
	grp.Go(func() error {
		log.Debug().Msg("Starting discovery loop")
		c.SendPing()

		peerLastSeen := make(map[string]time.Time)
		peerTimeout := 30 * time.Second

		for {
			t := time.NewTimer(5 * time.Second)

			select {
			case <-t.C:
				for addr, lastSeen := range peerLastSeen {
					if time.Since(lastSeen) > peerTimeout {
						log.Debug().Str("addr", addr).Msg("peer lost")
						msgCh <- PeerLost{addr: addr}
						delete(peerLastSeen, addr)
					}
				}
				c.SendPing()

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

					_, peerFound := peerLastSeen[msg.RemoteAddress.String()]
					if !peerFound {
						log.Info().Str("addr", msg.RemoteAddress.String()).Msg("new peer discovered")
						msgCh <- PeerDiscovered{msg.RemoteAddress.String()}
					}
					peerLastSeen[msg.RemoteAddress.String()] = time.Now()
					log.Debug().Str("addr", msg.RemoteAddress.String()).Msg("peer lastSeen updated")
				}

			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	return grp.Wait()
}
