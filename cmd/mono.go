package cmd

import (
	"fmt"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/service"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/store/common"
)

func init() {
	flags := MonolithicCMD.PersistentFlags()
	flags.BoolVar(&dev, "dev", false, "Enable development mode. Enables CORS")
	flags.BoolVar(&autoCert, "auto-cert", false, "Run TLS server with auto generated certificate/key pair")
	flags.BoolVar(&enableTLS, "tls", false, "Enable TLS secure connexion")
	flags.StringVar(&certFilename, "cert", "", "Certificate filename")
	flags.StringVar(&keyFilename, "key", "", "Key filename")
	flags.StringArrayVar(&domains, "domains", nil, "Domains name for auto cert")
	flags.StringVar(&workingDir, "dir", "./", "Data directory")
	flags.StringVar(&fsDir, "fs", "./files", "File storage root directory")
	flags.StringVar(&wwwDir, "www", "./www", "Web apps directory (apache www equivalent)")
	flags.StringVar(&dbURI, "db", "store:store@(127.0.0.1:3306)/store?charset=utf8", "MySQL database uri")
}

var MonolithicCMD = &cobra.Command{
	Use:   "mono",
	Short: "Runs Store backend application",
	Run: func(cmd *cobra.Command, args []string) {
		if !dev && len(domains) == 0 && autoCert {
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

		s := service.New(service.Config{
			Dev:          dev,
			TLSAuto:      autoCert,
			Domains:      domains,
			FSRootDir:    fsDir,
			WorkingDir:   workingDir,
			WebDir:       wwwDir,
			CertFilename: certFilename,
			KeyFilename:  keyFilename,
			TLS:          enableTLS,
			DSN:          dbURI,
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

var VersionCMD = &cobra.Command{
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
