package main

import (
	"fmt"
	"ppa-control/lib/protocol"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

	"encoding/hex"

	"github.com/google/gopacket/layers"
)

var (
	// Styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7B2CBF"))

	timeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4361EE"))

	timeOffsetStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4CC9F0"))

	timezoneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	addressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	directionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F72585")).
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0, 1)

	messageTypeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#4895EF"))

	fieldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B5179E"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4CC9F0"))

	hexStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4361EE")).
			Italic(true)
)

type PacketHandler struct {
	whiteListedPackets map[protocol.MessageType]bool
	blackListedPackets map[protocol.MessageType]bool
	showAllPackets     bool

	printHexdump   bool
	captureTimeout int
	lastPacketTime time.Time
}

func NewPacketHandler(printPackets string, printHexdump bool, captureTimeout int) *PacketHandler {
	ph := &PacketHandler{
		whiteListedPackets: make(map[protocol.MessageType]bool),
		blackListedPackets: make(map[protocol.MessageType]bool),
		showAllPackets:     false,
		printHexdump:       printHexdump,
		captureTimeout:     captureTimeout,
		lastPacketTime:     time.Time{},
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

func (ph *PacketHandler) handlePacket(packet gopacket.Packet) {
	ip4Layer := packet.Layer(layers.LayerTypeIPv4)
	if ip4Layer == nil {
		return
	}
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		return
	}
	payload := udpLayer.LayerPayload()
	if payload == nil {
		return
	}
	iPv4 := ip4Layer.(*layers.IPv4)
	udp := udpLayer.(*layers.UDP)

	currentTime := packet.Metadata().Timestamp
	var timeOffset string
	if !ph.lastPacketTime.IsZero() {
		diff := currentTime.Sub(ph.lastPacketTime)
		if diff.Abs() < time.Second {
			timeOffset = fmt.Sprintf(" (+%dms)", diff.Milliseconds())
		} else {
			timeOffset = fmt.Sprintf(" (+%.3fs)", diff.Seconds())
		}
	} else {
		timeOffset = " (+0ms)"
	}
	ph.lastPacketTime = currentTime

	if udp.SrcPort == 5001 || udp.DstPort == 5001 {
		hdr, err := protocol.ParseHeader(payload)
		if err != nil {
			// Determine message direction for error case
			var direction string
			if udp.SrcPort == 5001 {
				direction = directionStyle.Render("Device → Client")
			} else {
				direction = directionStyle.Render("Client → Device")
			}

			// Format time components separately
			t := currentTime
			timeStr := fmt.Sprintf("%02d:%02d:%02d.%06d",
				t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000)

			fmt.Printf("%s\n%s\n%s %s %s\n%s\n",
				headerStyle.Render("----"),
				direction,
				timeStyle.Render(timeStr),
				timeOffsetStyle.Render(timeOffset),
				timezoneStyle.Render(t.Format("MST")),
				addressStyle.Render(fmt.Sprintf("%s:%d → %s:%d",
					iPv4.SrcIP, udp.SrcPort,
					iPv4.DstIP, udp.DstPort)))
			fmt.Printf("%s\n", hexStyle.Render(hex.Dump(payload)))
			fmt.Printf("Error: %s\n", err)
			return
		}

		if !ph.shouldProcessPacket(hdr.MessageType) {
			return
		}

		// Determine message direction
		var direction string
		if udp.SrcPort == 5001 {
			direction = directionStyle.Render("Device → Client")
		} else {
			direction = directionStyle.Render("Client → Device")
		}

		// Format time components separately
		t := currentTime
		timeStr := fmt.Sprintf("%02d:%02d:%02d.%06d",
			t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000)

		fmt.Printf("%s\n%s\n%s %s %s\n%s\n",
			headerStyle.Render("----"),
			direction,
			timeStyle.Render(timeStr),
			timeOffsetStyle.Render(timeOffset),
			timezoneStyle.Render(t.Format("MST")),
			addressStyle.Render(fmt.Sprintf("%s:%d → %s:%d",
				iPv4.SrcIP, udp.SrcPort,
				iPv4.DstIP, udp.DstPort)))

		if ph.printHexdump {
			fmt.Printf("\n%s\n%s\n",
				fieldStyle.Render("Complete payload:"),
				hexStyle.Render(hex.Dump(payload)))
		}

		fmt.Printf("%s %s (%s)\n",
			fieldStyle.Render("MessageType:"),
			messageTypeStyle.Render(hdr.MessageType.String()),
			hexStyle.Render(fmt.Sprintf("%x", byte(hdr.MessageType))))

		fmt.Printf("%s %s\n",
			fieldStyle.Render("ProtocolId:"),
			valueStyle.Render(fmt.Sprintf("%x", hdr.ProtocolId)))

		fmt.Printf("%s %s (%s)\n",
			fieldStyle.Render("Status:"),
			valueStyle.Render(hdr.Status.String()),
			hexStyle.Render(fmt.Sprintf("%x", byte(hdr.Status))))

		fmt.Printf("%s %s\n",
			fieldStyle.Render("DeviceUniqueId:"),
			valueStyle.Render(fmt.Sprintf("%x", hdr.DeviceUniqueId)))

		fmt.Printf("%s %s\n",
			fieldStyle.Render("SequenceNumber:"),
			valueStyle.Render(fmt.Sprintf("%x", hdr.SequenceNumber)))

		fmt.Printf("%s %s\n",
			fieldStyle.Render("ComponentId:"),
			valueStyle.Render(fmt.Sprintf("%x", hdr.ComponentId)))

		fmt.Printf("%s %s\n",
			fieldStyle.Render("Reserved:"),
			valueStyle.Render(fmt.Sprintf("%x", hdr.Reserved)))

		if len(payload) > 12 {
			if protocol.IsMessageTypeUnknown(hdr.MessageType) && !ph.printHexdump {
				fmt.Printf("\n%s\n%s\n",
					fieldStyle.Render("Payload:"),
					hexStyle.Render(hex.Dump(payload[12:])))
			}
		}

		// catch standard statuses
		switch hdr.Status {
		case protocol.StatusErrorServer:
			fmt.Printf("%s %s\n",
				fieldStyle.Render("Status:"),
				valueStyle.Render("ErrorServer"))
			return
		case protocol.StatusErrorClient:
			fmt.Printf("%s %s\n",
				fieldStyle.Render("Status:"),
				valueStyle.Render("ErrorClient"))
			return
		case protocol.StatusWaitClient:
			fmt.Printf("%s %s\n",
				fieldStyle.Render("Status:"),
				valueStyle.Render("WaitClient"))
			return
		case protocol.StatusWaitServer:
			fmt.Printf("%s %s\n",
				fieldStyle.Render("Status:"),
				valueStyle.Render("WaitServer"))
			return
		}

		switch hdr.MessageType {
		case protocol.MessageTypePing:
		case protocol.MessageTypeLiveCmd:
			liveCmd, err := protocol.ParseLiveCmd(payload[12:])
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}
			fmt.Printf("%s %s\n",
				fieldStyle.Render("LiveCmd.CrtFlags:"),
				valueStyle.Render(fmt.Sprintf("%x", liveCmd.CrtFlags)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("LiveCmd.OptFlags:"),
				valueStyle.Render(fmt.Sprintf("%x", liveCmd.OptFlags)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("LiveCmd.Path:"),
				valueStyle.Render(fmt.Sprintf("%x", liveCmd.Path)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("LiveCmd.Value:"),
				valueStyle.Render(fmt.Sprintf("%x", liveCmd.Value)))
		case protocol.MessageTypeDeviceData:
			if hdr.Status == protocol.StatusResponseClient {
				deviceData, err := protocol.ParseDeviceDataResponse(payload[12:])
				if err != nil {
					fmt.Printf("Error: %s\n", err)
					return
				}

				fmt.Printf("%s %s\n",
					fieldStyle.Render("DeviceData.CrtFlags:"),
					valueStyle.Render(fmt.Sprintf("%x", deviceData.CrtFlags)))
				fmt.Printf("%s %s\n",
					fieldStyle.Render("DeviceData.OptFlags:"),
					valueStyle.Render(fmt.Sprintf("%x", deviceData.OptFlags)))
				fmt.Printf("%s %s\n",
					fieldStyle.Render("DeviceData.DeviceTypeId:"),
					valueStyle.Render(fmt.Sprintf("%x", deviceData.DeviceTypeId)))
				fmt.Printf("%s %s\n",
					fieldStyle.Render("DeviceData.SubnetPrefixLength:"),
					valueStyle.Render(fmt.Sprintf("%x", deviceData.SubnetPrefixLength)))
				fmt.Printf("%s %s\n",
					fieldStyle.Render("DeviceData.DiagnosticState:"),
					valueStyle.Render(fmt.Sprintf("%x", deviceData.DiagnosticState)))
				fmt.Printf("%s %s\n",
					fieldStyle.Render("DeviceData.FirmwareVersion:"),
					valueStyle.Render(fmt.Sprintf("%x", deviceData.FirmwareVersion)))
				fmt.Printf("%s %s\n",
					fieldStyle.Render("DeviceData.SerialNumber:"),
					valueStyle.Render(fmt.Sprintf("%x", deviceData.SerialNumber)))
				fmt.Printf("%s %s\n",
					fieldStyle.Render("DeviceData.GatewayIP:"),
					valueStyle.Render(formatIpv4Address(deviceData.GatewayIP)))
				fmt.Printf("%s %s\n",
					fieldStyle.Render("DeviceData.StaticIP:"),
					valueStyle.Render(formatIpv4Address(deviceData.StaticIP)))
				fmt.Printf("%s %s\n",
					fieldStyle.Render("DeviceData.HardwareFeatures:"),
					valueStyle.Render(fmt.Sprintf("%x", deviceData.HardwareFeatures)))
				fmt.Printf("%s %s\n",
					fieldStyle.Render("DeviceData.StartPresetId:"),
					valueStyle.Render(fmt.Sprintf("%x", deviceData.StartPresetId)))
				fmt.Printf("%s '%s'\n",
					fieldStyle.Render("DeviceData.DeviceName:"),
					valueStyle.Render(string(deviceData.DeviceName[:])))
				fmt.Printf("%s %s\n",
					fieldStyle.Render("Device.VendorID:"),
					valueStyle.Render(fmt.Sprintf("%x", deviceData.VendorID)))
			} else {
				fmt.Printf("%s %s\n",
					fieldStyle.Render("Status:"),
					valueStyle.Render(hdr.Status.String()))
			}
		case protocol.MessageTypePresetRecall:
			switch hdr.Status {
			case protocol.StatusRequestServer:
				fallthrough
			case protocol.StatusResponseClient:
				fallthrough
			case protocol.StatusCommandClient:
				presetRecall, err := protocol.ParsePresetRecall(payload[12:])
				if err != nil {
					fmt.Printf("Error: %s\n", err)
					return
				}

				fmt.Printf("%s %s\n",
					fieldStyle.Render("PresetRecall.CrtFlags:"),
					valueStyle.Render(fmt.Sprintf("%x", presetRecall.CrtFlags)))
				fmt.Printf("%s %s\n",
					fieldStyle.Render("PresetRecall.OptFlags:"),
					valueStyle.Render(fmt.Sprintf("%x", presetRecall.OptFlags)))
				fmt.Printf("%s %s\n",
					fieldStyle.Render("PresetRecall.PresetId:"),
					valueStyle.Render(fmt.Sprintf("%x", presetRecall.IndexPosition)))
			}
		}
	}
}

func formatIpv4Address(ip [4]byte) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}
