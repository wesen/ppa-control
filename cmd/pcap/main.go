package main

import (
	"fmt"
	"github.com/google/gopacket"
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
	fmt.Printf("%s\n", packet.String())
}
