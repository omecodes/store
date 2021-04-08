package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	objectsCMD.AddCommand(objectsCollectionsCMD)
}

var objectsCMD = &cobra.Command{
	Use:   "objects",
	Short: "Manage objects store",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}
