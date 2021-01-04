package objects

import (
	"context"
	"github.com/omecodes/store/pb"
)

type Collection interface {
	Save(ctx context.Context, createdAt int64, id string, data string) error
	Select(ctx context.Context, count int, filter ObjectFilter, resolver ObjectResolver) ([]*pb.Object, uint32, error)
	RangeSelect(ctx context.Context, after int64, before int64, count int, filter ObjectFilter, resolver ObjectResolver) ([]*pb.Object, uint32, error)
}

type CollectionItem struct {
	Id   string `json:"id"`
	Date int64  `json:"date"`
	Data string `json:"data"`
}
