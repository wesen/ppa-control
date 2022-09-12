package cmds

import "github.com/spf13/cobra"

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Set the volume of one or more clients",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(volumeCmd)
}