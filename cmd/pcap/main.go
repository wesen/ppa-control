package main

import (
	"encoding/hex"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: pcap-dump <filename>")
		os.Exit(1)
	}

	fileName := os.Args[1]
	fmt.Printf("Opening %s\n", fileName)

	if handle, err := pcap.OpenOffline(fileName); err != nil {
		panic(err)
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			handlePacket(packet) // Do something with a packet here.
		}
	}
}

func handlePacket(packet gopacket.Packet) {
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
		fmt.Printf("src: %s:%d dst: %s:%d\n",
			iPv4.SrcIP,
			udp.SrcPort,
			iPv4.DstIP,
			udp.DstPort)
		fmt.Printf("%s\n", hex.Dump(payload))
	}
}
