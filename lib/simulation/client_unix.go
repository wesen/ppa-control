//go:build linux

package simulation

import (
	"context"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
	"net"
	"syscall"
)

func ListenUDPBroadcast(ctx context.Context, addr string, iface string) (net.PacketConn, error) {
	lc := &net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var e2 error = nil
			err := c.Control(func(fd uintptr) {
				log.Info().
					Int("fd", int(fd)).
					Msg("Binding socket to interface")
				e2 = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
				if e2 != nil {
					log.Error().
						Err(e2).
						Msg("Failed to set SO_REUSEADDR")
					return
				} else {
					log.Debug().
						Msg("Set SO_REUSEADDR")
				}
				e2 = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				if e2 != nil {
					log.Error().
						Err(e2).
						Msg("Failed to set SO_REUSEPORT")
					return
				} else {
					log.Debug().
						Msg("Set SO_REUSEPORT")
				}
				if iface != "" {
					log.Debug().
						Str("interface", iface).
						Msg("Binding to interface")
					e2 = syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, iface)
					if e2 != nil {
						log.Error().
							Err(e2).
							Str("interface", iface).
							Msg("Failed to BINDTODEVICE")

					}
					log.Debug().Err(e2).Msg("Called setsockoptstring")
				}
				e2 = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
				if e2 != nil {
					log.Error().
						Err(e2).
						Msg("Failed to set SO_BROADCAST")
					return
				} else {
					log.Debug().
						Msg("Set SO_BROADCAST")
				}
			})
			if err != nil {
				return err
			}
			if e2 != nil {
				return e2
			}

			return nil
		},
	}

	c, err := lc.ListenPacket(ctx, "udp4", addr)
	if err != nil {
		return nil, err
	}

	return c, nil
}
