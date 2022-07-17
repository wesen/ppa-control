package server

import (
	"context"
	"fmt"
	"github.com/augustoroman/hexdump"
	"net"
	"time"
)

const MaxBufferSize = 1024
const Timeout = 10 * time.Second

func RunServer(ctx context.Context, address string) (err error) {
	pc, err := net.ListenPacket("udp", address)
	if err != nil {
		return
	}
	defer pc.Close()

	doneChan := make(chan error, 1)
	buffer := make([]byte, MaxBufferSize)

	go func() {
		for {
			fmt.Printf("server: waiting\n")
			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				doneChan <- err
				return
			}

			fmt.Printf("server: packet-received: bytes=%d from=%s\n", n, addr.String())
			deadline := time.Now().Add(Timeout)
			err = pc.SetWriteDeadline(deadline)
			if err != nil {
				doneChan <- err
				return
			}

			n, err = pc.WriteTo(buffer[:n], addr)
			if err != nil {
				doneChan <- err
				return
			}
			fmt.Printf("server: packet-written: bytes=%d to=%s\nserver: %s\n",
				n, addr.String(),
				hexdump.Dump(buffer[:n]))
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Println("server: cancelled")
		err = ctx.Err()
	case err = <-doneChan:
		if err != nil {
			fmt.Printf("server: got error: %s\n", err)
		}
	}

	return
}
