package files

import "context"

type ctxSourceManager struct{}
type ctxSources struct{}
type ctxFS struct{}
type ctxSourceResolver struct{}

func ContextWithSourceManager(parent context.Context, manager SourceManager) context.Context {
	return context.WithValue(parent, ctxSourceManager{}, manager)
}

func ContextWithFS(parent context.Context, fs FS) context.Context {
	return context.WithValue(parent, ctxFS{}, fs)
}

func ContextWithSource(parent context.Context, source *Source) context.Context {
	o := parent.Value(ctxSources{})
	if o == nil {
		m := make(map[string]*Source)
		return context.WithValue(parent, ctxSources{}, m)
	}
	m := o.(map[string]*Source)
	m[source.ID] = source
	return parent
}

func GetFS(ctx context.Context) FS {
	o := ctx.Value(ctxFS{})
	if o == nil {
		return nil
	}
	return o.(FS)
}

func GetSource(ctx context.Context, id string) *Source {
	o := ctx.Value(ctxSources{})
	if o == nil {
		return nil
	}

	m := o.(map[string]*Source)
	return m[id]
}

func GetSourceManager(ctx context.Context) SourceManager {
	o := ctx.Value(ctxSourceManager{})
	if o == nil {
		return nil
	}
	return o.(SourceManager)
}
