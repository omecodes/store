package files

import "context"

type ctxSourceManager struct{}
type ctxSource struct{}
type ctxFS struct{}

func ContextWithSourceManager(parent context.Context, manager SourceManager) context.Context {
	return context.WithValue(parent, ctxSourceManager{}, manager)
}

func GetFS(ctx context.Context) FS {
	o := ctx.Value(ctxFS{})
	if o == nil {
		return nil
	}
	return o.(FS)
}

func GetSourceManager(ctx context.Context) SourceManager {
	o := ctx.Value(ctxSourceManager{})
	if o == nil {
		return nil
	}
	return o.(SourceManager)
}

func ContextWithFS(parent context.Context, fs FS) context.Context {
	return context.WithValue(parent, ctxFS{}, fs)
}
