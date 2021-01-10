package main

import (
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	"github.com/omecodes/common/utils/prompt"
	oms "github.com/omecodes/store"
)

var (
	jwtSecret string
	dev       bool
	regPort   int
	caPort    int
	apiPort   int
	bindIP    string
	domains   []string
	dbURI     string
	wDir      string
	cmd       *cobra.Command
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
		Short: "Start master node of Omestore micro service",
		Run:   run,
	}

	flags := startCMD.PersistentFlags()
	flags.BoolVar(&dev, "dev", false, "Runs in development mode when the flag is set")
	flags.IntVar(&regPort, "reg", 9090, "Registry server port")
	flags.IntVar(&caPort, "ca", 9092, "CA server port")
	flags.IntVar(&apiPort, "api", 8080, "API server port")
	flags.StringVar(&bindIP, "ip", "", "Http server address")
	flags.StringArrayVar(&domains, "domain", nil, "Domain names")
	flags.StringVar(&jwtSecret, "jwt-secret", "", "Secret used to verify JWT hmac based signature")
	flags.StringVar(&dbURI, "db-uri", "bome:bome@(127.0.0.1:3306)/bome?charset=utf8", "MySQL database source name")

	if err := cobra.MarkFlagRequired(flags, "ip"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "domain"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "jwt-secret"); err != nil {
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
	config := oms.MsConfig{
		Name:         "om-master",
		BindIP:       bindIP,
		Domains:      domains,
		RegistryPort: regPort,
		CAPort:       caPort,
		APIPort:      apiPort,
		DBUri:        dbURI,
		WorkingDir:   wDir,
		JWTSecret:    jwtSecret,
		Dev:          dev,
	}

	server := oms.NewMSServer(config)

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
