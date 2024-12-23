package client

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"ppa-control/lib/protocol"
	"ppa-control/lib/utils"
	"syscall"
	"time"

	"github.com/augustoroman/hexdump"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const MaxBufferSize = 1024
const Timeout = 10 * time.Second

// A pkg has multiple target addresses, and no source address (?).
//That actually won't work either, will it...

// A pkg should have a list of targeted *UDPAddr, to avoid having to pass strings with ports around (?)
// I actually think the current structure is not too bad for the clients...
// We just need to have the multiclient be able to add and remove clients (?)
//
// What about the discovery loop? We also need to recognize new interfaces...

type ReceivedMessage struct {
	Header        *protocol.BasicHeader
	RemoteAddress net.Addr
	Interface     string
	Body          interface{}
	Data          []byte
	Client        Client
}

type SingleDevice struct {
	AddrPort    string
	Interface   string
	SendChannel chan *bytes.Buffer
	ComponentId uint
	seqCmd      uint16
}

func NewSingleDevice(address string, iface string, componentId uint) *SingleDevice {
	return &SingleDevice{
		// This channel is buffered to avoid blocking senders.
		SendChannel: make(chan *bytes.Buffer, 10),
		Interface:   iface,
		AddrPort:    address,
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
	log.Debug().
		Str("address", c.AddrPort).
		Str("interface", c.Interface).
		Int("length", buf.Len()).
		Msg("Sending ping to SendChannel")
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
	// TODO potentially need mutex here
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
	log.Debug().
		Str("address", c.AddrPort).
		Str("interface", c.Interface).
		Int("length", buf.Len()).
		Msg("Sending preset recall")
	c.SendChannel <- buf
}

func (c *SingleDevice) SendMasterVolume(volume float32) {
	buf := new(bytes.Buffer)
	bh := protocol.NewBasicHeader(
		protocol.MessageTypeDeviceData,
		protocol.StatusCommandClient,
		[4]byte{0, 0, 0, 0},
		c.seqCmd,
		byte(c.ComponentId),
	)
	// volume = 1 -> 0dB
	// volume = 0 -> -72dB

	twentyDB := 0x3e8
	minusEigthyDB := 0x00
	gain := uint32(volume * float32(twentyDB-minusEigthyDB))

	// TODO potentially need mutex here
	c.seqCmd += 1

	err := protocol.EncodeHeader(buf, bh)
	if err != nil {
		log.Warn().Str("error", err.Error()).Msg("Failed to encode header")
		return
	}

	err = binary.Write(buf, binary.LittleEndian, []int8{01, 00, 03, 06})
	if err != nil {
		log.Warn().Str("error", err.Error()).Msg("Failed to encode header")
		return
	}

	err = binary.Write(buf, binary.LittleEndian, gain)
	if err != nil {
		log.Warn().Str("error", err.Error()).Msg("Failed to encode header")
		return
	}

	log.Debug().
		Str("address", c.AddrPort).
		Str("interface", c.Interface).
		Int("length", buf.Len()).
		Msg("Sending master volume")

	fmt.Printf("%s\n", hexdump.Dump(buf.Bytes()[:buf.Len()]))
	c.SendChannel <- buf
}

// Run is the main loop for the client. It will listen for messages on the sendChannel
// and emit them on the UDP socket, and it will listen for incoming packets on the UDP socket,
// parse them and emit them on the receiveChannel.
func (c *SingleDevice) Run(ctx context.Context, receivedCh chan<- ReceivedMessage) (err error) {
	raddr, err := net.ResolveUDPAddr("udp", c.AddrPort)
	if err != nil {
		return
	}
	conn, err := utils.ListenUDP(ctx, "0.0.0.0:0", c.Interface)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return c.readLoop(ctx, conn, receivedCh)
	})

	grp.Go(func() error {
		return c.sendLoop(ctx, conn, raddr)
	})

	log.Info().Str("address", c.AddrPort).Msg("Client started")
	return grp.Wait()
}

func (c *SingleDevice) sendLoop(ctx context.Context, conn net.PacketConn, raddr *net.UDPAddr) error {
	log.Info().Str("address", c.AddrPort).Msg("Starting send loop")
	defer func() {
		log.Info().Str("address", c.AddrPort).Msg("Exiting send loop")
	}()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("SingleDevice.go: cancelled")
			return ctx.Err()

		case buf := <-c.SendChannel:
			go func() {
				log.Debug().Str("address", c.AddrPort).
					Int("len", buf.Len()).
					Msg("Sending packet")
				deadline := time.Now().Add(Timeout)
				err := conn.SetWriteDeadline(deadline)
				if err != nil {
					log.Warn().
						Err(err).
						Str("address", c.AddrPort).
						Msg("Failed to set write deadline")
				}
				n, err := conn.WriteTo(buf.Bytes(), raddr)
				if err != nil {
					log.Warn().
						Err(err).
						Str("to", c.AddrPort).
						Str("local", conn.LocalAddr().String()).
						Int("length", buf.Len()).
						Bytes("data", buf.Bytes()).
						Msg("Failed to write to connection")
				} else {
					log.Debug().
						Str("to", c.AddrPort).
						Str("from", conn.LocalAddr().String()).
						Int("length", buf.Len()).
						Int("written", n).
						Msg("Written packet")
				}
			}()
		}
	}
}

func (c *SingleDevice) readLoop(ctx context.Context, conn net.PacketConn, receivedCh chan<- ReceivedMessage) error {
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
			Str("iface", c.Interface).
			Str("local", conn.LocalAddr().String()).
			Bytes("data", buffer[:nRead]).
			Msg("Received packet")

		hdr, err := protocol.ParseHeader(buffer[:nRead])

		if err != nil {
			log.Warn().Err(err).
				Bytes("payload", buffer[:nRead]).
				Msg("Could not decode incoming message")

			if receivedCh != nil {
				receivedCh <- ReceivedMessage{
					Header:        nil,
					Interface:     c.Interface,
					RemoteAddress: addr,
					Client:        c,
					Body:          nil,
					Data:          buffer[:nRead],
				}
			}
			continue
		}

		// TODO parse body further

		if receivedCh != nil {
			receivedCh <- ReceivedMessage{
				Header:        hdr,
				RemoteAddress: addr,
				Interface:     c.Interface,
				Client:        c,
				Body:          nil,
				Data:          buffer[:nRead],
			}
		}
	}
}

func (c *SingleDevice) Name() string {
	return fmt.Sprintf("SingleDevice-%s", c.AddrPort)
}
