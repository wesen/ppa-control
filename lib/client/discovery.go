package client

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net"
	"sync"
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

	receivedCh := make(chan ReceivedMessage)

	// make a waitGroup for all the client goroutines
	clientWg := &sync.WaitGroup{}

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		log.Debug().Msg("Starting discovery loop")

		clients := make(map[string]Client)
		clientCancels := make(map[string]context.CancelFunc)
		clientMutex := &sync.Mutex{}
		// maybe we need a running list of currently active clients (?)
		// or maybe the wait group is enough...

		ifaces, err := net.Interfaces()
		if err != nil {
			log.Error().Err(err).Msg("failed to get interfaces")
			return err
		}

		// create an initial set of clients,
		// but we should make this a method that recognizes added and removed interfaces.
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 {
				// interface down
				continue
			}
			if iface.Flags&net.FlagLoopback != 0 {
				// loopback interface
				continue
			}

			addrs, err := iface.Addrs()
			if err != nil {
				log.Error().Err(err).Msg("failed to get interface addresses")
				return err
			}
			// go over IP address range
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip == nil || ip.IsLoopback() {
					continue
				}

				// we don't do IPv6
				if ip.To4() == nil {
					continue
				}

				log.Info().
					Str("interface", iface.Name).
					Str("ip", ip.String()).
					Str("netmask", ip.DefaultMask().String()).
					Msg("found ip address")

				// now, create a client bound to the interface with the address being 255.255.255.255
				c := NewClient(broadcastAddr, 0xfe)

				clientCtx, cancel := context.WithCancel(ctx)
				func() {
					clientMutex.Lock()
					defer clientMutex.Unlock()

					clients[iface.Name] = c
					clientCancels[iface.Name] = cancel
				}()

				go func() {
					clientWg.Add(1)
					// we need to run that loop to read from the socket, because reading is blocking.
					// the send loop of the client is not that necessary.
					// however, we need to be able to dynamically add and remove clients, because the interfaces
					// might change.
					log.Info().Str("iface", iface.Name).Msg("starting client")
					// the context here should be specific to each client
					err := c.Run(clientCtx, &receivedCh)

					// remove the client from the hashtables
					func() {
						clientMutex.Lock()
						defer clientMutex.Unlock()

						delete(clients, iface.Name)
						delete(clientCancels, iface.Name)
					}()

					clientWg.Done()
					if err != nil {
						log.Error().Err(err).Msg("error while running client")
					}

					log.Info().Str("iface", iface.Name).Err(err).Msg("client stopped")
				}()
			}
		}

		for _, c := range clients {
			c.SendPing()
		}

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
				// here, we need to iterate over all the interfaces, create a new client if necessasry, and send a ping from it
				for _, c := range clients {
					c.SendPing()
				}

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
				log.Info().Msg("waiting for clients to stop")
				for _, cancel := range clientCancels {
					cancel()
				}
				clientWg.Wait()

				return ctx.Err()
			}
		}
	})

	return grp.Wait()
}
