package objects

import (
	"github.com/omecodes/store/pb"
	"io"
)

type idsListCursor struct {
	ids           []string
	getObjectFunc func(string) (*pb.Object, error)
	pos           int
}

func (i *idsListCursor) Browse() (*pb.Object, error) {
	if i.pos == len(i.ids) {
		return nil, io.EOF
	}

	id := i.ids[i.pos]
	i.pos++

	return i.getObjectFunc(id)
}

func (i *idsListCursor) Close() error {
	return nil
}
