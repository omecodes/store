package main

import (
	"github.com/omecodes/store/cmd"
	"os"
)

func main() {
	if err := cmd.CMD.Execute(); err != nil {
		os.Exit(-1)
	}
}