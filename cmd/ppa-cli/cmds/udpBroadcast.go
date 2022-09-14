package cmds

import (
	"fmt"
	"github.com/spf13/cobra"
	"net"
)

var udpBroadcastCommand = &cobra.Command{
	Use:   "udp-broadcast",
	Short: "Send a simple udp broadcast command to itself",
	Run: func(cmd *cobra.Command, args []string) {
		address, _ := cmd.PersistentFlags().GetString("address")
		server, _ := cmd.PersistentFlags().GetBool("server")
		port, _ := cmd.PersistentFlags().GetUint("port")

		listenAddr := ":0"
		if server {
			listenAddr = fmt.Sprintf("%s:%d", address, port)
		}

		pc, err := net.ListenPacket("udp4", listenAddr)
		if err != nil {
			panic(err)
		}
		defer pc.Close()

		if server {
			fmt.Printf("Listening on %s\n", pc.LocalAddr().String())
			buf := make([]byte, 1024)
			n, addr, err := pc.ReadFrom(buf)
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s sent this: %s\n", addr, buf[:n])
		} else {
			dstAddr := fmt.Sprintf("%s:%d", address, port)
			addr, err := net.ResolveUDPAddr("udp4", dstAddr)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Sending to %s\n", dstAddr)
			_, err = pc.WriteTo([]byte("data to transmit"), addr)
			if err != nil {
				panic(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(udpBroadcastCommand)
	udpBroadcastCommand.PersistentFlags().StringP("address", "a", "localhost", "AddrPort to listen on")
	udpBroadcastCommand.PersistentFlags().BoolP("server", "s", false, "Run as server")
	udpBroadcastCommand.PersistentFlags().UintP("port", "p", 5005, "Port to listen on")
}
