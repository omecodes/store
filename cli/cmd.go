package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"path"
)

var (
	CMD            *cobra.Command
	authentication string

	apiLocation string
	input       string
	port        int
	output      string
	ids         []string
	noTLS       bool

	name   string
	format string
)

func init() {
	CMD = &cobra.Command{
		Use:   "scli",
		Short: "Store command line tool",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	flags := CMD.PersistentFlags()
	flags.StringVar(&authentication, "auth", "", "User authentication <user:password>")
	flags.BoolVar(&noTLS, "no-tls", false, "Uses insecure connection")
	flags.IntVar(&port, "p", 443, "Store server API port")
	flags.StringVar(&apiLocation, "api-location", "/api", "API path")

	CMD.AddCommand(authCMD)
	CMD.AddCommand(objectsCMD)
	CMD.AddCommand(filesCMD)
}

func fullAPILocation() string {
	if noTLS {
		return fmt.Sprintf("http://localhost:%d%s", port, path.Join("/", apiLocation))
	}
	return fmt.Sprintf("https://localhost:%d%s", port, path.Join("/", apiLocation))
}
