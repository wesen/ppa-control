package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/augustoroman/hexdump"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"net"
	"ppa-control/lib/protocol"
	"time"
)

const MaxBufferSize = 1024
const Timeout = 10 * time.Second

type Client interface {
	Run(ctx context.Context) error
	SendPresetRecallByPresetIndex(index int)
	Ping()
}

type client struct {
	Address        string
	SendChannel    chan *bytes.Buffer
	ReceiveChannel chan *bytes.Buffer
	ComponentId    int
	seqCmd         uint16
}

type multiClient struct {
	clients []Client
}

func NewClient(address string, componentId int) *client {
	return &client{
		SendChannel:    make(chan *bytes.Buffer),
		ReceiveChannel: make(chan *bytes.Buffer),
		Address:        address,
		ComponentId:    componentId,
		seqCmd:         1,
	}
}

func NewMultiClient(clients []Client) *multiClient {
	return &multiClient{clients: clients}
}

func (c *client) Ping() {
	buf := new(bytes.Buffer)
	bh := protocol.NewBasicHeader(
		protocol.MessageTypePing,
		protocol.StatusCommandClient,
		[4]byte{0, 0, 0, 0},
		c.seqCmd,
		byte(c.ComponentId),
	)

	c.seqCmd++

	protocol.EncodeHeader(buf, bh)
	c.SendChannel <- buf
}

func (mc *multiClient) SendPing() {
	for _, c := range mc.clients {
		c.Ping()
	}
}

func (c *client) SendPresetRecallByPresetIndex(index int) {
	buf := new(bytes.Buffer)
	bh := protocol.NewBasicHeader(
		protocol.MessageTypePresetRecall,
		protocol.StatusCommandClient,
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

func (mc *multiClient) SendPresetRecallByPresetIndex(index int) {
	for _, c := range mc.clients {
		c.SendPresetRecallByPresetIndex(index)
	}
}

func (c *client) Run(ctx context.Context) (err error) {
	raddr, err := net.ResolveUDPAddr("udp", c.Address)
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return
	}
	defer conn.Close()

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		defer func() {
			log.Printf("read-loop exiting\n")
		}()
		for {
			buffer := make([]byte, MaxBufferSize)
			//deadline := time.Now().Add(Timeout)
			//err = conn.SetReadDeadline(deadline)
			//if err != nil {
			//	return err
			//}

			// Block on read
			nRead, addr, err := conn.ReadFrom(buffer)
			if err != nil {
				return err
			}

			fmt.Printf("client.go: packet-received: bytes=%d from=%s\nclient.go: %s\n",
				nRead, addr.String(), hexdump.Dump(buffer[:nRead]))
		}
	})

	grp.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("client.go: cancelled")
				// XXX I guess we could close the socket here
				conn.Close()
				return ctx.Err()

			case buf := <-c.SendChannel:
				// Send
				n, err := io.Copy(conn, buf)
				if err != nil {
					return err
				}
				fmt.Printf("client.go: packet-written: bytes=%d\n", n)
			}
		}
	})

	return grp.Wait()
}

func (c *multiClient) Run(ctx context.Context) (err error) {
	grp, ctx := errgroup.WithContext(ctx)

	for _, c := range c.clients {
		grp.Go(func() error {
			return c.Run(ctx)
		})
	}
	return grp.Wait()
}
