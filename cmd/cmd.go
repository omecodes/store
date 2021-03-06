package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	domains                    []string
	ip                         string
	caPort, registryPort, port int
	name                       string
	dev                        bool
	caSecret                   string
	caCert                     string
	caAddress                  string
	registryAddress            string

	adminAuth string

	certFilename string
	keyFilename  string

	autoCert   bool
	enableTLS  bool
	workingDir string
	fsDir      string
	wwwDir     string
	dbURI      string
)

func init() {
	CMD.AddCommand(VersionCMD, MonolithicCMD, ServiceCMD)
}

var CMD = &cobra.Command{
	Use:   "store",
	Short: "Manage your store instance",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			fmt.Println(err)
		}
	},
}
