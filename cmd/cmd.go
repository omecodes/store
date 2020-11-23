package cmd

import (
	"fmt"
	"github.com/omecodes/common/env/app"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/service/v2"
	"github.com/spf13/cobra"
)

var (
	addr           string
	dsn            string
	cmdParams      service.Params
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
			fmt.Println("   Version: ", oms.Version)
			fmt.Println("  Revision: ", oms.Revision)
			fmt.Println("Build date: ", oms.BuildDate)
			fmt.Println("   License: ", oms.License)
			fmt.Println()
		},
	}

	eventsCMD := &cobra.Command{
		Use:   "events",
		Short: "Runs the events server",
		Run: func(cmd *cobra.Command, args []string) {
			runEventsServer()
		},
	}
	flags = eventsCMD.PersistentFlags()
	flags.StringVar(&addr, "addr", "", "Events server bind address")

	command.AddCommand(eventsCMD, versionCMD)

	startCommand := application.StartCommand()
	service.SetCMDFlags(startCommand, &cmdParams, true)
}

func Get() *cobra.Command {
	return command
}
