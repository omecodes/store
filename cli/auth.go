package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	authCMD.AddCommand(accessCMD)
	authCMD.AddCommand(credentialsCMD)
}

var authCMD = &cobra.Command{
	Use:   "auth",
	Short: "Manages authentication",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}
