package admin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/libome/crypt"
)

func init() {
	flags := authCMD.PersistentFlags()
	flags.IntVar(&passwordLen, "len", 16, "Password length")
	flags.IntVar(&rounds, "rnd", 50000, "Password derivation rounds count")
	flags.StringVar(&output, "out", "admin", "output filename")
}

var authCMD = &cobra.Command{
	Use:   "auth",
	Short: "Generates a new password for admin user",
	Run:   generateAdminAuth,
}

func generateAdminAuth(cmd *cobra.Command, args []string) {
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
}
