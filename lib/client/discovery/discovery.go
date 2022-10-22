package discovery

import (
	"context"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"ppa-control/lib/client"
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

func Discover(
	ctx context.Context,
	msgCh chan PeerInformation,
	discoveryInterfaces []string,
	port uint16) error {
	receivedCh := make(chan client.ReceivedMessage)

	interfaceManager := NewInterfaceManager(port, &receivedCh)
	interfaceDiscoverer := NewInterfaceDiscoverer(interfaceManager, discoveryInterfaces)

	// start the discoverer:
	//   - GR1: interfaceDiscoverer.Run() which writes to addedInterfaceCh and removedInterfaceCh
	//   - GR2: Run() which reads from addedInterfaceCh and removedInterfaceCh
	//
	// GR1 and GR2 use the same context tree
	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return interfaceDiscoverer.Run(ctx)
	})

	grp.Go(func() error {
		log.Debug().Msg("Starting discovery loop")

		peerLastSeen := make(map[string]time.Time)
		peerTimeout := 30 * time.Second

		for {
			t := time.NewTimer(5 * time.Second)

			select {
			case <-t.C:
				for addr, lastSeen := range peerLastSeen {
					if time.Since(lastSeen) > peerTimeout {
						log.Debug().Str("addr", addr).Msg("peer lost")
						delete(peerLastSeen, addr)
						select {
						case msgCh <- PeerLost{addr: addr}:
						case <-ctx.Done():
							return ctx.Err()
						}
					}
				}

				interfaceManager.SendPing()

			case newInterface := <-interfaceDiscoverer.addedInterfaceCh:
				log.Debug().Str("iface", newInterface).Msg("new interface discovered")
				err, c := interfaceManager.StartInterfaceClient(ctx, newInterface)
				if err != nil {
					return err
				}
				// immediately send welcome ping
				c.SendPing()

			case removedInterface := <-interfaceDiscoverer.removedInterfaceCh:
				log.Debug().Str("iface", removedInterface).Msg("interface removed")
				err := interfaceManager.CancelInterfaceClient(removedInterface)
				if err != nil {
					return err
				}

			case msg := <-receivedCh:
				if msg.Header != nil {
					log.Info().Str("from", msg.RemoteAddress.String()).
						Str("client", msg.Client.Name()).
						Str("type", msg.Header.MessageType.String()).
						Str("status", msg.Header.Status.String()).
						Msg("received message")

					_, peerFound := peerLastSeen[msg.RemoteAddress.String()]
					peerLastSeen[msg.RemoteAddress.String()] = time.Now()
					if !peerFound {
						log.Info().Str("addr", msg.RemoteAddress.String()).Msg("new peer discovered")
						select {
						case msgCh <- PeerDiscovered{msg.RemoteAddress.String()}:
						case <-ctx.Done():
							return ctx.Err()
						}
					}
					log.Debug().Str("addr", msg.RemoteAddress.String()).Msg("peer lastSeen updated")
				} else {
					log.Debug().Str("from", msg.RemoteAddress.String()).
						Str("client", msg.Client.Name()).
						Msg("received unknown message")
				}

			case <-ctx.Done():
				log.Info().Msg("waiting for clients to stop")

				interfaceManager.Wait()

				return ctx.Err()
			}
		}
	})

	return grp.Wait()
}
