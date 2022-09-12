package protocol

//go:generate stringer -type=MessageType
//go:generate stringer -type=StatusType

import (
	"bytes"
	"encoding/binary"
	"io"
)

type MessageType byte

const (
	MessageTypePing         MessageType = 0
	MessageTypeLiveCmd      MessageType = 1
	MessageTypeDeviceData   MessageType = 2
	MessageTypePresetRecall MessageType = 4
)

type StatusType uint16

const (
	StatusCommandClient  StatusType = 0x0102
	StatusRequestClient  StatusType = 0x0106
	StatusResponseClient StatusType = 0x0101
	StatusErrorClient    StatusType = 0x0109
	StatusWaitClient     StatusType = 0x0141
	StatusCommandServer  StatusType = 0x0002
	StatusRequestServer  StatusType = 0x0006
	StatusResponseServer StatusType = 0x0001
	StatusErrorServer    StatusType = 0x0009
	StatusWaitServer     StatusType = 0x0041
)

type BasicHeader struct {
	MessageType    MessageType
	ProtocolId     byte // always 1
	Status         StatusType
	DeviceUniqueId [4]byte
	SequenceNumber uint16
	ComponentId    byte
	Reserved       byte // leave 0
}

func ParseHeader(buf []byte) (*BasicHeader, error) {
	w := bytes.NewReader(buf)
	h := &BasicHeader{}
	err := binary.Read(w, binary.LittleEndian, &h.MessageType)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &h.ProtocolId)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &h.Status)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &h.DeviceUniqueId)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &h.SequenceNumber)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &h.ComponentId)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &h.Reserved)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func NewBasicHeader(
	messageType MessageType,
	status StatusType,
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
	// XXX traditional way to do it
	// binary.LittleEndian.PutUint16(w, p.CrtFlags)
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

func ParsePresetRecall(buf []byte) (*PresetRecall, error) {
	w := bytes.NewReader(buf)
	pr := &PresetRecall{}
	err := binary.Read(w, binary.LittleEndian, &pr.CrtFlags)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &pr.OptFlags)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &pr.IndexPosition)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &pr.Reserved)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

type LevelType byte

const (
	LevelTypeInput          LevelType = 1
	LevelTypeOutput         LevelType = 2
	LevelTypeEq             LevelType = 3
	LevelTypeGain           LevelType = 4
	LevelTypeEqType         LevelType = 5
	LevelTypeQuality        LevelType = 7
	LevelTypeActive         LevelType = 8
	LevelTypeMute           LevelType = 9
	LevelTypeDelay          LevelType = 10
	LevelTypePhaseInversion LevelType = 11
)

type DeviceDataRequest struct {
	CrtFlags uint8
	OptFlags uint8
}

func ParseDeviceDataRequest(buf []byte) (*DeviceDataRequest, error) {
	w := bytes.NewReader(buf)
	d := &DeviceDataRequest{}
	err := binary.Read(w, binary.LittleEndian, &d.CrtFlags)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.OptFlags)
	if err != nil {
		return nil, err
	}

	return d, nil
}

type DeviceDataResponse struct {
	CrtFlags           uint8
	OptFlags           uint8
	DeviceTypeId       uint16
	SubnetPrefixLength uint8
	DiagnosticState    uint8
	FirmwareVersion    uint32
	SerialNumber       uint16
	Reserved           uint32
	GatewayIP          [4]byte
	StaticIP           [4]byte
	HardwareFeatures   uint32
	StartPresetId      uint8
	Reserved2          [6]byte
	DeviceName         [32]byte
	VendorID           uint8
}

func ParseDeviceDataResponse(buf []byte) (*DeviceDataResponse, error) {
	w := bytes.NewReader(buf)
	d := &DeviceDataResponse{}
	err := binary.Read(w, binary.LittleEndian, &d.CrtFlags)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.OptFlags)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.DeviceTypeId)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.SubnetPrefixLength)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.DiagnosticState)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.FirmwareVersion)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.SerialNumber)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.Reserved)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.GatewayIP)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.StaticIP)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.HardwareFeatures)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.StartPresetId)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.Reserved2)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.DeviceName)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &d.VendorID)
	if err != nil {
		return nil, err
	}

	return d, nil
}

type LiveCmd struct {
	CrtFlags uint8
	OptFlags uint8
	Path     [10]byte
	Value    uint32
}

func ParseLiveCmd(buf []byte) (*LiveCmd, error) {
	w := bytes.NewReader(buf)
	lc := &LiveCmd{}
	err := binary.Read(w, binary.LittleEndian, &lc.CrtFlags)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &lc.OptFlags)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &lc.Path)
	if err != nil {
		return nil, err
	}
	err = binary.Read(w, binary.LittleEndian, &lc.Value)
	if err != nil {
		return nil, err
	}

	return lc, nil
}
