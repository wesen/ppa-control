package client

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net"
	"ppa-control/lib/utils"
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

type interfaceDiscoverer struct {
	im                 *interfaceManager
	acceptedInterfaces map[string]bool
	addedInterfaceCh   chan string
	removedInterfaceCh chan string
}

func newInterfaceDiscoverer(im *interfaceManager, acceptedInterfaces []string) *interfaceDiscoverer {
	acceptedInterfacesMap := make(map[string]bool)
	for _, iface := range acceptedInterfaces {
		acceptedInterfacesMap[iface] = true
	}
	return &interfaceDiscoverer{
		im:                 im,
		acceptedInterfaces: acceptedInterfacesMap,
		addedInterfaceCh:   make(chan string),
		removedInterfaceCh: make(chan string),
	}
}

func (id *interfaceDiscoverer) updateInterfaces(currentInterfaces []string) (newInterfaces []string, removedInterfaces []string, err error) {
	validInterfaces, err := utils.GetValidInterfaces()
	if err != nil {
		log.Error().Err(err).Msg("failed to get interfaces")
		return nil, nil, err
	}
	log.Debug().Msgf("valid interfaces: %v", validInterfaces)

	newInterfaces = make([]string, 0)
	removedInterfaces = make([]string, 0)

	currentInterfacesMap := make(map[string]bool)
	for _, iface := range currentInterfaces {
		currentInterfacesMap[iface] = true
	}
	// create a hashmap of validInterfaces
	validInterfacesMap := make(map[string]net.Interface)
	for _, iface := range validInterfaces {
		validInterfacesMap[iface.Name] = iface
	}

	for ifaceName := range currentInterfacesMap {
		if _, ok := validInterfacesMap[ifaceName]; !ok {
			removedInterfaces = append(removedInterfaces, ifaceName)
		}
	}

	for _, iface := range validInterfaces {
		if len(id.acceptedInterfaces) > 0 {
			if _, ok := id.acceptedInterfaces[iface.Name]; !ok {
				// not an accepted interface
				continue
			}
		}
		if _, ok := currentInterfacesMap[iface.Name]; ok {
			// already know interface
			continue
		}

		newInterfaces = append(newInterfaces, iface.Name)
	}

	return
}

