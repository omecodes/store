package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/omestore/info"
	"github.com/omecodes/omestore/store"
	"github.com/spf13/cobra"
	"os"
)

var (
	port int
	dsn  string
	host string
	dir  string
	cid string
	secret string
)

func init() {
	flags := com.PersistentFlags()
	flags.StringVar(&cid, "cid", "", "Ome client ID")
	flags.StringVar(&secret, "secret", "", "Ome client secret")
	flags.StringVar(&host, "host", "127.0.0.1", "Domain or IP")
	flags.StringVar(&dir, "d", "", "Data directory")
	flags.StringVar(&dsn, "m", "root:toor@(127.0.0.1:3306)/omestore?charset=utf8", "MySQL database source name")
	flags.IntVar(&port, "p", 80, "HTTP server port")

	_ = cobra.MarkFlagRequired(flags, "cid")
	_ = cobra.MarkFlagRequired(flags, "secret")

	com.AddCommand(&cobra.Command{
		Use: "version",
		Short: "Version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println()
			fmt.Println("   Version: ", info.Version)
			fmt.Println("  Revision: ", info.Revision)
			fmt.Println("Build date: ", info.BuildDate)
			fmt.Println("   License: ", info.License)
			fmt.Println()
		},
	})
}

var com = &cobra.Command{
	Use:   "omestore",
	Short: "Runs omestore server",
	Run: func(cmd *cobra.Command, args []string) {
		server := store.NewServer(store.Config{
			Dir:    dir,
			DSN:    dsn,
			Domain: host,
			Port:   port,
		})
		err := server.Start()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		defer server.Stop()
		<-prompt.QuitSignal()
	},
}

func main() {
	fmt.Printf("running store at %s:%d\n", host, port)
	err := com.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
