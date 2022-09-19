package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/augustoroman/hexdump"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net"
	"os"
	"ppa-control/lib/protocol"
	"syscall"
	"time"
)

const MaxBufferSize = 1024
const Timeout = 10 * time.Second

type Client interface {
	Run(ctx context.Context, c *chan ReceivedMessage) error
	SendPresetRecallByPresetIndex(index int)
	SendPing()
	Name() string
}

type ReceivedMessage struct {
	Header        *protocol.BasicHeader
	RemoteAddress net.Addr
	Body          interface{}
	Data          []byte
	Client        Client
}

type SingleDevice struct {
	Address     string
	SendChannel chan *bytes.Buffer
	ComponentId int
	seqCmd      uint16
}

func NewClient(address string, componentId int) *SingleDevice {
	return &SingleDevice{
		SendChannel: make(chan *bytes.Buffer),
		Address:     address,
		ComponentId: componentId,
		seqCmd:      1,
	}
}

func (c *SingleDevice) SendPing() {
	buf := new(bytes.Buffer)
	bh := protocol.NewBasicHeader(
		protocol.MessageTypePing,
		protocol.StatusRequestServer,
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
	log.Debug().Str("address", c.Address).Msg("Sending ping to SendChannel")
	c.SendChannel <- buf
}

func (c *SingleDevice) SendPresetRecallByPresetIndex(index int) {
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

// Run is the main loop for the client.  It will listen for messages on the sendChannel
// and emit them on the UDP socket, and it will listen for incoming packets on the UDP socket,
// parse them and emit them on the receiveChannel.
func (c *SingleDevice) Run(ctx context.Context, receivedCh *chan ReceivedMessage) (err error) {
	raddr, err := net.ResolveUDPAddr("udp", c.Address)
	if err != nil {
		return
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return
	}
	defer conn.Close()

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return c.readLoop(ctx, conn, receivedCh)
	})

	grp.Go(func() error {
		return c.sendLoop(ctx, conn, raddr)
	})

	log.Info().Str("address", c.Address).Msg("Client started")
	return grp.Wait()
}

func (c *SingleDevice) sendLoop(ctx context.Context, conn *net.UDPConn, raddr *net.UDPAddr) error {
	log.Info().Str("address", c.Address).Msg("Starting send loop")
	defer func() {
		log.Info().Str("address", c.Address).Msg("Exiting send loop")
	}()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("SingleDevice.go: cancelled")
			return ctx.Err()

		case buf := <-c.SendChannel:
			go func() {
				log.Debug().Str("address", c.Address).Int("len", buf.Len()).Msg("Sending packet")
				deadline := time.Now().Add(Timeout)
				err := conn.SetWriteDeadline(deadline)
				if err != nil {
					log.Warn().
						Err(err).
						Str("address", c.Address).
						Msg("Failed to set write deadline")
				}
				n, err := conn.WriteToUDP(buf.Bytes(), raddr)
				if err != nil {
					log.Warn().
						Err(err).
						Str("to", c.Address).
						Str("local", conn.LocalAddr().String()).
						Int("length", buf.Len()).
						Bytes("data", buf.Bytes()).
						Msg("Failed to write to connection")
				} else {
					log.Debug().
						Str("to", c.Address).
						Str("from", conn.LocalAddr().String()).
						Int("length", buf.Len()).
						Int("written", n).
						Msg("Written packet")

				}
			}()
		}
	}
}

func (c *SingleDevice) readLoop(ctx context.Context, conn *net.UDPConn, receivedCh *chan ReceivedMessage) error {
	defer func() {
		log.Info().
			Str("address", conn.LocalAddr().String()).
			Msg("Exiting read loop")
	}()

	for {
		if ctx.Err() != nil {
			log.Debug().Msg("Context error, exiting read loop")
			return ctx.Err()
		}

		buffer := make([]byte, MaxBufferSize)

		log.Trace().
			Str("address", conn.LocalAddr().String()).
			Msg("Reading from connection")

		deadline := time.Now().Add(200 * time.Millisecond)
		err := conn.SetReadDeadline(deadline)
		if err != nil {
			log.Warn().
				Str("address", conn.LocalAddr().String()).
				Str("error", err.Error()).
				Msg("Failed to set read deadline")
			return err
		}

		nRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			switch v := err.(type) {
			case *net.OpError:
				if v.Timeout() {
					log.Trace().Msg("Read timeout")
					continue
				}
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
			return err
		}

		if zerolog.GlobalLevel() == zerolog.DebugLevel {
			fmt.Printf("%s\n", hexdump.Dump(buffer[:nRead]))
		}
		log.Info().Int("received", nRead).
			Str("from", addr.String()).
			Str("local", conn.LocalAddr().String()).
			Bytes("data", buffer[:nRead]).
			Msg("Received packet")

		hdr, err := protocol.ParseHeader(buffer[:nRead])

		if err != nil {
			log.Warn().Err(err).
				Bytes("payload", buffer[:nRead]).
				Msg("Could not decode incoming message")

			if receivedCh != nil {
				*receivedCh <- ReceivedMessage{
					Header:        nil,
					RemoteAddress: addr,
					Client:        c,
					Body:          nil,
					Data:          buffer[:nRead],
				}
			}
			continue
		}

		// XXX parse body further

		if receivedCh != nil {
			*receivedCh <- ReceivedMessage{
				Header:        hdr,
				RemoteAddress: addr,
				Client:        c,
				Body:          nil,
				Data:          buffer[:nRead],
			}
		}
	}
}

func (c *SingleDevice) Name() string {
	return fmt.Sprintf("SingleDevice-%s", c.Address)
}
