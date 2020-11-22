package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"

	"github.com/omecodes/common/env/app"
	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/service/v2"
)

var (
	dsn            string
	cmdParams      service.Params
	application    *app.App
	options        []app.Option
	command        *cobra.Command
	defaultOptions = []app.Option{
		app.WithRunCommandFunc(start),
		app.WithVersion("1.0.1"),
	}
)

func init() {
	if options == nil {
		options = defaultOptions
	}

	application = app.New("Ome", "ome-store", options...)
	command = application.GetCommand()
	command.AddCommand(&cobra.Command{
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
	})

	startCommand := application.StartCommand()
	service.SetCMDFlags(startCommand, &cmdParams, true)
	flags := startCommand.PersistentFlags()
	flags.StringVar(&dsn, "dsn", "root:toor@(127.0.0.1:3306)/firedata?charset=utf8", "MySQL database source name")
}

func start() {
	ctx := context.Background()
	cmdParams.Name = "oms"
	cmdParams.NoRegistry = true

	box, err := service.CreateBox(ctx, &cmdParams)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	server := oms.NewServer(oms.Config{
		DSN: dsn,
		Box: box,
	})

	err = server.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	defer server.Stop()
	<-prompt.QuitSignal()
}

func main() {
	err := command.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
