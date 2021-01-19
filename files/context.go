package files

import "context"

type ctxSourceManager struct{}
type ctxSource struct{}
type ctxFS struct{}

func ContextWithSourceManager(parent context.Context, manager SourceManager) context.Context {
	return context.WithValue(parent, ctxSourceManager{}, manager)
}

func GetSource(ctx context.Context) *Source {
	o := ctx.Value(ctxSource{})
	if o == nil {
		return nil
	}
	return o.(*Source)
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

func ContextWithSource(parent context.Context, source *Source) context.Context {
	return context.WithValue(parent, ctxSource{}, source)
}

func ContextWithFS(parent context.Context, fs FS) context.Context {
	return context.WithValue(parent, ctxFS{}, fs)
}
