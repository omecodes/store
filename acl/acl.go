package acl

import (
	"context"
)

type (
	Matcher interface {
		Matches(ctx context.Context, id string) (bool, error)
	}

	UserSet interface {
		Matcher
		ForEach(ctx context.Context, processFunc func(string) error) error
	}

	ObjectSet interface {
		Matcher
	}
)