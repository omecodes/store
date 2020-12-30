package oms

import (
	"context"
)

type JSON map[string]interface{}

type Selector interface {
	Select(j JSON) (bool, error)
}

type SelectorFunc func(JSON) (bool, error)

func (f SelectorFunc) Select(j JSON) (bool, error) {
	return f(j)
}

type Collection interface {
	Save(ctx context.Context, item *CollectionItem) error
	Select(ctx context.Context, before int64, count int, selector Selector) ([]*CollectionItem, error)
}

type CollectionItem struct {
	Id   string `json:"id"`
	Date int64  `json:"date"`
	Data string `json:"data"`
}
