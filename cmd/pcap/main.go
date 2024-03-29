package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"os"
	"ppa-control/lib/protocol"
	"strings"
)

var (
	printPackets = flag.String("print-packets", "deviceData,liveCmd,unknown,ping,presetRecall", "Print packets, comma-separated list of deviceData,ping,liveCmd,presetRecall,unknown")
	printHexdump = flag.Bool("print-hexdump", false, "Print hexdump")
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Usage: pcap-dump [-print-packets (deviceData,liveCmd,etc...)] <filename>")
		os.Exit(1)
	}

	packetsToPrints := map[protocol.MessageType]bool{}
	for _, p := range strings.Split(*printPackets, ",") {
		switch p {
		case "deviceData":
			packetsToPrints[protocol.MessageTypeDeviceData] = true
		case "ping":
			packetsToPrints[protocol.MessageTypePing] = true
		case "liveCmd":
			packetsToPrints[protocol.MessageTypeLiveCmd] = true
		case "presetRecall":
			packetsToPrints[protocol.MessageTypePresetRecall] = true
		case "unknown":
			packetsToPrints[protocol.MessageTypeUnknown] = true
		}
	}

	fileName := args[0]
	fmt.Printf("Opening %s\n", fileName)

	if handle, err := pcap.OpenOffline(fileName); err != nil {
		panic(err)
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			handlePacket(packet, packetsToPrints) // Do something with a packet here.
		}
	}
}

func handlePacket(packet gopacket.Packet, packetsToPrint map[protocol.MessageType]bool) {
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

	if udp.SrcPort == 5001 || udp.DstPort == 5001 {
		hdr, err := protocol.ParseHeader(payload)
		if err != nil {
			fmt.Printf("----\nsrc: %s:%d dst: %s:%d\n",
				iPv4.SrcIP,
				udp.SrcPort,
				iPv4.DstIP,
				udp.DstPort)
			fmt.Printf("%s--\n", hex.Dump(payload))
			fmt.Printf("Error: %s\n", err)
			return
		}

		isUnknown := false

		switch hdr.MessageType {
		case protocol.MessageTypeDeviceData:
		case protocol.MessageTypePing:
		case protocol.MessageTypeLiveCmd:
		case protocol.MessageTypePresetRecall:
		default:
			isUnknown = true
		}

		// skip unknown message types
		if !packetsToPrint[hdr.MessageType] && !isUnknown {
			return
		}
		if isUnknown && !packetsToPrint[protocol.MessageTypeUnknown] {
			return
		}

		fmt.Printf("----\nsrc: %s:%d dst: %s:%d\n",
			iPv4.SrcIP,
			udp.SrcPort,
			iPv4.DstIP,
			udp.DstPort)
		if *printHexdump {
			fmt.Printf("%s--\n", hex.Dump(payload))
		}

		// the format method checks to see if there is a string method, and thus if
		fmt.Printf("MessageType: %s (%x)\n", hdr.MessageType, byte(hdr.MessageType))
		fmt.Printf("ProtocolId: %x\n", hdr.ProtocolId)
		fmt.Printf("Status: %s (%x)\n", hdr.Status, byte(hdr.Status))
		fmt.Printf("DeviceUniqueId: %x\n", hdr.DeviceUniqueId)
		fmt.Printf("SequenceNumber: %x\n", hdr.SequenceNumber)
		fmt.Printf("ComponentId: %x\n", hdr.ComponentId)
		fmt.Printf("Reserved: %x\n", hdr.Reserved)

		if len(payload) > 12 {
			if isUnknown || *printHexdump {
				fmt.Printf("\nPayload: %s\n", hex.Dump(payload[12:]))
			}
		}

		// catch standard statuses
		switch hdr.Status {
		case protocol.StatusErrorServer:
			fmt.Printf("Status: ErrorServer\n")
			return
		case protocol.StatusErrorClient:
			fmt.Printf("Status: ErrorClient\n")
			return
		case protocol.StatusWaitClient:
			fmt.Printf("Status: WaitClient\n")
			return
		case protocol.StatusWaitServer:
			fmt.Printf("Status: WaitServer\n")
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
			fmt.Printf("LiveCmd.CrtFlags: %x\n", liveCmd.CrtFlags)
			fmt.Printf("LiveCmd.OptFlags: %x\n", liveCmd.OptFlags)
			fmt.Printf("LiveCmd.Path: %x\n", liveCmd.Path)
			fmt.Printf("LiveCmd.Path: %x\n", liveCmd.Value)
		case protocol.MessageTypeDeviceData:
			if hdr.Status == protocol.StatusResponseClient {
				deviceData, err := protocol.ParseDeviceDataResponse(payload[12:])
				if err != nil {
					fmt.Printf("Error: %s\n", err)
					return
				}

				fmt.Printf("DeviceData.CrtFlags: %x\n", deviceData.CrtFlags)
				fmt.Printf("DeviceData.OptFlags: %x\n", deviceData.OptFlags)
				fmt.Printf("DeviceData.DeviceTypeId: %x\n", deviceData.DeviceTypeId)
				fmt.Printf("DeviceData.SubnetPrefixLength: %x\n", deviceData.SubnetPrefixLength)
				fmt.Printf("DeviceData.DiagnosticState: %x\n", deviceData.DiagnosticState)
				fmt.Printf("DeviceData.FirmwareVersion: %x\n", deviceData.FirmwareVersion)
				fmt.Printf("DeviceData.SerialNumber: %x\n", deviceData.SerialNumber)
				fmt.Printf("DeviceData.GatewayIP: %s\n", formatIpv4Address(deviceData.GatewayIP))
				fmt.Printf("DeviceData.StaticIP: %s\n", formatIpv4Address(deviceData.StaticIP))
				fmt.Printf("DeviceData.HardwareFeatures: %x\n", deviceData.HardwareFeatures)
				fmt.Printf("DeviceData.StartPresetId: %x\n", deviceData.StartPresetId)
				fmt.Printf("DeviceData.DeviceName: '%s'\n", deviceData.DeviceName)
				fmt.Printf("Device.VendorID: %x\n", deviceData.VendorID)

			} else {
				fmt.Printf("Status: %s\n", hdr.Status)
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

				fmt.Printf("PresetRecall.CrtFlags: %x\n", presetRecall.CrtFlags)
				fmt.Printf("PresetRecall.OptFlags: %x\n", presetRecall.OptFlags)
				fmt.Printf("PresetRecall.PresetId: %x\n", presetRecall.IndexPosition)
			}
		}
	}
}

func formatIpv4Address(ipv4Bytes [4]byte) string {
	return fmt.Sprintf("%d.%d.%d.%d", ipv4Bytes[3], ipv4Bytes[2], ipv4Bytes[1], ipv4Bytes[0])
}
