package admin

import (
	"github.com/spf13/cobra"
)

var (
	rootCMD *cobra.Command

	password string

	server    string
	input     string
	output    string
	accessIDs []string

	passwordLen int
	rounds      int
	name        string
	format      string
)

func init() {
	rootCMD = &cobra.Command{
		Use:   "admin",
		Short: "Admin command tool",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	rootCMD.AddCommand(authCMD)
	rootCMD.AddCommand(accessCMD)
}

func CMD() *cobra.Command {
	return rootCMD
}
