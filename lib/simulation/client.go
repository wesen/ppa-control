package simulation

import (
	"bytes"
	"context"
	"fmt"
	"github.com/augustoroman/hexdump"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net"
	"ppa-control/lib/protocol"
	"time"
)

const MaxBufferSize = 1024
const Timeout = 10 * time.Second

type Preset struct {
}

type client struct {
	SendChannel           chan *bytes.Buffer
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
		SendChannel:           make(chan *bytes.Buffer),
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

	deadline := time.Now().Add(Timeout)
	err = conn.SetWriteDeadline(deadline)
	if err != nil {
		return err
	}

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		defer func() {
			log.Info().Msg("read-loop exiting\n")
		}()

		for {
			buffer := make([]byte, MaxBufferSize)

			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.Error().Str("error", err.Error()).Msg("Could not read from UDP")
				return err
			}

			log.Info().Int("received", n).
				Str("from", addr.String()).
				Msg("Received packet")
			fmt.Printf("%s\n", hexdump.Dump(buffer[:n]))

			if n > 0 {

				switch protocol.MessageType(buffer[0]) {
				case protocol.MessageTypePing:
					{
						err := c.handlePing(buffer)
						if err != nil {
							log.Warn().Str("error", err.Error()).Msg("Could not handle ping")
						}
					}
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
				log.Info().Msg("write-loop exiting")
				return ctx.Err()

			case buf := <-c.SendChannel:
				// Send
				n, err := conn.Write(buf.Bytes())
				if err != nil {
					return err
				}
				log.Info().Int("bytes", n).Msg("Wrote packet")
			}
		}
	})

	err = grp.Wait()
	log.Info().Msg("Exiting client loop")
	return err
}

func (c *client) handlePing(buffer []byte) error {
	hdr, err := protocol.ParseHeader(buffer)
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

	c.SendChannel <- buf

	return nil
}
