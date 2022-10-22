package discovery

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"go.uber.org/atomic"
	"ppa-control/lib/client"
	"sync"
)

type InterfaceName = string

type InterfaceManager struct {
	port       uint16
	receivedCh *chan client.ReceivedMessage

	// used to Wait for all clients to be done
	wg sync.WaitGroup

	// one Client and cancel per interface, protected by mutex
	mutex   sync.RWMutex
	clients map[InterfaceName]client.Client
	cancels map[InterfaceName]context.CancelFunc

	waiting atomic.Bool
}

func NewInterfaceManager(port uint16, receivedCh *chan client.ReceivedMessage) *InterfaceManager {
	return &InterfaceManager{
		clients:    make(map[InterfaceName]client.Client),
		cancels:    make(map[InterfaceName]context.CancelFunc),
		receivedCh: receivedCh,
		port:       port,
		waiting:    *atomic.NewBool(false),
	}
}

func (im *InterfaceManager) SendPing() {
	im.mutex.RLock()
	defer func() {
		log.Trace().Msg("unlocking SendPing mutex")
		im.mutex.RUnlock()
	}()

	for _, c := range im.clients {
		c.SendPing()
	}
}

func (im *InterfaceManager) DoesInterfaceExist(iface InterfaceName) bool {
	im.mutex.RLock()
	defer func() {
		log.Trace().Msg("unlocking DoesInterfaceExist mutex")
		defer im.mutex.RUnlock()
	}()

	_, ok := im.clients[iface]
	return ok
}

// StartInterfaceClient will create and start a client for the given interface,
// with target address the broadcast address 255.255.255.255 .
func (im *InterfaceManager) StartInterfaceClient(ctx context.Context, iface InterfaceName) (error, *client.SingleDevice) {
	if im.waiting.Load() {
		panic("cannot add interface while waiting for clients to be done")
	}

	log.Debug().Str("iface", iface).Msg("adding interface")
	if im.DoesInterfaceExist(iface) {
		log.Error().Str("iface", iface).Msg("interface already exists")
		return fmt.Errorf("interface %s already exists", iface), nil
	}

	log.Debug().Str("iface", iface).Msg("creating client")

	broadcastAddr := fmt.Sprintf("255.255.255.255:%d", im.port)

	// now, create a client bound to the interface with the address being 255.255.255.255
	c := client.NewSingleDevice(broadcastAddr, 0xfe)

	clientCtx, cancel := context.WithCancel(ctx)
	log.Debug().Str("iface", iface).Msg("adding client")
	func() {
		im.mutex.Lock()
		defer func() {
			log.Trace().Str("iface", iface).Msg("unlocking StartInterfaceClient mutex")
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

	return nil, c
}

// CancelInterfaceClient will cancel the client for the given interface.
// The interface client will be removed from the list of interfaces once it
// is done.
func (im *InterfaceManager) CancelInterfaceClient(iface string) error {
	if !im.DoesInterfaceExist(iface) {
		return fmt.Errorf("interface %s does not exist", iface)
	}

	im.mutex.RLock()
	defer im.mutex.RUnlock()
	im.cancels[iface]()

	return nil
}

// Wait waits for all clients to be done.
func (im *InterfaceManager) Wait() {
	// sanity check to make sure we are not adding or removing clients while waiting
	im.waiting.Store(true)

	log.Debug().Msg("waiting for all clients to be done")
	// Underlying rule that Add has to be invoked before Wait,
	// so how do we guarantee this
	im.wg.Wait()
	log.Debug().Msg("all clients are done")
}

// GetClientInterfaces returns a list of the names of the interfaces that have a client.
// This also returns the names of interfaces that have been cancelled but haven't yet been removed.
func (im *InterfaceManager) GetClientInterfaces() []string {
	im.mutex.RLock()
	defer func() {
		log.Trace().Msg("unlocking GetClientInterfaces mutex")
		im.mutex.RUnlock()
	}()

	var interfaces []string
	for iface := range im.clients {
		interfaces = append(interfaces, iface)
	}

	return interfaces
}
