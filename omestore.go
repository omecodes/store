package main

import (
	"context"
	"fmt"
	"os"

	"github.com/omecodes/service"

	_ "github.com/go-sql-driver/mysql"
	"github.com/omecodes/common/env/app"
	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/omestore/info"
	"github.com/omecodes/omestore/store"
	"github.com/spf13/cobra"
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

	application = app.New("Ome", "firedata", options...)
	command = application.GetCommand()
	command.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println()
			fmt.Println("   Version: ", info.Version)
			fmt.Println("  Revision: ", info.Revision)
			fmt.Println("Build date: ", info.BuildDate)
			fmt.Println("   License: ", info.License)
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
	cmdParams.Name = "firedata"
	cmdParams.NoRegistry = true

	box, err := service.CreateBox(ctx, &cmdParams)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	server := store.NewServer(store.Config{
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
