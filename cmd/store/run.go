package main

import (
	"crypto/tls"
	"fmt"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/libome/crypt"
	oms "github.com/omecodes/omestore"
	"os"
)

func runStore() {

	if jwtSecret == "" {
		if err := application.StartCommand().Help(); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		return
	}

	var tc *tls.Config
	if certFilename != "" || keyFilename != "" {
		if certFilename == "" {
			log.Fatal("missing certificate file path")
		}

		if keyFilename == "" {
			log.Fatal("missing key file path")
		}

		cert, err := crypt.LoadCertificate(certFilename)
		if err != nil {
			log.Fatal("loading certificate", log.Err(err))
		}

		key, err := crypt.LoadPrivateKey(nil, keyFilename)
		if err != nil {
			log.Fatal("loading key", log.Err(err))
		}

		tc = &tls.Config{
			Certificates: []tls.Certificate{{
				Certificate: [][]byte{cert.Raw},
				PrivateKey:  key,
			}},
		}
	}

	s := oms.NewMNServer(oms.MNConfig{
		JwtSecret:   jwtSecret,
		DSN:         dsn,
		BindAddress: addr,
		App:         application,
		TLS:         tc,
	})

	err := s.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer s.Stop()

	select {
	case <-prompt.QuitSignal():
	case err = <-s.Errors:
		log.Error("server error", log.Err(err))
	}
}
