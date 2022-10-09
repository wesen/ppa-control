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

func GetValidInterfaces() ([]net.Interface, error) {
	res := make([]net.Interface, 0)

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Error().Err(err).Msg("failed to get interfaces")
		return nil, err
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
			continue
		}

		isInterfaceValid := false

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

			log.Debug().
				Str("interface", iface.Name).
				Str("ip", ip.String()).
				Str("netmask", ip.DefaultMask().String()).
				Msg("found ip address")

			isInterfaceValid = true
		}

		if isInterfaceValid {
			res = append(res, iface)
		}
	}

	return res, nil
}
