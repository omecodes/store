package cli

import (
	"encoding/json"
	"fmt"
	"github.com/omecodes/store/objects"
	"github.com/spf13/cobra"
	"io"
	"os"
)

func init() {
	flags := createCollectionsCMD.PersistentFlags()
	flags.StringVar(&input, "in", "", "Input file containing sequence of JSON encoded collection")
	if err := cobra.MarkFlagRequired(flags, "in"); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	objectsCollectionsCMD.AddCommand(createCollectionsCMD)
	objectsCollectionsCMD.AddCommand(getCollectionsCMD)
}

var objectsCollectionsCMD = &cobra.Command{
	Use:   "collections",
	Short: "Manage collections",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var getCollectionsCMD = &cobra.Command{
	Use:   "get",
	Short: "List all collections",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		err = listCollections(output)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	},
}

var createCollectionsCMD = &cobra.Command{
	Use:   "new",
	Short: "Create a new collection",
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

		decoder := json.NewDecoder(file)
		for {
			var col *objects.Collection
			err = decoder.Decode(&col)
			if err == io.EOF {
				return
			}

			err = putCollections(col)
			if err != nil {
				fmt.Println(err)
			}
		}
	},
}