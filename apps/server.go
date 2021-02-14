package main

import (
	"fmt"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/store/common"
)

var (
	dev        bool
	autoCert   bool
	adminInfo  string
	workingDir string
	fsDir      string
	wwwDir     string
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

			if wwwDir != "" {
				wwwDir, err = filepath.Abs(wwwDir)
				if err != nil {
					fmt.Println(err)
					os.Exit(-1)
				}
				err = os.MkdirAll(wwwDir, os.ModePerm)
				if err != nil {
					fmt.Println(err)
					os.Exit(-1)
				}
			}

			s := store.New(store.Config{
				Dev:        dev,
				AutoCert:   autoCert,
				Domains:    domains,
				FSRootDir:  fsDir,
				WorkingDir: workingDir,
				WebDir:     wwwDir,
				AdminInfo:  adminInfo,
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
				logs.Error("server error", logs.Err(err))
			}
		},
	}

	flags := runCMD.PersistentFlags()
	flags.BoolVar(&dev, "dev", false, "Enable development mode")
	flags.BoolVar(&autoCert, "autocert", false, "Run with ACME autocert")
	flags.StringArrayVar(&domains, "domains", nil, "Domains name for auto cert")
	flags.StringVar(&workingDir, "dir", "./", "Data directory")
	flags.StringVar(&fsDir, "fs", "./files", "File storage root directory")
	flags.StringVar(&wwwDir, "web-dir", "./www", "Web apps directory (apache www equivalent)")
	flags.StringVar(&adminInfo, "admin", "", "Admin password info")
	flags.StringVar(&dsn, "db-uri", "bome:bome@(127.0.0.1:3306)/bome?charset=utf8", "MySQL database uri")
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
			fmt.Println("   Version: ", common.Version)
			fmt.Println("  Revision: ", common.Revision)
			fmt.Println("Build date: ", common.BuildDate)
			fmt.Println("   License: ", common.License)
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
