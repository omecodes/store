package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/omecodes/common/utils/log"
	"github.com/omecodes/common/utils/prompt"
	oms "github.com/omecodes/store"
	"github.com/omecodes/store/cmd/admin"
	"github.com/spf13/cobra"
)

var (
	dev        bool
	adminInfo  string
	workingDir string
	fsDir      string
	resDir     string
	webappsDir string
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
			if !dev && len(domains) == 0 {
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

			if fsDir != "" {
				fsDir, err = filepath.Abs(fsDir)
				if err != nil {
					fmt.Println(err)
					os.Exit(-1)
				}
				err = os.MkdirAll(fsDir, os.ModePerm)
				if err != nil {
					fmt.Println(err)
					os.Exit(-1)
				}
			}

			if webappsDir != "" {
				webappsDir, err = filepath.Abs(webappsDir)
				if err != nil {
					fmt.Println(err)
					os.Exit(-1)
				}
				err = os.MkdirAll(webappsDir, os.ModePerm)
				if err != nil {
					fmt.Println(err)
					os.Exit(-1)
				}
			}

			if resDir != "" {
				resDir, err = filepath.Abs(resDir)
				if err != nil {
					fmt.Println(err)
					os.Exit(-1)
				}
				err = os.MkdirAll(resDir, os.ModePerm)
				if err != nil {
					fmt.Println(err)
					os.Exit(-1)
				}
			}

			s := oms.NewServer(oms.MNConfig{
				Dev:            dev,
				Domains:        domains,
				FSRootDir:      fsDir,
				WorkingDir:     workingDir,
				WebAppsDir:     webappsDir,
				StaticFilesDir: resDir,
				AdminInfo:      adminInfo,
				DSN:            dsn,
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
	flags.BoolVar(&dev, "dev", false, "Enable development mode")
	flags.StringArrayVar(&domains, "domains", nil, "Domains name for auto cert")
	flags.StringVar(&workingDir, "dir", "", "Data directory")
	flags.StringVar(&fsDir, "fs", "./files", "Files root directory")
	flags.StringVar(&resDir, "res", "./static", "Files root directory")
	flags.StringVar(&webappsDir, "apps", "./apps", "Files root directory")
	flags.StringVar(&adminInfo, "admin", "", "Admin password info")
	flags.StringVar(&dsn, "dsn", "store:store@(127.0.0.1:3306)/store?charset=utf8", "MySQL database uri")
	if err := cobra.MarkFlagRequired(flags, "admin"); err != nil {
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
	command.AddCommand(admin.Cmd)
}

func main() {
	err := command.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
