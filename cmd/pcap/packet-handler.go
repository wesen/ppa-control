package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"ppa-control/lib/protocol"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"gopkg.in/yaml.v3"
)

// PacketData represents a structured packet for JSON/YAML output
type PacketData struct {
	Timestamp   string                `json:"timestamp" yaml:"timestamp"`
	TimeOffset  string                `json:"time_offset" yaml:"time_offset"`
	Direction   string                `json:"direction" yaml:"direction"`
	Source      string                `json:"source" yaml:"source"`
	Destination string                `json:"destination" yaml:"destination"`
	Header      *protocol.BasicHeader `json:"header" yaml:"header"`
	Payload     interface{}           `json:"payload,omitempty" yaml:"payload,omitempty"`
	HexDump     string                `json:"hex_dump,omitempty" yaml:"hex_dump,omitempty"`
}

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
	printHexdump       bool
	captureTimeout     int
	lastPacketTime     time.Time
	outputFormat       string
	jsonPackets        []PacketData // For collecting JSON array output
}

func NewPacketHandler(printPackets string, excludePackets string, printHexdump bool, captureTimeout int, outputFormat string) *PacketHandler {
	ph := &PacketHandler{
		whiteListedPackets: make(map[protocol.MessageType]bool),
		blackListedPackets: make(map[protocol.MessageType]bool),
		showAllPackets:     false,
		printHexdump:       printHexdump,
		captureTimeout:     captureTimeout,
		lastPacketTime:     time.Time{},
		outputFormat:       outputFormat,
		jsonPackets:        make([]PacketData, 0),
	}
	ph.parsePacketsToPrint(printPackets)
	ph.parsePacketsToExclude(excludePackets)
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
		} else if val, err := strconv.ParseUint(p, 10, 8); err == nil {
			// Direct numeric message type
			messageType = protocol.MessageType(val)
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

func (ph *PacketHandler) parsePacketsToExclude(excludePackets string) {
	if excludePackets == "" {
		return
	}

	for _, p := range strings.Split(excludePackets, ",") {
		p = strings.TrimSpace(p)

		var messageType protocol.MessageType
		if strings.HasPrefix(p, "type:") {
			typeByte, err := parseHexByte(strings.TrimPrefix(p, "type:"))
			if err != nil {
				continue // Skip invalid type bytes
			}
			messageType = protocol.MessageType(typeByte)
		} else if val, err := strconv.ParseUint(p, 10, 8); err == nil {
			// Direct numeric message type
			messageType = protocol.MessageType(val)
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
		ph.blackListedPackets[messageType] = true
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

	// Output JSON array at the end if using json format
	if ph.outputFormat == "json" && len(ph.jsonPackets) > 0 {
		if jsonData, err := json.MarshalIndent(ph.jsonPackets, "", "  "); err == nil {
			fmt.Println(string(jsonData))
		}
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
			// Output JSON array at the end if using json format
			if ph.outputFormat == "json" && len(ph.jsonPackets) > 0 {
				if jsonData, err := json.MarshalIndent(ph.jsonPackets, "", "  "); err == nil {
					fmt.Println(string(jsonData))
				}
			}
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
			timeOffset = fmt.Sprintf("+%dms", diff.Milliseconds())
		} else {
			timeOffset = fmt.Sprintf("+%.3fs", diff.Seconds())
		}
	} else {
		timeOffset = "+0ms"
	}
	ph.lastPacketTime = currentTime

	if udp.SrcPort == 5001 || udp.DstPort == 5001 {
		hdr, err := protocol.ParseHeader(payload)
		if err != nil {
			// Handle error case
			packetData := PacketData{
				Timestamp:   currentTime.Format("15:04:05.000000"),
				TimeOffset:  timeOffset,
				Direction:   ph.getDirection(udp.SrcPort == 5001),
				Source:      fmt.Sprintf("%s:%d", iPv4.SrcIP, udp.SrcPort),
				Destination: fmt.Sprintf("%s:%d", iPv4.DstIP, udp.DstPort),
				HexDump:     hex.Dump(payload),
			}

			ph.outputPacket(packetData)
			return
		}

		if !ph.shouldProcessPacket(hdr.MessageType) {
			return
		}

		packetData := PacketData{
			Timestamp:   currentTime.Format("15:04:05.000000"),
			TimeOffset:  timeOffset,
			Direction:   ph.getDirection(udp.SrcPort == 5001),
			Source:      fmt.Sprintf("%s:%d", iPv4.SrcIP, udp.SrcPort),
			Destination: fmt.Sprintf("%s:%d", iPv4.DstIP, udp.DstPort),
			Header:      hdr,
		}

		if ph.printHexdump {
			packetData.HexDump = hex.Dump(payload)
		}

		// Parse payload based on message type
		if len(payload) > 12 {
			switch hdr.MessageType {
			case protocol.MessageTypeLiveCmd:
				if cmd, err := protocol.ParseLiveCmd(payload[12:]); err == nil {
					packetData.Payload = cmd
				}
			case protocol.MessageTypeDeviceData:
				if hdr.Status == protocol.StatusResponseClient {
					if data, err := protocol.ParseDeviceDataResponse(payload[12:]); err == nil {
						packetData.Payload = data
					}
				}
			case protocol.MessageTypePresetRecall:
				if recall, err := protocol.ParsePresetRecall(payload[12:]); err == nil {
					packetData.Payload = recall
				}
			}
		}

		ph.outputPacket(packetData)
	}
}

func (ph *PacketHandler) getDirection(fromDevice bool) string {
	if fromDevice {
		return "Device → Client"
	}
	return "Client → Device"
}

func (ph *PacketHandler) outputPacket(data PacketData) {
	switch ph.outputFormat {
	case "json":
		ph.jsonPackets = append(ph.jsonPackets, data)
	case "jsonl":
		if jsonData, err := json.Marshal(data); err == nil {
			fmt.Println(string(jsonData))
		}
	case "yaml":
		if yamlData, err := yaml.Marshal(data); err == nil {
			fmt.Printf("---\n%s", string(yamlData))
		}
	default:
		ph.outputTextFormat(data)
	}
}

func (ph *PacketHandler) outputTextFormat(data PacketData) {
	fmt.Printf("%s\n%s\n%s %s %s\n%s\n",
		headerStyle.Render("----"),
		directionStyle.Render(data.Direction),
		timeStyle.Render(data.Timestamp),
		timeOffsetStyle.Render(data.TimeOffset),
		timezoneStyle.Render(time.Now().Format("MST")),
		addressStyle.Render(fmt.Sprintf("%s → %s", data.Source, data.Destination)))

	if data.HexDump != "" {
		fmt.Printf("\n%s\n%s\n",
			fieldStyle.Render("Complete payload:"),
			hexStyle.Render(data.HexDump))
	}

	if data.Header != nil {
		fmt.Printf("%s %s (%s)\n",
			fieldStyle.Render("MessageType:"),
			messageTypeStyle.Render(data.Header.MessageType.String()),
			hexStyle.Render(fmt.Sprintf("%x", byte(data.Header.MessageType))))

		fmt.Printf("%s %s\n",
			fieldStyle.Render("ProtocolId:"),
			valueStyle.Render(fmt.Sprintf("%x", data.Header.ProtocolId)))

		fmt.Printf("%s %s (%s)\n",
			fieldStyle.Render("Status:"),
			valueStyle.Render(data.Header.Status.String()),
			hexStyle.Render(fmt.Sprintf("%x", byte(data.Header.Status))))

		fmt.Printf("%s %s\n",
			fieldStyle.Render("DeviceUniqueId:"),
			valueStyle.Render(fmt.Sprintf("%x", data.Header.DeviceUniqueId)))

		fmt.Printf("%s %s\n",
			fieldStyle.Render("SequenceNumber:"),
			valueStyle.Render(fmt.Sprintf("%x", data.Header.SequenceNumber)))

		fmt.Printf("%s %s\n",
			fieldStyle.Render("ComponentId:"),
			valueStyle.Render(fmt.Sprintf("%x", data.Header.ComponentId)))

		fmt.Printf("%s %s\n",
			fieldStyle.Render("Reserved:"),
			valueStyle.Render(fmt.Sprintf("%x", data.Header.Reserved)))
	}

	if data.Payload != nil {
		switch p := data.Payload.(type) {
		case *protocol.LiveCmd:
			fmt.Printf("%s %s\n",
				fieldStyle.Render("LiveCmd.CrtFlags:"),
				valueStyle.Render(fmt.Sprintf("%x", p.CrtFlags)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("LiveCmd.OptFlags:"),
				valueStyle.Render(fmt.Sprintf("%x", p.OptFlags)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("LiveCmd.Path:"),
				valueStyle.Render(fmt.Sprintf("%x", p.Path)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("LiveCmd.Value:"),
				valueStyle.Render(fmt.Sprintf("%x", p.Value)))
		case *protocol.DeviceDataResponse:
			fmt.Printf("%s %s\n",
				fieldStyle.Render("DeviceData.CrtFlags:"),
				valueStyle.Render(fmt.Sprintf("%x", p.CrtFlags)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("DeviceData.OptFlags:"),
				valueStyle.Render(fmt.Sprintf("%x", p.OptFlags)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("DeviceData.DeviceTypeId:"),
				valueStyle.Render(fmt.Sprintf("%x", p.DeviceTypeId)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("DeviceData.SubnetPrefixLength:"),
				valueStyle.Render(fmt.Sprintf("%x", p.SubnetPrefixLength)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("DeviceData.DiagnosticState:"),
				valueStyle.Render(fmt.Sprintf("%x", p.DiagnosticState)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("DeviceData.FirmwareVersion:"),
				valueStyle.Render(fmt.Sprintf("%x", p.FirmwareVersion)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("DeviceData.SerialNumber:"),
				valueStyle.Render(fmt.Sprintf("%x", p.SerialNumber)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("DeviceData.GatewayIP:"),
				valueStyle.Render(formatIpv4Address(p.GatewayIP)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("DeviceData.StaticIP:"),
				valueStyle.Render(formatIpv4Address(p.StaticIP)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("DeviceData.HardwareFeatures:"),
				valueStyle.Render(fmt.Sprintf("%x", p.HardwareFeatures)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("DeviceData.StartPresetId:"),
				valueStyle.Render(fmt.Sprintf("%x", p.StartPresetId)))
			fmt.Printf("%s '%s'\n",
				fieldStyle.Render("DeviceData.DeviceName:"),
				valueStyle.Render(string(p.DeviceName[:])))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("Device.VendorID:"),
				valueStyle.Render(fmt.Sprintf("%x", p.VendorID)))
		case *protocol.PresetRecall:
			fmt.Printf("%s %s\n",
				fieldStyle.Render("PresetRecall.CrtFlags:"),
				valueStyle.Render(fmt.Sprintf("%x", p.CrtFlags)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("PresetRecall.OptFlags:"),
				valueStyle.Render(fmt.Sprintf("%x", p.OptFlags)))
			fmt.Printf("%s %s\n",
				fieldStyle.Render("PresetRecall.PresetId:"),
				valueStyle.Render(fmt.Sprintf("%x", p.IndexPosition)))
		}
	}
}

func formatIpv4Address(ip [4]byte) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}