// Run will discover new and removed interfaces.
// we use the interfaceManager to get the current list of clients
func (id *interfaceDiscoverer) Run(ctx context.Context) error {
	scanInterfaces := func() error {
		log.Debug().Msg("scanning interfaces")

		clientInterfaces := id.im.getClientInterfaces()

		log.Debug().Msgf("current interfaces: %v", clientInterfaces)
		newInterfaces, removedInterfaces, err := id.updateInterfaces(clientInterfaces)
		if err != nil {
			log.Error().Err(err).Msg("failed to update interfaces")
			return err
		}

		// use non-blocking primitives to avoid hanging if the main loop gets cancelled
		for _, iface := range newInterfaces {
			log.Debug().Str("iface", iface).Msg("adding interface")
			select {
			case id.addedInterfaceCh <- iface:
				log.Debug().Str("iface", iface).Msg("added interface")
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		for _, iface := range removedInterfaces {
			log.Debug().Str("iface", iface).Msg("removing interface")
			select {
			case id.removedInterfaceCh <- iface:
				log.Debug().Str("iface", iface).Msg("removed interface")
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	}

	err := scanInterfaces()
	if err != nil {
		return err
	}

	for {
		t := time.NewTimer(5 * time.Second)

		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-t.C:
			err := scanInterfaces()
			if err != nil {
				return err
			}
		}
	}
}

type interfaceManager struct {
	port       uint16
	receivedCh *chan ReceivedMessage

	// used to wait for all clients to be done
	wg sync.WaitGroup

	// one Client and cancel per interface, protected by mutex
	mutex   sync.RWMutex
	clients map[string]Client
	cancels map[string]context.CancelFunc
}

func newClientMap(port uint16, receivedCh *chan ReceivedMessage) *interfaceManager {
	return &interfaceManager{
		clients:    make(map[string]Client),
		cancels:    make(map[string]context.CancelFunc),
		receivedCh: receivedCh,
		port:       port,
	}
}

func (im *interfaceManager) sendPing() {
	im.mutex.RLock()
	defer func() {
		log.Trace().Msg("unlocking sendPing mutex")
		im.mutex.RUnlock()
	}()

	for _, client := range im.clients {
		client.SendPing()
	}
}

func (im *interfaceManager) doesInterfaceExist(iface string) bool {
	im.mutex.RLock()
	defer func() {
		log.Trace().Msg("unlocking doesInterfaceExist mutex")
		defer im.mutex.RUnlock()
	}()

	_, ok := im.clients[iface]
	return ok
}

func (im *interfaceManager) addInterface(ctx context.Context, iface string) error {
	log.Debug().Str("iface", iface).Msg("adding interface")
	if im.doesInterfaceExist(iface) {
		log.Error().Str("iface", iface).Msg("interface already exists")
		return fmt.Errorf("interface %s already exists", iface)
	}

	log.Debug().Str("iface", iface).Msg("creating client")

	broadcastAddr := fmt.Sprintf("255.255.255.255:%d", im.port)

	// now, create a client bound to the interface with the address being 255.255.255.255
	c := NewClient(broadcastAddr, 0xfe)

	clientCtx, cancel := context.WithCancel(ctx)
	log.Debug().Str("iface", iface).Msg("adding client")
	func() {
		im.mutex.Lock()
		defer func() {
			log.Trace().Str("iface", iface).Msg("unlocking addInterface mutex")
			im.mutex.Unlock()
		}()

		im.clients[iface] = c
		im.cancels[iface] = cancel
	}()

	log.Debug().Str("iface", iface).Msg("starting client")

	go func() {
		im.wg.Add(1)

		log.Info().Str("iface", iface).Msg("starting client")
		err := c.Run(clientCtx, im.receivedCh)

		// remove the client from the hashtables once done
		func() {
			im.mutex.Lock()
			defer im.mutex.Unlock()

			delete(im.clients, iface)
			delete(im.cancels, iface)
		}()

		im.wg.Done()
		if err != nil {
			log.Error().Err(err).Msg("error while running client")
		}

		log.Info().Str("iface", iface).Err(err).Msg("client stopped")
	}()

	return nil
}

func (im *interfaceManager) removeInterface(iface string) error {
	if !im.doesInterfaceExist(iface) {
		return fmt.Errorf("interface %s does not exist", iface)
	}

	im.mutex.RLock()
	defer im.mutex.Unlock()
	im.cancels[iface]()

	return nil
}

func (im *interfaceManager) wait() {
	log.Debug().Msg("waiting for all clients to be done")
	im.wg.Wait()
	log.Debug().Msg("all clients are done")
}

func (im *interfaceManager) getClientInterfaces() []string {
	im.mutex.RLock()
	defer func() {
		log.Trace().Msg("unlocking getClientInterfaces mutex")
		im.mutex.RUnlock()
	}()

	var interfaces []string
	for iface := range im.clients {
		interfaces = append(interfaces, iface)
	}

	return interfaces
}

func Discover(ctx context.Context, msgCh chan PeerInformation, discoveryInterfaces []string, port uint16) error {
	receivedCh := make(chan ReceivedMessage)

	clientMap := newClientMap(port, &receivedCh)
	id := newInterfaceDiscoverer(clientMap, discoveryInterfaces)

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return id.Run(ctx)
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
						msgCh <- PeerLost{addr: addr}
						delete(peerLastSeen, addr)
					}
				}

				clientMap.sendPing()

			case newInterface := <-id.addedInterfaceCh:
				log.Debug().Str("iface", newInterface).Msg("new interface discovered")
				err := clientMap.addInterface(ctx, newInterface)
				if err != nil {
					return err
				}

			case removedInterface := <-id.removedInterfaceCh:
				log.Debug().Str("iface", removedInterface).Msg("interface removed")
				err := clientMap.removeInterface(removedInterface)
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
				clientMap.wait()

				return ctx.Err()
			}
		}
	})

	return grp.Wait()
}
