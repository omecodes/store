package main

import (
	"fmt"
	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/common/utils/prompt"
	oms "github.com/omecodes/omestore"
	"github.com/spf13/cobra"
	"os"
	"path"
	"path/filepath"
)

var (
	autoCert   bool
	jwtSecret  string
	workingDir string
	dsn        string
	domains    []string
	command    *cobra.Command
)

func init() {
	command = &cobra.Command{
		Use:   path.Base(os.Args[0]),
		Short: "Monolithic object backend",
		Run: func(cmd *cobra.Command, args []string) {
			_ = command.Help()
		},
	}

	runCMD := &cobra.Command{
		Use:   "run",
		Short: "Runs a objects backend application",
		Run: func(cmd *cobra.Command, args []string) {
			if autoCert && len(domains) == 0 {
				fmt.Println("Flag --domains is required when --auto-cert is set")
				os.Exit(-1)
			}

			var err error
			if workingDir == "" {
				workingDir, err = filepath.Abs("./")
				if err != nil {
					fmt.Println(err)
					os.Exit(-1)
				}
			}

			s := oms.NewMNServer(oms.MNConfig{
				WorkingDir: workingDir,
				Domains:    domains,
				AutoCert:   autoCert,
				JwtSecret:  jwtSecret,
				DSN:        dsn,
			})

			err = s.Start()
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
		},
	}

	flags := runCMD.PersistentFlags()
	flags.BoolVar(&autoCert, "auto-cert", false, "Secure listen using LetsEncrypt certificate")
	flags.StringArrayVar(&domains, "domains", nil, "Domains name for auto cert")
	flags.StringVar(&workingDir, "dir", "", "Data directory")
	flags.StringVar(&jwtSecret, "jwt-secret", "", "Secret used to verify JWT hmac based signature")
	flags.StringVar(&dsn, "dsn", "oms:oms@(127.0.0.1:3306)/oms?charset=utf8", "MySQL database uri")
	if err := cobra.MarkFlagRequired(flags, "jwt-secret"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	command.AddCommand(runCMD)

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
	command.AddCommand(versionCMD)
}

func main() {
	err := command.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
