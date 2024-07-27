package main

import (
	"fmt"
	"ppa-control/lib/protocol"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

type PacketHandler struct {
	whiteListedPackets map[protocol.MessageType]bool
	blackListedPackets map[protocol.MessageType]bool
	showAllPackets     bool

	printHexdump   bool
	captureTimeout int
}

func NewPacketHandler(printPackets string, printHexdump bool, captureTimeout int) *PacketHandler {
	ph := &PacketHandler{
		whiteListedPackets: make(map[protocol.MessageType]bool),
		blackListedPackets: make(map[protocol.MessageType]bool),
		showAllPackets:     false,
		printHexdump:       printHexdump,
		captureTimeout:     captureTimeout,
	}
	ph.parsePacketsToPrint(printPackets)
	return ph
}

func (ph *PacketHandler) parsePacketsToPrint(printPackets string) {
	for _, p := range strings.Split(printPackets, ",") {
		p = strings.TrimSpace(p)
		if p == "all" {
			ph.showAllPackets = true
			continue
		}

		isBlacklisted := strings.HasPrefix(p, "-")
		if isBlacklisted {
			p = strings.TrimPrefix(p, "-")
		}

		var messageType protocol.MessageType
		if strings.HasPrefix(p, "type:") {
			typeByte, err := parseHexByte(strings.TrimPrefix(p, "type:"))
			if err != nil {
				continue // Skip invalid type bytes
			}
			messageType = protocol.MessageType(typeByte)
		} else {
			switch p {
			case "deviceData":
				messageType = protocol.MessageTypeDeviceData
			case "ping":
				messageType = protocol.MessageTypePing
			case "liveCmd":
				messageType = protocol.MessageTypeLiveCmd
			case "presetRecall":
				messageType = protocol.MessageTypePresetRecall
			case "unknown":
				messageType = protocol.MessageTypeUnknown
			default:
				continue // Skip invalid message types
			}
		}

		if isBlacklisted {
			ph.blackListedPackets[messageType] = true
		} else {
			ph.whiteListedPackets[messageType] = true
		}
	}
}

func (ph *PacketHandler) shouldProcessPacket(messageType protocol.MessageType) bool {
	isUnknown := protocol.IsMessageTypeUnknown(messageType)
	if ph.whiteListedPackets[messageType] || ph.blackListedPackets[messageType] {
		isUnknown = false
	}

	if ph.showAllPackets {
		// If "all" packets are shown, only reject if it's unknown and blacklisted
		if isUnknown && ph.blackListedPackets[protocol.MessageTypeUnknown] {
			return false
		}
		return !ph.blackListedPackets[messageType]
	}

	// Handle unknown message types
	if isUnknown {
		return ph.whiteListedPackets[protocol.MessageTypeUnknown] && !ph.blackListedPackets[protocol.MessageTypeUnknown]
	}

	// Handle known message types
	return ph.whiteListedPackets[messageType] && !ph.blackListedPackets[messageType]
}

func parseHexByte(s string) (byte, error) {
	val, err := strconv.ParseUint(s, 16, 8)
	if err != nil {
		return 0, err
	}
	return byte(val), nil
}

func (ph *PacketHandler) HandlePcapFile(fileName string) {
	handle, err := pcap.OpenOffline(fileName)
	if err != nil {
		panic(err)
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		ph.handlePacket(packet)
	}
}

func (ph *PacketHandler) CapturePackets(interfaceName string) {
	handle, err := pcap.OpenLive(interfaceName, 65536, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}
	defer handle.Close()

	err = handle.SetBPFFilter("udp port 5001")
	if err != nil {
		panic(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	var timeout <-chan time.Time
	if ph.captureTimeout > 0 {
		timeout = time.After(time.Duration(ph.captureTimeout) * time.Second)
	}

	fmt.Printf("Capturing packets on interface %s\n", interfaceName)

	for {
		select {
		case packet := <-packetSource.Packets():
			ph.handlePacket(packet)
		case <-timeout:
			fmt.Println("Capture timeout reached")
			return
		}
	}
}
