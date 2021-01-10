package admin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/libome/crypt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var (
	passwordLen int
	rounds      int
	name        string
	format      string
	Cmd         *cobra.Command
)

func init() {
	Cmd = &cobra.Command{
		Use:   "security",
		Short: "Security utility",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	password := &cobra.Command{
		Use:   "admin",
		Short: "Generates a new password for admin user",
		Run: func(cmd *cobra.Command, args []string) {
			phrase, err := prompt.Text("pass phrase", false)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			_, info, err := crypt.Generate(phrase, passwordLen)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}

			data, err := json.Marshal(info)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}

			encoded := base64.RawStdEncoding.EncodeToString(data)

			err = ioutil.WriteFile(name, []byte(encoded), os.ModePerm)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
		},
	}

	flags := password.PersistentFlags()
	flags.IntVar(&passwordLen, "len", 16, "Password length")
	flags.IntVar(&rounds, "rnd", 50000, "Password derivation rounds count")
	flags.StringVar(&name, "out", "admin", "output filename")

	Cmd.AddCommand(password)
}
