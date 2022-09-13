package utils

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
)

func GetLocalAddresses() ([]*net.IPAddr, error) {
	var res []*net.IPAddr

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not list interfaces")
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Error().Str("error", err.Error()).Msg("Could not get interface addresses")
			continue
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPAddr:
				res = append(res, v)
				log.Info().
					Str("interface", i.Name).
					Str("address", v.String()).
					Str("defaultMask", fmt.Sprintf("%x\n", v.IP.DefaultMask())).
					Msg("found interface")
			}
		}
	}

	return res, nil
}

func GetLocalMulticastAddresses() ([]*net.IPAddr, error) {
	var res []*net.IPAddr

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not list interfaces")
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Warn().Str("error", err.Error()).Msg("Could not get addresses for interface")
			continue
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPAddr:
				res = append(res, v)
				log.Info().
					Str("interface", i.Name).
					Str("address", v.String()).
					Str("defaultMask", fmt.Sprintf("%x\n", v.IP.DefaultMask())).
					Msg("found interface")
			}
		}
	}

	return res, nil
}
