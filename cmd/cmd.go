package cmd

import (
	"fmt"
	"github.com/omecodes/common/env/app"
	"github.com/spf13/cobra"
)

var (
	addr           string
	certFilename   string
	keyFilename    string
	selfSignedTLS  bool
	dsn            string
	application    *app.App
	options        []app.Option
	command        *cobra.Command
	defaultOptions = []app.Option{
		app.WithRunCommandFunc(runStore),
		app.WithVersion("1.0.1"),
	}
)

func init() {
	if options == nil {
		options = defaultOptions
	}

	application = app.New("Ome", "ome-store", options...)
	command = application.GetCommand()

	flags := command.PersistentFlags()
	flags.StringVar(&dsn, "dsn", "bome:bome@(127.0.0.1:3306)/bome?charset=utf8", "MySQL database source name")
	versionCMD := &cobra.Command{
		Use:   "version",
		Short: "Version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println()
			fmt.Println("   Version: ", Version)
			fmt.Println("  Revision: ", Revision)
			fmt.Println("Build date: ", BuildDate)
			fmt.Println("   License: ", License)
			fmt.Println()
		},
	}
	command.AddCommand(versionCMD)

	startCommand := application.StartCommand()
	flags = startCommand.PersistentFlags()
	flags.StringVar(&addr, "addr", ":8080", "Http server bind address")
	flags.StringVar(&certFilename, "cert", "", "Certificate file path")
	flags.StringVar(&keyFilename, "key", "", "Key file path")
	flags.BoolVar(&selfSignedTLS, "ss-tls", false, "Is certificate self-signed")
}

func Get() *cobra.Command {
	return command
}
