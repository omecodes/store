package cli

import (
	"encoding/json"
	"fmt"
	pb "github.com/omecodes/store/gen/go/proto"
	"github.com/spf13/cobra"
	"io"
	"os"
)

func init() {
	flags := createFileSourceCMD.PersistentFlags()
	flags.StringVar(&input, "in", "", "Input file containing sequence of JSON encoded collection")
	if err := cobra.MarkFlagRequired(flags, "in"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	flags = deleteFileSourcesCMD.PersistentFlags()
	flags.StringArrayVar(&ids, "id", nil, "source id")
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
		cl := newClient()

		sources, err := cl.ListFileSources()
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		writeToFile(sources, "sources.json")
	},
}

var createFileSourceCMD = &cobra.Command{
	Use:   "new",
	Short: "Create one or many file sources",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		file, err := os.Open(input)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		defer func() {
			_ = file.Close()
		}()

		cl := newClient()
		decoder := json.NewDecoder(file)
		for {
			var access *pb.Access
			err = decoder.Decode(&access)
			if err == io.EOF {
				return
			}

			err = cl.CreateFileAccess(access)
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

		cl := newClient()
		for _, id := range ids {
			err = cl.DeleteFilesSource(id)
			if err != nil {
				fmt.Println(err)
			}
		}
	},
}
