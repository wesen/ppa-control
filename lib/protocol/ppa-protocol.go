package protocol

import (
	"encoding/binary"
	"io"
)

const (
	MessageTypePing         byte = 0
	MessageTypeLiveCmd           = 1
	MessageTypeDeviceData        = 2
	MessageTypePresetRecall      = 4
)

const (
	StatusCommand  uint16 = 0x0102
	StatusRequest         = 0x0106
	StatusResponse        = 0x0101
	StatusError           = 0x0109
	StatusWait            = 0x0141
)

type BasicHeader struct {
	MessageType    byte
	ProtocolId     byte // always 1
	Status         uint16
	DeviceUniqueId [4]byte
	SequenceNumber uint16
	ComponentId    byte
	Reserved       byte // leave 0
}

func NewBasicHeader(
	messageType byte,
	status uint16,
	deviceUniqueId [4]byte,
	sequenceNumber uint16,
	componentId byte) *BasicHeader {
	return &BasicHeader{
		MessageType:    messageType,
		ProtocolId:     1,
		Status:         status,
		DeviceUniqueId: deviceUniqueId,
		SequenceNumber: sequenceNumber,
		ComponentId:    componentId,
		Reserved:       1, // Change
	}
}

func EncodeHeader(w io.Writer, h *BasicHeader) error {
	err := binary.Write(w, binary.LittleEndian, h.MessageType)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, h.ProtocolId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, h.Status)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, h.DeviceUniqueId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, h.SequenceNumber)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, h.ComponentId)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, h.Reserved)
	if err != nil {
		return err
	}
	return nil
}

const (
	RecallByPresetIndex    uint8 = 0
	RecallByPresetPosition       = 2
)

type PresetRecall struct {
	CrtFlags      uint8
	OptFlags      uint8
	IndexPosition uint8
	Reserved      uint8 // leave 0
}

func EncodePresetRecall(w io.Writer, pr *PresetRecall) error {
	err := binary.Write(w, binary.LittleEndian, pr.CrtFlags)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, pr.OptFlags)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, pr.IndexPosition)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, pr.Reserved)
	if err != nil {
		return err
	}

	return nil
}

func NewPresetRecall(crtFlags uint8, optFlags uint8, indexPosition uint8) *PresetRecall {
	return &PresetRecall{
		CrtFlags:      crtFlags,
		OptFlags:      optFlags,
		IndexPosition: indexPosition,
		Reserved:      0,
	}
}
