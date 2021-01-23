package admin

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/omecodes/store/auth"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/omecodes/common/utils/prompt"
)

func init() {
	flags := saveAccessCMD.PersistentFlags()
	flags.StringVar(&input, "in", "", "Input file containing sequence of JSON encoded access")
	flags.StringVar(&server, "server", "", "Server address")
	flags.StringVar(&password, "password", "", "admin password")
	if err := cobra.MarkFlagRequired(flags, "server"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "in"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	flags = getAccessesCMD.PersistentFlags()
	flags.StringVar(&server, "server", "", "Server address")
	flags.StringVar(&password, "password", "", "admin password")
	flags.StringVar(&output, "out", "", "Output file")
	if err := cobra.MarkFlagRequired(flags, "server"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	accessCMD.AddCommand(saveAccessCMD)
	accessCMD.AddCommand(getAccessesCMD)
}

var accessCMD = &cobra.Command{
	Use:   "access",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var saveAccessCMD = &cobra.Command{
	Use:   "save",
	Short: "save accesses",
	Run:   saveAccess,
}

var getAccessesCMD = &cobra.Command{
	Use:   "get",
	Short: "Get all accesses",
	Run:   getAllAccesses,
}

func saveAccess(cmd *cobra.Command, args []string) {
	var err error
	if password == "" {
		password, err = prompt.Password("password")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}

	file, err := os.Open(input)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	defer func() {
		_ = file.Close()
	}()

	for {
		var access *auth.APIAccess
		err = json.NewDecoder(file).Decode(&access)
		if err == io.EOF {
			return
		}

		h := sha512.New()
		secret := h.Sum([]byte(access.Secret))
		access.Secret = hex.EncodeToString(secret)
	}
}

func getAllAccesses(cmd *cobra.Command, args []string) {

}
