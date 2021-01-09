package search

import "github.com/omecodes/store/pb"

type Cursor interface {
	Next() (string, error)
	Close() error
}

type Store interface {
	SaveWordMapping(word string, field string, ids ...string) error
	SaveNumberMapping(num int64, field string, ids ...string) error
	Search(expression *pb.BooleanExp) (Cursor, error)
}
