package cmds

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	logger "ppa-control/lib/log"
	"ppa-control/lib/server"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts a PPA server",
	Run: func(cmd *cobra.Command, args []string) {
		address, _ := cmd.PersistentFlags().GetString("address")
		port, _ := cmd.PersistentFlags().GetUint("port")
		ctx := context.Background()
		serverString := fmt.Sprintf("%s:%d", address, port)
		fmt.Printf("Starting server on %s\n", serverString)

		grp, ctx := errgroup.WithContext(ctx)
		logger.Logger.Sugar().Infof("Starting server", "address", serverString)
		grp.Go(func() error {
			return server.RunServer(ctx, serverString)
		})

		err := grp.Wait()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.PersistentFlags().StringP("address", "a", "localhost", "Address to listen on")
	serverCmd.PersistentFlags().UintP("port", "p", 5005, "Port to listen on")
}