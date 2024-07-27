package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "pcap-dump [flags] [pcap file]",
	Short: "A tool to capture and analyze network packets",
	Long:  `pcap-dump is a CLI tool that captures and analyzes network packets using the pcap library.`,
	Run:   run,
}

func init() {
	rootCmd.Flags().String("print-packets", "deviceData,liveCmd,unknown,ping,presetRecall", "Print packets, comma-separated list of deviceData,ping,liveCmd,presetRecall,unknown")
	rootCmd.Flags().Bool("print-hexdump", false, "Print hexdump")
	rootCmd.Flags().String("interface", "", "Network interface to capture packets from")
	rootCmd.Flags().Int("timeout", 0, "Capture timeout in seconds (0 for unlimited)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	printPackets, _ := cmd.Flags().GetString("print-packets")
	printHexdump, _ := cmd.Flags().GetBool("print-hexdump")
	interfaceName, _ := cmd.Flags().GetString("interface")
	captureTimeout, _ := cmd.Flags().GetInt("timeout")

	handler := NewPacketHandler(printPackets, printHexdump, captureTimeout)

	if interfaceName != "" {
		handler.CapturePackets(interfaceName)
	} else if len(args) == 1 {
		fileName := args[0]
		fmt.Printf("Opening %s\n", fileName)
		handler.HandlePcapFile(fileName)
	} else {
		cmd.Usage()
		os.Exit(1)
	}
}
