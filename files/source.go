package files

import (
	"context"
)

const activeUserVar = "{user}"

const (
	SchemeFS     = "files"
	SchemeSource = "ref"
	SchemeHTTP   = "http"
	SchemeHTTPS  = "https"
	SchemeAWS    = "aws"
)

type ctxSourceManager struct{}

type SourceManager interface {
	Save(ctx context.Context, source *Source) (string, error)
	Get(ctx context.Context, id string) (*Source, error)
	Delete(ctx context.Context, id string) error
	UserSources(ctx context.Context, username string) ([]*Source, error)
}

func ContextWithSourceManager(parent context.Context, manager SourceManager) context.Context {
	return context.WithValue(parent, ctxSourceManager{}, manager)
}

func getSourceManager(ctx context.Context) SourceManager {
	o := ctx.Value(ctxSourceManager{})
	if o == nil {
		return nil
	}
	return o.(SourceManager)
}
