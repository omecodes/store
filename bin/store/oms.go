package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/omecodes/omestore/cmd"
)

func main() {
	err := cmd.Get().Execute()
	if err != nil {
		fmt.Println(err)
	}
}
