package cmds

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"ppa-control/lib/client"
	logger "ppa-control/lib/log"
	"ppa-control/lib/utils"
	"strings"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping one or multiple PPA servers",
	Run: func(cmd *cobra.Command, args []string) {
		var discoveryAddrs []string

		addresses, _ := cmd.PersistentFlags().GetString("addresses")
		discoveryAddresses, _ := cmd.PersistentFlags().GetString("discover")
		componentId, _ := cmd.PersistentFlags().GetUint("componentId")

		port, _ := cmd.PersistentFlags().GetUint("port")
		if discoveryAddresses == "*" {
			localAddresses, err := utils.GetLocalMulticastAddresses()
			if err != nil {
				logger.Sugar.Error("error", err.Error())
				return
			}

			for _, addr := range localAddresses {
				discoveryAddrs = append(discoveryAddrs, fmt.Sprintf("%s:%d", addr.String(), port))
			}
		} else {
			for _, addr := range strings.Split(discoveryAddresses, ",") {
				discoveryAddrs = append(discoveryAddrs, fmt.Sprintf("%s:%d", addr, port))
			}
		}

		var clients []client.Client
		for _, addr := range strings.Split(addresses, ",") {
			clients = append(clients, client.NewClient(fmt.Sprintf("%s:%d", addr, port), int(componentId)))
		}
		multiClient := client.NewMultiClient(clients)
		if len(discoveryAddrs) > 0 {
			err := multiClient.Discover(discoveryAddrs)
			if err != nil {
				logger.Sugar.Error("error", err.Error())
				return
			}
		}

		ctx := context.Background()
		err := multiClient.Run(ctx)
		if err != nil {
			logger.Sugar.Error("error", err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
	pingCmd.PersistentFlags().StringP(
		"addresses", "a", "*",
		"Addresses to ping, comma separated",
	)
	pingCmd.PersistentFlags().StringP(
		"discover", "d", "",
		"Addresses to use as discovery targets, use * for all local interfaces, comma separated",
	)
	pingCmd.PersistentFlags().UintP(
		"componentId", "c", 0xFF,
		"Component ID to use for devices")

	pingCmd.PersistentFlags().UintP("port", "p", 5005, "Port to ping on")
}
