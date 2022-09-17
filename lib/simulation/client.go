package simulation

import (
	"bytes"
	"context"
	"fmt"
	"github.com/augustoroman/hexdump"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net"
	"net/netip"
	"ppa-control/lib/protocol"
	"syscall"
	"time"
)

const MaxBufferSize = 1024
const Timeout = 10 * time.Second

type Preset struct {
}

type Response struct {
	Buffer   *bytes.Buffer
	AddrPort netip.AddrPort
}

type Request struct {
	Buffer   *bytes.Buffer
	AddrPort netip.AddrPort
}

type SimulatedDevice struct {
	SendChannel           chan Response
	ReceiveChannel        chan *bytes.Buffer
	deviceUniqueId        [4]byte
	componentId           byte
	address               string
	name                  string
	presets               []Preset
	currentlyActivePreset int
	currentVolume         float32
}

func NewClient(address string, name string, deviceUniqueId [4]byte, componentId byte) *SimulatedDevice {
	return &SimulatedDevice{
		SendChannel:           make(chan Response),
		ReceiveChannel:        make(chan *bytes.Buffer),
		deviceUniqueId:        deviceUniqueId,
		componentId:           componentId,
		address:               address,
		name:                  name,
		currentlyActivePreset: 0,
		currentVolume:         0.0,
	}
}

func (c *SimulatedDevice) Run(ctx context.Context) (err error) {
	addr, err := net.ResolveUDPAddr("udp", c.address)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	rawConn, err := conn.SyscallConn()
	if err != nil {
		return err
	}
	err = rawConn.Control(func(fd uintptr) {
		err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
		if err != nil {
			log.Error().Err(err).
				Msg("Could not set socket option broadcast")
		}
	})
	if err != nil {
		return err
	}

	// TODO add proper passing of interface name by using a settings builder pattern
	ifname := ""

	if ifname != "" {
		c, err := conn.SyscallConn()
		if err != nil {
			panic(err)
		}
		err = c.Control(func(fd uintptr) {
			fmt.Printf("Binding socket %d to interface %s\n", fd, ifname)
			err = syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, ifname)
			if err != nil {
				panic(err)
			}
		})
		if err != nil {
			panic(err)
		}
	}

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		defer func() {
			log.Info().Msg("read-loop exiting\n")
		}()

		for {
			buffer := make([]byte, MaxBufferSize)

			log.Debug().
				Str("address", conn.LocalAddr().String()).
				Msg("Waiting for data")

			//deadline := time.Now().Add(Timeout)
			//err = conn.SetReadDeadline(deadline)
			//if err != nil {
			//	log.Warn().
			//		Err(err).
			//		Msg("Could not set read deadline")
			//}

			n, srcAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.Error().Str("error", err.Error()).Msg("Could not read from UDP")
				return err
			}

			log.Info().Int("received", n).
				Str("from", srcAddr.String()).
				Str("local", conn.LocalAddr().String()).
				Msg("Received packet")
			fmt.Printf("%s\n", hexdump.Dump(buffer[:n]))

			request := &Request{
				Buffer:   bytes.NewBuffer(buffer[:n]),
				AddrPort: srcAddr.AddrPort(),
			}

			if n > 0 {

				switch protocol.MessageType(buffer[0]) {
				case protocol.MessageTypePing:
					{
						err := c.handlePing(request)
						if err != nil {
							log.Warn().Str("error", err.Error()).Msg("Could not handle ping")
						}
					}
				case protocol.MessageTypeLiveCmd:
					{
						err := c.handleLiveCmd(request)
						if err != nil {
							log.Warn().Str("error", err.Error()).Msg("Could not handle live command")
						}
					}
				case protocol.MessageTypeDeviceData:
					{
						err := c.handleDeviceData(request)
						if err != nil {
							log.Warn().Str("error", err.Error()).Msg("Could not handle device data")
						}
					}
				case protocol.MessageTypePresetRecall:
					{
						err := c.handlePresetRecall(request)
						if err != nil {
							log.Warn().Str("error", err.Error()).Msg("Could not handle preset recall")
						}
					}
				case protocol.MessageTypePresetSave:
					{
						err := c.handlePresetSave(request)
						if err != nil {
							log.Warn().Str("error", err.Error()).Msg("Could not handle preset save")
						}
					}
				case protocol.MessageTypeUnknown:
				}
			}
		}
	})

	grp.Go(func() error {
		defer func() {
			log.Info().Msg("write-loop exiting\n")
		}()
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("write-loop exiting")
				return ctx.Err()

			case response := <-c.SendChannel:
				// Send
				log.Debug().
					Int("len", response.Buffer.Len()).
					Bytes("data", response.Buffer.Bytes()).
					Msg("Sending packet")

				deadline := time.Now().Add(Timeout)
				err = conn.SetWriteDeadline(deadline)
				if err != nil {
					return err
				}
				n, err := conn.WriteToUDPAddrPort(response.Buffer.Bytes(), response.AddrPort)
				if err != nil {
					log.Warn().
						Err(err).
						Str("to", response.AddrPort.String()).
						Str("local", conn.LocalAddr().String()).
						Int("length", response.Buffer.Len()).
						Bytes("data", response.Buffer.Bytes()).
						Msg("Could not write to UDP")
				} else {
					log.Info().
						Str("to", response.AddrPort.String()).
						Str("local", conn.LocalAddr().String()).
						Int("length", response.Buffer.Len()).
						Bytes("data", response.Buffer.Bytes()).
						Int("written", n).
						Msg("Wrote packet")
				}
			}
		}
	})

	err = grp.Wait()
	log.Info().Msg("Exiting SimulatedDevice loop")
	return err
}

