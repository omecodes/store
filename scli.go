package main

import (
	"github.com/omecodes/store/cli"
	"os"
)

func main() {
	if err := cli.CMD.Execute(); err != nil {
		os.Exit(-1)
	}
}
