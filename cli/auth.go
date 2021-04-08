package cli

import (
	"fmt"
	"github.com/omecodes/common/utils/prompt"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	authCMD.AddCommand(accessCMD)
	authCMD.AddCommand(usersCMD)
}

var authCMD = &cobra.Command{
	Use:   "auth",
	Short: "Generates a new password for admin user",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func promptAuthentication() (username, password string, err error) {
	if len(authentication) != 0 {
		parts := strings.Split(strings.Trim(authentication, " "), ":")
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
	}

	username, err = prompt.Text("username", false)
	if err != nil {
		return
	}
	password, err = prompt.Password(fmt.Sprintf("%s's password", username))
	return
}
