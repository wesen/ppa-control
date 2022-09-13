package cmds

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"ppa-control/lib/simulation"
)

var simulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Starts a simulated PPA device",
	Run: func(cmd *cobra.Command, args []string) {
		address, _ := cmd.PersistentFlags().GetString("address")
		port, _ := cmd.PersistentFlags().GetUint("port")
		ctx := context.Background()
		serverString := fmt.Sprintf("%s:%d", address, port)
		fmt.Printf("Starting simulated PPA device on %s\n", serverString)

		grp, ctx := errgroup.WithContext(ctx)

		client := simulation.NewClient(serverString, "SimulatedClient", [4]byte{0, 1, 2, 3}, 0xFF)
		grp.Go(func() error {
			return client.Run(ctx)
		})

		err := grp.Wait()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(simulateCmd)
	simulateCmd.PersistentFlags().StringP("address", "a", "localhost", "AddrPort to listen on")
	simulateCmd.PersistentFlags().UintP("port", "p", 5005, "Port to listen on")
}
