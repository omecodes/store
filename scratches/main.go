package main

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	se "github.com/omecodes/store/search-engine"
)

func main() {
	query := &se.SearchQuery{
		Query: &se.SearchQuery_Text{
			Text: &se.StrQuery{
				Bool: &se.StrQuery_Contains{
					Contains: &se.Contains{
						Field: "",
						Value: "abid",
					},
				},
			},
		},
	}

	buff := bytes.NewBuffer(nil)
	_ = (&jsonpb.Marshaler{}).Marshal(buff, query)

	fmt.Println(string(buff.String()))
}
