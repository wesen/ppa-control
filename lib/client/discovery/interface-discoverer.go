package discovery

import (
	"context"
	"github.com/rs/zerolog/log"
	"net"
	"ppa-control/lib/utils"
	"time"
)

type InterfaceDiscoverer struct {
	im                 *InterfaceManager
	acceptedInterfaces map[InterfaceName]struct{}
	addedInterfaceCh   chan string
	removedInterfaceCh chan string
}

func NewInterfaceDiscoverer(im *InterfaceManager, acceptedInterfaces []InterfaceName) *InterfaceDiscoverer {
	// TODO(@manuel) - The idiomatic way to do sets is a map of struct{}
	acceptedInterfacesMap := make(map[InterfaceName]struct{})
	for _, iface := range acceptedInterfaces {
		acceptedInterfacesMap[iface] = struct{}{}
	}
	return &InterfaceDiscoverer{
		im:                 im,
		acceptedInterfaces: acceptedInterfacesMap,
		addedInterfaceCh:   make(chan InterfaceName),
		removedInterfaceCh: make(chan InterfaceName),
	}
}

func (id *InterfaceDiscoverer) updateInterfaces(currentInterfaces []InterfaceName) (newInterfaces []InterfaceName, removedInterfaces []InterfaceName, err error) {
	validInterfaces, err := utils.GetValidInterfaces()
	if err != nil {
		log.Error().Err(err).Msg("failed to get interfaces")
		return nil, nil, err
	}
	log.Debug().Msgf("valid interfaces: %v", validInterfaces)

	newInterfaces = make([]InterfaceName, 0)
	removedInterfaces = make([]InterfaceName, 0)

	currentInterfacesMap := make(map[InterfaceName]struct{})
	for _, iface := range currentInterfaces {
		currentInterfacesMap[iface] = struct{}{}
	}
	// create a hashmap of validInterfaces
	validInterfacesMap := make(map[InterfaceName]net.Interface)
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
// we use the InterfaceManager to get the current list of clients
func (id *InterfaceDiscoverer) Run(ctx context.Context) error {
	scanInterfaces := func() error {
		log.Debug().Msg("scanning interfaces")

		clientInterfaces := id.im.GetClientInterfaces()

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
