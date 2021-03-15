package admin

import (
	"encoding/json"
	"fmt"
	"github.com/omecodes/common/utils/prompt"
	"github.com/omecodes/store/files"
	"github.com/spf13/cobra"
	"io"
	"os"
)

func init() {
	flags := createFileSourceCMD.PersistentFlags()
	flags.StringVar(&input, "in", "", "Input file containing sequence of JSON encoded collection")
	flags.StringVar(&server, "server", "", "Server API location")
	flags.StringVar(&password, "password", "", "Admin password")
	if err := cobra.MarkFlagRequired(flags, "server"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "in"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	flags = getFileSourcesCMD.PersistentFlags()
	flags.StringVar(&server, "server", "", "Server API location")
	flags.StringVar(&password, "password", "", "Admin password")
	if err := cobra.MarkFlagRequired(flags, "server"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	flags = deleteFileSourcesCMD.PersistentFlags()
	flags.StringVar(&server, "server", "", "Server API location")
	flags.StringVar(&password, "password", "", "admin password")
	flags.StringArrayVar(&ids, "id", nil, "source id")
	if err := cobra.MarkFlagRequired(flags, "server"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := cobra.MarkFlagRequired(flags, "id"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	fileSourcesCMD.AddCommand(createFileSourceCMD)
	fileSourcesCMD.AddCommand(getFileSourcesCMD)
	fileSourcesCMD.AddCommand(deleteFileSourcesCMD)
}

var fileSourcesCMD = &cobra.Command{
	Use:   "sources",
	Short: "Manage file sources",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var getFileSourcesCMD = &cobra.Command{
	Use:   "get",
	Short: "Get list of all file sources",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if password == "" {
			password, err = prompt.Password("password")
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
		}

		err = listFileSources(password, output)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	},
}

var createFileSourceCMD = &cobra.Command{
	Use:   "new",
	Short: "Create one or many file sources",
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
			var source *files.Source
			err = decoder.Decode(&source)
			if err == io.EOF {
				return
			}

			err = putFileSource(password, source)
			if err != nil {
				fmt.Println(err)
			}
		}
	},
}

var deleteFileSourcesCMD = &cobra.Command{
	Use:   "del",
	Short: "Delete one or many file sources",
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
			err = deleteFileSources(password, id)
			if err != nil {
				fmt.Println(err)
			}
		}
	},
}
