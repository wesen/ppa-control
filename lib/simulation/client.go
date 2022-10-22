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
	"ppa-control/lib/utils"
	"time"
)

const MaxBufferSize = 1024
const Timeout = 10 * time.Second

type Preset struct {
}

type Response struct {
	Buffer *bytes.Buffer
	Addr   net.Addr
}

type Request struct {
	Buffer *bytes.Buffer
	Addr   net.Addr
}

type SimulatedDeviceSettings struct {
	UniqueId    [4]byte
	ComponentId byte
	Name        string
	Address     string
	Port        uint16
	// if not empty, bind to the given interface
	Interface string
}

type SimulatedDevice struct {
	SendChannel           chan Response
	ReceiveChannel        chan *bytes.Buffer
	Settings              SimulatedDeviceSettings
	presets               []Preset
	currentlyActivePreset int
	currentVolume         float32
}

func NewSimulatedDevice(settings SimulatedDeviceSettings) *SimulatedDevice {
	return &SimulatedDevice{
		SendChannel:           make(chan Response),
		ReceiveChannel:        make(chan *bytes.Buffer),
		Settings:              settings,
		currentlyActivePreset: 0,
		currentVolume:         0.0,
	}
}

func (sd *SimulatedDevice) Run(ctx context.Context) (err error) {
	serverString := fmt.Sprintf("%s:%d", sd.Settings.Address, sd.Settings.Port)
	conn, err := utils.ListenUDP(ctx, serverString, sd.Settings.Interface)

	if err != nil {
		log.Error().
			Err(err).
			Str("interface", sd.Settings.Interface).
			Msg("Could not bind to interface")
		return err
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

			// TODO add timeout for simulated device

			//deadline := time.Now().Add(Timeout)
			//err = conn.SetReadDeadline(deadline)
			//if err != nil {
			//	log.Warn().
			//		Err(err).
			//		Msg("Could not set read deadline")
			//}

			n, srcAddr, err := conn.ReadFrom(buffer)
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
				Buffer: bytes.NewBuffer(buffer[:n]),
				Addr:   srcAddr,
			}

			if n > 0 {

				switch protocol.MessageType(buffer[0]) {
				case protocol.MessageTypePing:
					{
						err := sd.handlePing(request)
						if err != nil {
							log.Warn().Str("error", err.Error()).Msg("Could not handle ping")
						}
					}
				case protocol.MessageTypeLiveCmd:
					{
						err := sd.handleLiveCmd(request)
						if err != nil {
							log.Warn().Str("error", err.Error()).Msg("Could not handle live command")
						}
					}
				case protocol.MessageTypeDeviceData:
					{
						err := sd.handleDeviceData(request)
						if err != nil {
							log.Warn().Str("error", err.Error()).Msg("Could not handle device data")
						}
					}
				case protocol.MessageTypePresetRecall:
					{
						err := sd.handlePresetRecall(request)
						if err != nil {
							log.Warn().Str("error", err.Error()).Msg("Could not handle preset recall")
						}
					}
				case protocol.MessageTypePresetSave:
					{
						err := sd.handlePresetSave(request)
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

			case response := <-sd.SendChannel:
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
				n, err := conn.WriteTo(response.Buffer.Bytes(), response.Addr)
				if err != nil {
					log.Warn().
						Err(err).
						Str("to", response.Addr.String()).
						Str("local", conn.LocalAddr().String()).
						Int("length", response.Buffer.Len()).
						Bytes("data", response.Buffer.Bytes()).
						Msg("Could not write to UDP")
				} else {
					log.Info().
						Str("to", response.Addr.String()).
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

func (sd *SimulatedDevice) handlePing(req *Request) error {
	hdr, err := protocol.ParseHeader(req.Buffer.Bytes())
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not parse ping header")
		return err
	}

	response := protocol.NewBasicHeader(
		protocol.MessageTypePing,
		protocol.StatusResponseServer,
		sd.Settings.UniqueId,
		hdr.SequenceNumber,
		sd.Settings.ComponentId)

	buf := new(bytes.Buffer)
	err = protocol.EncodeHeader(buf, response)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not encode header")
		return err
	}

	sd.SendChannel <- Response{
		Buffer: buf,
		Addr:   req.Addr,
	}

	return nil
}

func (sd *SimulatedDevice) handleLiveCmd(req *Request) error {
	hdr, err := protocol.ParseHeader(req.Buffer.Bytes())
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not parse ping header")
		return err
	}

	response := protocol.NewBasicHeader(
		protocol.MessageTypePing,
		protocol.StatusResponseServer,
		sd.Settings.UniqueId,
		hdr.SequenceNumber,
		sd.Settings.ComponentId)

	buf := new(bytes.Buffer)
	err = protocol.EncodeHeader(buf, response)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not encode header")
		return err
	}

	sd.SendChannel <- Response{
		Buffer: buf,
		Addr:   req.Addr,
	}

	return nil
}

func (sd *SimulatedDevice) handleDeviceData(req *Request) error {
	hdr, err := protocol.ParseHeader(req.Buffer.Bytes())
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not parse ping header")
		return err
	}

	response := protocol.NewBasicHeader(
		protocol.MessageTypePing,
		protocol.StatusResponseServer,
		sd.Settings.UniqueId,
		hdr.SequenceNumber,
		sd.Settings.ComponentId)

	buf := new(bytes.Buffer)
	err = protocol.EncodeHeader(buf, response)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not encode header")
		return err
	}

	sd.SendChannel <- Response{
		Buffer: buf,
		Addr:   req.Addr,
	}

	return nil

}

func (sd *SimulatedDevice) handlePresetRecall(req *Request) error {
	hdr, err := protocol.ParseHeader(req.Buffer.Bytes())
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not parse ping header")
		return err
	}

	response := protocol.NewBasicHeader(
		protocol.MessageTypePing,
		protocol.StatusResponseServer,
		sd.Settings.UniqueId,
		hdr.SequenceNumber,
		sd.Settings.ComponentId)

	buf := new(bytes.Buffer)
	err = protocol.EncodeHeader(buf, response)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not encode header")
		return err
	}

	sd.SendChannel <- Response{
		Buffer: buf,
		Addr:   req.Addr,
	}

	return nil
}

func (sd *SimulatedDevice) handlePresetSave(req *Request) error {
	hdr, err := protocol.ParseHeader(req.Buffer.Bytes())
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not parse ping header")
		return err
	}

	response := protocol.NewBasicHeader(
		protocol.MessageTypePing,
		protocol.StatusResponseServer,
		sd.Settings.UniqueId,
		hdr.SequenceNumber,
		sd.Settings.ComponentId)

	buf := new(bytes.Buffer)
	err = protocol.EncodeHeader(buf, response)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Could not encode header")
		return err
	}

	sd.SendChannel <- Response{
		Buffer: buf,
		Addr:   req.Addr,
	}

	return nil

}
