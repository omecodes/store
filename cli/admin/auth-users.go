package admin

import (
	"encoding/json"
	"fmt"
	"github.com/omecodes/libome/logs"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/libome/crypt"
	"github.com/omecodes/store/auth"
)

func init() {
	flags := saveUserCMD.PersistentFlags()
	flags.StringVar(&input, "in", "", "Input file containing sequence of JSON encoded access")
	flags.StringVar(&server, "server", "", "Server API location")
	flags.StringVar(&password, "password", "", "admin password")
	if err := cobra.MarkFlagRequired(flags, "server"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "in"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	flags = getUsersCMD.PersistentFlags()
	flags.StringVar(&server, "server", "", "Server API location")
	flags.StringVar(&password, "password", "", "admin password")
	flags.StringVar(&output, "out", "", "Output file")
	if err := cobra.MarkFlagRequired(flags, "server"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	flags = deleteUserCMD.PersistentFlags()
	flags.StringVar(&server, "server", "", "Server API location")
	flags.StringVar(&password, "password", "", "admin password")
	flags.StringArrayVar(&ids, "id", nil, "Access ID")
	if err := cobra.MarkFlagRequired(flags, "server"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "id"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	usersCMD.AddCommand(saveUserCMD)
	usersCMD.AddCommand(getUsersCMD)
	usersCMD.AddCommand(deleteUserCMD)
}

var usersCMD = &cobra.Command{
	Use:   "users",
	Short: "Manage store accesses",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var saveUserCMD = &cobra.Command{
	Use:   "new",
	Short: "Save user credentials",
	Run: func(cmd *cobra.Command, args []string) {
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
			var userCredentials *auth.UserCredentials
			err = decoder.Decode(&userCredentials)
			if err == io.EOF {
				return
			}

			if userCredentials.Password == "" {
				userCredentials.Password, err = crypt.GenerateVerificationCode(16)
				if err != nil {
					fmt.Println(err)
					os.Exit(-1)
				}
				logs.Info("Generated password for user", logs.Details("user", userCredentials.Username), logs.Details("password", userCredentials.Password))
			}

			err = putUser(password, userCredentials)
			if err != nil {
				fmt.Println(err)
			}
		}
	},
}

var getUsersCMD = &cobra.Command{
	Use:   "get",
	Short: "Get all accesses",
	Run: func(cmd *cobra.Command, args []string) {
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
	},
}

var deleteUserCMD = &cobra.Command{
	Use:   "del",
	Short: "Delete accesses",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if password == "" {
			password, err = prompt.Password("password")
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
		}

		for _, id := range ids {
			err = deleteAccess(password, id)
			if err != nil {
				fmt.Println(err)
			}
		}
	},
}
