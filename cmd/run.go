package cmd

import (
	"context"
	"fmt"
	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/omestore/events"
	"github.com/omecodes/omestore/oms"
	"github.com/omecodes/service/v2"
	"os"
)

func runStore() {
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

func runEventsServer() {
	cfg := &events.Config{
		Address:   "",
		Table:     "",
		DBUri:     "",
		TlsConfig: nil,
	}

	server, err := events.Serve(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer server.Stop()

	<-prompt.QuitSignal()
}
