package cmds

import "github.com/spf13/cobra"

var recallCmd = &cobra.Command{
	Use:   "recall",
	Short: "Recall a preset by index",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(recallCmd)
}
