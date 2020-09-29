package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/omestore/store"
	"github.com/spf13/cobra"
	"os"
)

var (
	port int
	dsn  string
	host string
	dir  string
)

func init() {
	flags := com.PersistentFlags()
	flags.StringVar(&host, "h", "127.0.0.1", "Domain or IP")
	flags.StringVar(&dir, "d", "", "Data directory")
	flags.StringVar(&dsn, "m", "omestore:omestore@(127.0.0.1:3306)/omestore?charset=utf8", "MySQL database source name")
	flags.IntVar(&port, "p", 80, "HTTP server port")
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
	err := com.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
