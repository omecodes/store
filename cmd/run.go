package cmd

import (
	"context"
	"fmt"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/omestore/events"
	"github.com/omecodes/omestore/server"
	"github.com/omecodes/service"
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

	s := server.New(server.Config{
		DSN: dsn,
		Box: box,
	})

	err = s.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	defer s.Stop()
	<-prompt.QuitSignal()
}

func runEventsServer() {
	cfg := &events.Config{
		Address:   "",
		Table:     "",
		DBUri:     "",
		TlsConfig: nil,
	}

	s, err := events.Serve(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer func() {
		if err := s.Stop(); err != nil {
			log.Error("server stop caused error", log.Err(err))
		}
	}()

	<-prompt.QuitSignal()
}
