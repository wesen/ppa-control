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

type client struct {
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

func NewClient(address string, name string, deviceUniqueId [4]byte, componentId byte) *client {
	return &client{
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

func (c *client) Run(ctx context.Context) (err error) {
	addr, err := net.ResolveUDPAddr("udp", c.address)
	if err != nil {
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return
	}
	defer conn.Close()

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
	log.Info().Msg("Exiting client loop")
	return err
}

func (c *client) handlePing(req *Request) error {
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

func (c *client) handleLiveCmd(req *Request) error {
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

func (c *client) handleDeviceData(req *Request) error {
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

func (c *client) handlePresetRecall(req *Request) error {
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

func (c *client) handlePresetSave(req *Request) error {
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
