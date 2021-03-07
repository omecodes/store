package session

import (
	"context"
	"github.com/gorilla/sessions"
)

type ctxCookieStore struct{}

func ContextWithCookieStore(parent context.Context, store *sessions.CookieStore) context.Context {
	return context.WithValue(parent, ctxCookieStore{}, store)
}
