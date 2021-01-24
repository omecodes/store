package admin

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/omecodes/libome/crypt"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common"
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

	flags = deleteAccessesCMD.PersistentFlags()
	flags.StringVar(&server, "server", "", "Server address")
	flags.StringVar(&password, "password", "", "admin password")
	flags.StringArrayVar(&accessIDs, "id", nil, "Access ID")
	if err := cobra.MarkFlagRequired(flags, "server"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "id"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	accessCMD.AddCommand(saveAccessCMD)
	accessCMD.AddCommand(getAccessesCMD)
	accessCMD.AddCommand(deleteAccessesCMD)
}

var accessCMD = &cobra.Command{
	Use:   "access",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var saveAccessCMD = &cobra.Command{
	Use:   "set",
	Short: "save accesses",
	Run:   saveAccess,
}

var getAccessesCMD = &cobra.Command{
	Use:   "get",
	Short: "Get all accesses",
	Run:   getAllAccesses,
}

var deleteAccessesCMD = &cobra.Command{
	Use:   "del",
	Short: "Delete accesses",
	Run:   deleteAccesses,
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

	decoder := json.NewDecoder(file)
	for {
		var access *auth.APIAccess
		err = decoder.Decode(&access)
		if err == io.EOF {
			return
		}

		key, info, err := crypt.Generate(password, 16)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		encrypted, err := crypt.AESGCMEncrypt(key, []byte(access.Secret))
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		encodedInfo, err := json.Marshal(info)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		h := sha512.New()
		secret := h.Sum([]byte(access.Secret))
		access.Secret = hex.EncodeToString(secret)

		access.Info = make(map[string]string)
		access.Info[common.AccessInfoEncryptedSecret] = hex.EncodeToString(encrypted)
		access.Info[common.AccessInfoSecretEncryptParams] = hex.EncodeToString(encodedInfo)

		err = putAccess(password, access)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func getAllAccesses(cmd *cobra.Command, args []string) {
	var err error
	if password == "" {
		password, err = prompt.Password("password")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}

	err = getAccesses(password, output)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func deleteAccesses(cmd *cobra.Command, args []string) {
	var err error
	if password == "" {
		password, err = prompt.Password("password")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}

	for _, id := range accessIDs {
		err = deleteAccess(password, id)
		if err != nil {
			fmt.Println(err)
		}
	}
}
