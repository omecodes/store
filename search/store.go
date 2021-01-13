package se

import "github.com/omecodes/store/pb"

type Cursor interface {
	Next() (string, error)
	Close() error
}

type Store interface {
	SaveWordMapping(word string, id string) error
	SaveNumberMapping(num int64, id string) error
	SavePropertiesMapping(id string, value string) error
	Search(query *pb.SearchQuery) (Cursor, error)
	DeleteObjectMappings(id string) error
}
