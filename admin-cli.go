package main

import (
	"fmt"
	"os"

	"github.com/omecodes/store/apps/cli/admin"
)

func main() {
	if err := admin.CMD().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
