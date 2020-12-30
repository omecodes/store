package oms

import "strings"

func sqlQuote(tableName string) string {
	tableName = strings.Replace(tableName, `'`, `''`, -1)
	if strings.Contains(tableName, `\`) {
		tableName = strings.Replace(tableName, `\`, `\\`, -1)
		tableName = ` E'` + tableName + `'`
	} else {
		tableName = `'` + tableName + `'`
	}
	return tableName
}
