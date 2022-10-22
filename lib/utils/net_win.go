//go:build windows

package utils

import "net"
import "context"

func ListenUDP(ctx context.Context, addr string, iface string) (net.PacketConn, error) {
	panic("not implemented")
}
