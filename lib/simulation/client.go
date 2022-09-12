package simulation

import (
	"bytes"
	"context"
	"fmt"
	"github.com/augustoroman/hexdump"
	"golang.org/x/sync/errgroup"
	"log"
	"net"
	"ppa-control/lib/protocol"
	"time"
)

const MaxBufferSize = 1024
const Timeout = 10 * time.Second

type client struct {
	SendChannel    chan *bytes.Buffer
	ReceiveChannel chan *bytes.Buffer
}

func NewClient() *client {
	return &client{
		SendChannel:    make(chan *bytes.Buffer),
		ReceiveChannel: make(chan *bytes.Buffer),
	}
}

func (c *client) Run(ctx context.Context, address string) (err error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return
	}
	defer conn.Close()

	deadline := time.Now().Add(Timeout)
	err = conn.SetWriteDeadline(deadline)
	err = conn.SetReadDeadline(deadline)
	if err != nil {
		return err
	}

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		defer func() {
			log.Printf("read-loop exiting\n")
		}()

		for {
			buffer := make([]byte, MaxBufferSize)

			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				return err
			}

			fmt.Printf("client.go: packet-received: bytes=%d from=%s\nclient.go: %s\n",
				n, addr.String(), hexdump.Dump(buffer[:n]))

			if n > 0 {

				switch protocol.MessageType(buffer[0]) {
				case protocol.MessageTypePing:
				case protocol.MessageTypeLiveCmd:
				case protocol.MessageTypeDeviceData:
				case protocol.MessageTypePresetRecall:
				case protocol.MessageTypePresetSave:
				case protocol.MessageTypeUnknown:
				}
				c.ReceiveChannel <- bytes.NewBuffer(buffer[:n])
			}
		}
	})

	grp.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("client.go: cancelled")

			case buf := <-c.SendChannel:
				// Send
				n, err := conn.Write(buf.Bytes())
				if err != nil {
					return err
				}
				fmt.Printf("client.go: packet-written: bytes=%d\n", n)
			}
		}
	})

	return grp.Wait()
}
