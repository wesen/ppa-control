package utils

import (
	"github.com/rs/zerolog/log"
	"net"
)

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

			log.Trace().
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
