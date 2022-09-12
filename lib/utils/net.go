package utils

import (
	"net"
	logger "ppa-control/lib/log"
)

func GetLocalAddresses() ([]*net.IPAddr, error) {
	var res []*net.IPAddr

	ifaces, err := net.Interfaces()
	if err != nil {
		logger.Sugar.Error("error", err.Error())
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			logger.Sugar.Warn("error", err.Error())
			continue
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPAddr:
				res = append(res, v)
				logger.Sugar.Info(
					"interface", i.Name,
					"address", v.String(),
					"defaultMask", v.IP.DefaultMask())
			}
		}
	}

	return res, nil
}

func GetLocalMulticastAddresses() ([]*net.IPAddr, error) {
	var res []*net.IPAddr

	ifaces, err := net.Interfaces()
	if err != nil {
		logger.Sugar.Error("error", err.Error())
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.MulticastAddrs()
		if err != nil {
			logger.Sugar.Warn("error", err.Error())
			continue
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPAddr:
				res = append(res, v)
				logger.Sugar.With(
					"interface", i.Name,
					"address", v.String(),
					"defaultMask", v.IP.DefaultMask()).Info("found interface")
			}
		}
	}

	return res, nil
}
