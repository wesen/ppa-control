//go:build darwin

package simulation

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"net"
	"syscall"
)

func ListenUDPBroadcast(ctx context.Context, addr string, iface string) (net.PacketConn, error) {
	lc := &net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var e2 error = nil
			err := c.Control(func(fd uintptr) {
				if iface != "" {
					ifaces, err := net.Interfaces()
					if err != nil {
						e2 = err
						return
					}
					ifaceIdx := -1
					for idx, iface_ := range ifaces {
						if iface_.Name == iface {
							ifaceIdx = idx
							break
						}
					}
					if ifaceIdx == -1 {
						e2 = errors.Errorf("interface %s not found", iface)
						return
					}

					log.Debug().Str("interface", iface).Msg("Binding to interface")
					e2 = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_BOUND_IF, ifaceIdx)
					if e2 != nil {
						log.Error().
							Err(e2).
							Int("fd", int(fd)).
							Str("interface", iface).
							Msg("Error binding to interface")
						return
					}
					log.Debug().
						Str("interface", iface).
						Msg("Bound to interface")
				}

				log.Info().
					Int("fd", int(fd)).
					Msg("Binding socket to interface")
				e2 = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
				if e2 != nil {
					log.Error().Err(e2).Msg("Error setting SO_REUSEADDR")
					return
				} else {
					log.Debug().Msg("Set SO_REUSEADDR")
				}
				e2 = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
				if e2 != nil {
					log.Error().Err(e2).Msg("Error setting SO_REUSEPORT")
					return
				} else {
					log.Debug().Msg("Set SO_REUSEPORT")
				}

				e2 = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
				if e2 != nil {
					log.Error().Err(e2).Msg("Error setting SO_BROADCAST")
					return
				} else {
					log.Debug().Msg("Set SO_BROADCAST")
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
		return nil, errors.Wrap(err, "failed to listen on UDP broadcast")
	}

	return c, nil
}
