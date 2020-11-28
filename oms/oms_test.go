package oms

import (
	"flag"
	"fmt"
	"github.com/omecodes/bome"
	"os"
	"strings"
)

var (
	testDBUri       string
	jsonTestEnabled bool
	testDialect     string
)

func init() {
	testDBUri = os.Getenv("OMS_TESTS_DB")
	if testDBUri == "" {
		testDBUri = "objects.db"
	}

	jsonTestEnabled = "1" == os.Getenv("OMS_JSON_TESTS_ENABLED")

	testDialect = os.Getenv("OMS_TESTS_DIALECT")
	if testDialect == "" {
		testDialect = bome.SQLite3
	}

	if flag.Lookup("test.v") != nil || strings.HasSuffix(os.Args[0], ".test") || strings.Contains(os.Args[0], "/_test/") {
		fmt.Println()
		fmt.Println()
		fmt.Println("TESTS_DIALECT: ", testDialect)
		fmt.Println("TESTS_DB     : ", testDBUri)
		fmt.Println("TESTS_ENABLED: ", jsonTestEnabled)
		fmt.Println()
		fmt.Println()
	}
}
