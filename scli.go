package main

import (
	"fmt"
	"github.com/omecodes/store/cli"
	"os"
)

func main() {
	if err := cli.CMD.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
