//go:build darwin

package simulation

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"net"
	"syscall"
)

func (sd *SimulatedDevice) bindToInterface(conn *net.UDPConn) error {
	if sd.Settings.Interface == "" {
		return nil
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	ifaceIdx := -1
	for idx, iface := range ifaces {
		if iface.Name == sd.Settings.Interface {
			ifaceIdx = idx
			break
		}
	}
	if ifaceIdx == -1 {
		return errors.Errorf("interface %s not found", sd.Settings.Interface)
	}

	c, err := conn.SyscallConn()
	if err != nil {
		panic(err)
	}
	var e2 error = nil
	err = c.Control(func(fd uintptr) {
		log.Info().
			Int("fd", int(fd)).
			Str("iface", sd.Settings.Interface).
			Int("ifaceIdx", ifaceIdx).
			Msg("Binding socket to interface")
		e2 = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_BOUND_IF, ifaceIdx)
		log.Debug().Err(e2).Msg("Called setsockoptint")
	})
	if e2 != nil {
		return e2
	}
	if err != nil {
		return err
	}

	return nil
}
