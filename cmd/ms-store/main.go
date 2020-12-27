package main

import (
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"

	"github.com/omecodes/common/utils/prompt"
	oms "github.com/omecodes/store"
)

var (
	id             string
	regAddr        string
	caAddr         string
	caApiKey       string
	caApiSecret    string
	caCertFilename string
	objectsPort    int
	aclPort        int
	bindIP         string
	domain         string
	wDir           string
	dbURI          string
	cmd            *cobra.Command
)

func init() {
	var err error
	wDir, err = filepath.Abs("./")
	if err != nil {
		fmt.Println("could not resolve working dir", err)
		os.Exit(-1)
	}

	execName := filepath.Base(os.Args[0])
	cmd = &cobra.Command{
		Use:   execName,
		Short: "oms",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	startCMD := &cobra.Command{
		Use:   "run",
		Short: "Start an Ome store node",
		Run:   run,
	}

	flags := startCMD.PersistentFlags()
	flags.StringVar(&id, "id", "", "Ome store node ID")
	flags.StringVar(&regAddr, "reg-addr", "", "Registry address")
	flags.StringVar(&caAddr, "ca-addr", "", "CA server address")
	flags.StringVar(&caCertFilename, "ca-cert", "", "CA certificate")
	flags.StringVar(&caApiKey, "ca-api-key", "", "CA access API key")
	flags.StringVar(&caApiSecret, "ca-api-secret", "", "CA access API secret")
	flags.StringVar(&dbURI, "db-uri", "store:store@tcp(localhost:3306)/store?charset=utf8", "MySQL database uri")
	flags.IntVar(&objectsPort, "objects-port", 8080, "API server port")
	flags.IntVar(&aclPort, "acl-port", 8180, "API server port")
	flags.StringVar(&bindIP, "ip", "", "Http server address")
	flags.StringVar(&domain, "domain", "", "Domain name")

	if err := cobra.MarkFlagRequired(flags, "ip"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "domain"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "reg-addr"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "ca-addr"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "ca-cert"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "ca-api-key"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "ca-api-secret"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

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

	cmd.AddCommand(startCMD)
	cmd.AddCommand(versionCMD)
}

func run(cmd *cobra.Command, args []string) {
	config := &oms.StoreConfig{
		Name:           id,
		WorkingDir:     wDir,
		RegAddr:        regAddr,
		CaAddr:         caAddr,
		CaApiSecret:    caApiSecret,
		CaApiKey:       caApiKey,
		CaCertFilename: caCertFilename,
		Domain:         domain,
		IP:             bindIP,
		ObjectsPort:    objectsPort,
		ACLPort:        aclPort,
		DBUri:          dbURI,
	}

	server := oms.NewMSStore(config)

	err := server.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	defer func() {
		if err := server.Stop(); err != nil {
			fmt.Println(err)
		}
	}()

	<-prompt.QuitSignal()
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
