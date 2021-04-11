package cli

import (
	"encoding/json"
	"fmt"
	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/store/client"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	CMD            *cobra.Command
	authentication string

	apiLocation string
	input       string
	port        int
	output      string
	ids         []string
	noTLS       bool

	name   string
	format string
)

func init() {
	CMD = &cobra.Command{
		Use:   "scli",
		Short: "Store command line tool",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	flags := CMD.PersistentFlags()
	flags.StringVar(&authentication, "auth", "", "User authentication <user:password>")
	flags.BoolVar(&noTLS, "no-tls", false, "Uses insecure connection")
	flags.IntVar(&port, "p", 443, "Store server API port")
	flags.StringVar(&apiLocation, "api-location", "/api", "API path")

	CMD.AddCommand(authCMD)
	CMD.AddCommand(objectsCMD)
	CMD.AddCommand(filesCMD)
}

func promptAuthentication() (username, password string, err error) {
	if len(authentication) != 0 {
		parts := strings.Split(strings.Trim(authentication, " "), ":")
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
	}

	username, err = prompt.Text("username", false)
	if err != nil {
		return
	}
	password, err = prompt.Password(fmt.Sprintf("%s's password", username))
	return
}

func newClient() *client.Client {
	username, password, err := promptAuthentication()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	var opts []client.Option
	opts = append(opts, client.WithUserBasicAuthentication(username, password))
	opts = append(opts, client.WithAPILocation(apiLocation))
	if port > 0 {
		opts = append(opts, client.WithPort(port))
	}
	if noTLS {
		opts = append(opts, client.WithoutTLS())
	}

	return client.New("localhost", opts...)
}

func writeToFile(o interface{}, filename string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	defer func() {
		_ = file.Close()
	}()
	_ = json.NewEncoder(file).Encode(o)
}
