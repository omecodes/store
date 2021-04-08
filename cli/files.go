package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	filesCMD.AddCommand(fileSourcesCMD)
}

var filesCMD = &cobra.Command{
	Use:   "files",
	Short: "Manage files",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}