func (c *SimulatedDevice) handlePing(req *Request) error {
	hdr, err := protocol.ParseHeader(req.Buffer.Bytes())
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not parse ping header")
		return err
	}

	response := protocol.NewBasicHeader(
		protocol.MessageTypePing,
		protocol.StatusResponseServer,
		c.deviceUniqueId,
		hdr.SequenceNumber,
		c.componentId)

	buf := new(bytes.Buffer)
	err = protocol.EncodeHeader(buf, response)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not encode header")
		return err
	}

	c.SendChannel <- Response{
		Buffer:   buf,
		AddrPort: req.AddrPort,
	}

	return nil
}

func (c *SimulatedDevice) handleLiveCmd(req *Request) error {
	hdr, err := protocol.ParseHeader(req.Buffer.Bytes())
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not parse ping header")
		return err
	}

	response := protocol.NewBasicHeader(
		protocol.MessageTypePing,
		protocol.StatusResponseServer,
		c.deviceUniqueId,
		hdr.SequenceNumber,
		c.componentId)

	buf := new(bytes.Buffer)
	err = protocol.EncodeHeader(buf, response)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not encode header")
		return err
	}

	c.SendChannel <- Response{
		Buffer:   buf,
		AddrPort: req.AddrPort,
	}

	return nil
}

func (c *SimulatedDevice) handleDeviceData(req *Request) error {
	hdr, err := protocol.ParseHeader(req.Buffer.Bytes())
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not parse ping header")
		return err
	}

	response := protocol.NewBasicHeader(
		protocol.MessageTypePing,
		protocol.StatusResponseServer,
		c.deviceUniqueId,
		hdr.SequenceNumber,
		c.componentId)

	buf := new(bytes.Buffer)
	err = protocol.EncodeHeader(buf, response)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not encode header")
		return err
	}

	c.SendChannel <- Response{
		Buffer:   buf,
		AddrPort: req.AddrPort,
	}

	return nil

}

func (c *SimulatedDevice) handlePresetRecall(req *Request) error {
	hdr, err := protocol.ParseHeader(req.Buffer.Bytes())
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not parse ping header")
		return err
	}

	response := protocol.NewBasicHeader(
		protocol.MessageTypePing,
		protocol.StatusResponseServer,
		c.deviceUniqueId,
		hdr.SequenceNumber,
		c.componentId)

	buf := new(bytes.Buffer)
	err = protocol.EncodeHeader(buf, response)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not encode header")
		return err
	}

	c.SendChannel <- Response{
		Buffer:   buf,
		AddrPort: req.AddrPort,
	}

	return nil

}

func (c *SimulatedDevice) handlePresetSave(req *Request) error {
	hdr, err := protocol.ParseHeader(req.Buffer.Bytes())
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not parse ping header")
		return err
	}

	response := protocol.NewBasicHeader(
		protocol.MessageTypePing,
		protocol.StatusResponseServer,
		c.deviceUniqueId,
		hdr.SequenceNumber,
		c.componentId)

	buf := new(bytes.Buffer)
	err = protocol.EncodeHeader(buf, response)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not encode header")
		return err
	}

	c.SendChannel <- Response{
		Buffer:   buf,
		AddrPort: req.AddrPort,
	}

	return nil

}
