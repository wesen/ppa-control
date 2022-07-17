package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/augustoroman/hexdump"
	"io"
	"net"
	"ppa-control/lib/protocol"
	"time"
)

const MaxBufferSize = 1024
const Timeout = 10 * time.Second

type client struct {
	SendChannel    chan *bytes.Buffer
	ReceiveChannel chan *bytes.Buffer
	ComponentId    uint8
	seqCmd         uint16
}

func NewClient(componentId uint8) *client {
	return &client{
		SendChannel:    make(chan *bytes.Buffer),
		ReceiveChannel: make(chan *bytes.Buffer),
		ComponentId:    componentId,
		seqCmd:         1,
	}
}

func (c *client) SendPresetRecallByPresetIndex(index int) {
	buf := new(bytes.Buffer)
	bh := protocol.NewBasicHeader(
		protocol.MessageTypePresetRecall,
		protocol.StatusCommand,
		[4]byte{0, 0, 0, 0},
		c.seqCmd,
		byte(c.ComponentId),
	)
	pr := protocol.NewPresetRecall(protocol.RecallByPresetIndex, 0, byte(index))
	// XXX potentially need mutex here
	c.seqCmd += 1

	protocol.EncodeHeader(buf, bh)
	protocol.EncodePresetRecall(buf, pr)
	c.SendChannel <- buf
}

func (c *client) Run(ctx context.Context, address string) (err error) {
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return
	}
	defer conn.Close()

	go func() {

	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("client.go: cancelled")
			// XXX I guess we could close the socket here
			return ctx.Err()

		case buf := <-c.SendChannel:
			n, err := io.Copy(conn, buf)
			if err != nil {
				return err
			}
			fmt.Printf("client.go: packet-written: bytes=%d\n", n)
			buffer := make([]byte, MaxBufferSize)
			deadline := time.Now().Add(Timeout)
			err = conn.SetReadDeadline(deadline)
			if err != nil {
				return err
			}

			nRead, addr, err := conn.ReadFrom(buffer)
			if err != nil {
				return err
			}

			fmt.Printf("client.go: packet-received: bytes=%d from=%s\nclient.go: %s\n",
				nRead, addr.String(), hexdump.Dump(buffer[:nRead]))
		}
	}
}
