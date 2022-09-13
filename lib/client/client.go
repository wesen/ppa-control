package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/augustoroman/hexdump"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"io"
	"net"
	"os"
	"ppa-control/lib/protocol"
	"syscall"
	"time"
)

const MaxBufferSize = 1024
const Timeout = 10 * time.Second

type Client interface {
	Run(ctx context.Context) error
	SendPresetRecallByPresetIndex(index int)
	SendPing()
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

func (c *client) SendPing() {
	buf := new(bytes.Buffer)
	bh := protocol.NewBasicHeader(
		protocol.MessageTypePing,
		protocol.StatusCommandClient,
		[4]byte{0, 0, 0, 0},
		c.seqCmd,
		byte(c.ComponentId),
	)

	c.seqCmd++

	err := protocol.EncodeHeader(buf, bh)
	if err != nil {
		log.Warn().Str("error", err.Error()).Msg("Failed to encode header")
		return
	}
	c.SendChannel <- buf
}

func (mc *multiClient) SendPing() {
	for _, c := range mc.clients {
		c.SendPing()
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

	err := protocol.EncodeHeader(buf, bh)
	if err != nil {
		log.Warn().Str("error", err.Error()).Msg("Failed to encode header")
		return
	}
	err = protocol.EncodePresetRecall(buf, pr)
	if err != nil {
		log.Warn().Str("error", err.Error()).Msg("Failed to encode preset recall")
		return
	}
	log.Debug().Str("address", c.Address).Msg("Sending preset recall")
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

	deadline := time.Now().Add(Timeout)
	err = conn.SetWriteDeadline(deadline)
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
			// Block on read
			nRead, addr, err := conn.ReadFrom(buffer)
			if err != nil {
				switch v := err.(type) {
				case *net.OpError:
					switch v2 := v.Err.(type) {
					case *os.SyscallError:
						if v2.Syscall == "recvfrom" && v2.Err == syscall.ECONNREFUSED {
							log.Warn().Msg("Connection refused")
						}
					default:
					}
				default:
				}
				log.Warn().Err(err).Msg("Failed to read from connection")
				// ignore errors anyway
			}

			log.Info().Int("received", nRead).
				Str("from", addr.String()).
				Str("buffer", hexdump.Dump(buffer[:nRead])).
				Msg("Received packet")
		}
	})

	grp.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("client.go: cancelled")
				return ctx.Err()

			case buf := <-c.SendChannel:
				// Send
				n, err := io.Copy(conn, buf)
				if err != nil {
					log.Error().Err(err).Msg("Failed to write to connection")
					return err
				}
				log.Info().
					Int64("written", n).
					Msg("Written packet")
			}
		}
	})

	log.Info().Str("address", c.Address).Msg("Client started")
	return grp.Wait()
}

func (c *multiClient) Run(ctx context.Context) (err error) {
	grp, ctx := errgroup.WithContext(ctx)

	for _, c2 := range c.clients {
		grp.Go(func() error {
			return c2.Run(ctx)
		})
	}
	return grp.Wait()
}
