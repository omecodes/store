package cmd

import (
	"fmt"
	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/store/service"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
)

func init() {
	ServiceCMD.AddCommand(frontServiceCMD, aclServiceCMD, accessServiceCMD, sourcesServiceCMD, filesServiceCMD, objectsServiceCMD)
}

var ServiceCMD = &cobra.Command{
	Use:   "service",
	Short: "Runs a service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	flags := frontServiceCMD.PersistentFlags()
	flags.StringArrayVar(&domains, "domains", nil, "domain name list")
	flags.StringVar(&ip, "ip", "", "Bind ip")
	flags.StringVar(&adminAuth, "admin", "", "Admin authentication data")
	flags.BoolVar(&autoCert, "auto-cert", false, "Run TLS server with auto generated certificate/key pair")
	flags.BoolVar(&enableTLS, "tls", false, "Enable TLS secure connexion")
	flags.StringVar(&certFilename, "cert", "", "Certificate filename")
	flags.StringVar(&keyFilename, "key", "", "Key filename")
	flags.StringVar(&wwwDir, "www", "./www", "Web apps directory (apache www equivalent)")
	flags.StringVar(&dbURI, "db", "store:store@(127.0.0.1:3306)/store?charset=utf8", "MySQL database uri")
	flags.IntVar(&port, "port", 8080, "Gateway HTTP port")
	flags.IntVar(&caPort, "ca-port", 9090, "CA server gRPC port")
	flags.IntVar(&registryPort, "reg-port", 9091, "Registry server gRPC port")
	flags.StringVar(&name, "name", "", "Service name")
	flags.BoolVar(&dev, "dev", false, "Enables CORS")

	err := cobra.MarkFlagRequired(flags, "domains")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	err = cobra.MarkFlagRequired(flags, "ip")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	err = cobra.MarkFlagRequired(flags, "db")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

var frontServiceCMD = &cobra.Command{
	Use:   "front",
	Short: "Runs front service",
	Run: func(cmd *cobra.Command, args []string) {
		config := service.FrontConfig{
			Dev:             dev,
			Name:            name,
			Domains:         domains,
			AdminAuth:       adminAuth,
			IP:              ip,
			CAPort:          caPort,
			RegistryPort:    registryPort,
			Port:            port,
			WorkingDir:      "",
			Database:        dbURI,
			TLS:             enableTLS,
			TLSAuto:         autoCert,
			TLSCertFilename: certFilename,
			TLSKeyFilename:  keyFilename,
		}
		fs := service.NewFront(config)
		err := fs.Start()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		<-prompt.QuitSignal()
	},
}

func parseServiceFlags(flags *pflag.FlagSet, withDB bool) error {
	flags.StringArrayVar(&domains, "domains", nil, "domain name list")
	flags.StringVar(&ip, "ip", "", "Bind ip")
	if withDB {
		flags.StringVar(&dbURI, "db", "", "MySQL database uri")
	}
	flags.StringVar(&caAddress, "ca-addr", "", "CA server address")
	flags.StringVar(&caCert, "ca-cert", "", "CA server certificate file")
	flags.StringVar(&caSecret, "ca-secret", "", "CA server access shared secret")
	flags.StringVar(&registryAddress, "rg-addr", "", "Registry server address")
	flags.StringVar(&name, "name", "", "Service name")

	err := cobra.MarkFlagRequired(flags, "domains")
	if err != nil {
		return err
	}
	err = cobra.MarkFlagRequired(flags, "ip")
	if err != nil {
		return err
	}
	if withDB {
		err = cobra.MarkFlagRequired(flags, "db")
		if err != nil {
			return err
		}
	}
	err = cobra.MarkFlagRequired(flags, "ca-addr")
	if err != nil {
		return err
	}
	err = cobra.MarkFlagRequired(flags, "ca-cert")
	if err != nil {
		return err
	}
	err = cobra.MarkFlagRequired(flags, "ca-secret")
	if err != nil {
		return err
	}

	return nil
}

func init() {
	flags := aclServiceCMD.PersistentFlags()
	err := parseServiceFlags(flags, true)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	flags = accessServiceCMD.PersistentFlags()
	err = parseServiceFlags(flags, false)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	flags = filesServiceCMD.PersistentFlags()
	err = parseServiceFlags(flags, false)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	flags = objectsServiceCMD.PersistentFlags()
	err = parseServiceFlags(flags, true)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	flags = sourcesServiceCMD.PersistentFlags()
	err = parseServiceFlags(flags, true)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

var aclServiceCMD = &cobra.Command{
	Use:   "acl",
	Short: "Runs ACL service",
	Run: func(cmd *cobra.Command, args []string) {
		config := service.ACLConfig{
			Name:            name,
			Domain:          domains[0],
			IP:              ip,
			CAAddress:       caAddress,
			CASecret:        caSecret,
			CACertFilename:  caCert,
			RegistryAddress: registryAddress,
			Database:        dbURI,
		}

		as := service.NewACL(config)
		err := as.Start()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		//defer as.Stop()
		<-prompt.QuitSignal()
	},
}

var accessServiceCMD = &cobra.Command{
	Use:   "access",
	Short: "Runs access service",
	Run: func(cmd *cobra.Command, args []string) {
		config := service.AccessConfig{
			Name:            name,
			Domain:          domains[0],
			IP:              ip,
			CAAddress:       caAddress,
			CASecret:        caSecret,
			CACertFilename:  caCert,
			RegistryAddress: registryAddress,
		}

		as := service.NewAccess(config)
		err := as.Start()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		//defer as.Stop()
		<-prompt.QuitSignal()
	},
}

var sourcesServiceCMD = &cobra.Command{
	Use:   "sources",
	Short: "Runs sources service",
	Run: func(cmd *cobra.Command, args []string) {
		config := service.FileAccessesConfig{
			Name:            name,
			Domain:          domains[0],
			IP:              ip,
			CAAddress:       caAddress,
			CASecret:        caSecret,
			CACertFilename:  caCert,
			RegistryAddress: registryAddress,
			Database:        dbURI,
		}

		ss := service.NewSources(config)
		err := ss.Start()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		//defer as.Stop()
		<-prompt.QuitSignal()
	},
}

var filesServiceCMD = &cobra.Command{
	Use:   "files",
	Short: "Runs files service",
	Run: func(cmd *cobra.Command, args []string) {
		config := service.FilesConfig{
			Name:            name,
			Domain:          domains[0],
			IP:              ip,
			CAAddress:       caAddress,
			CASecret:        caSecret,
			CACertFilename:  caCert,
			RegistryAddress: registryAddress,
		}

		fs := service.NewFiles(config)
		err := fs.Start()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		//defer as.Stop()
		<-prompt.QuitSignal()
	},
}

var objectsServiceCMD = &cobra.Command{
	Use:   "objects",
	Short: "Runs objects service",
	Run: func(cmd *cobra.Command, args []string) {
		config := service.ObjectsConfig{
			Name:            name,
			Domain:          domains[0],
			IP:              ip,
			CAAddress:       caAddress,
			CASecret:        caSecret,
			CACertFilename:  caCert,
			RegistryAddress: registryAddress,
			Database:        dbURI,
		}

		fs := service.NewObjects(config)
		err := fs.Start()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		//defer as.Stop()
		<-prompt.QuitSignal()
	},
}
