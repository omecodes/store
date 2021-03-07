package session

import (
	"encoding/json"
	"github.com/omecodes/errors"
	"net/http"

	"github.com/gorilla/sessions"
)

type Cookie struct {
	http.Cookie
}

type WebSession struct {
	store       *sessions.CookieStore
	httpSession *sessions.Session
	r           *http.Request
}

func GetWebSession(name string, r *http.Request) (*WebSession, error) {
	s := new(WebSession)
	storeValue := r.Context().Value(ctxCookieStore{})
	if storeValue == nil {
		return nil, errors.Internal("context missing cookies store")
	}

	s.store = storeValue.(*sessions.CookieStore)
	s.r = r
	s.httpSession, _ = s.store.Get(r, name)
	return s, nil
}

func (s *WebSession) Put(key string, value interface{}) {
	s.httpSession.Values[key] = value
}

func (s *WebSession) Get(key string) interface{} {
	v, ok := s.httpSession.Values[key]
	if !ok {
		return nil
	}
	return v
}

func (s *WebSession) Delete(key string) {
	delete(s.httpSession.Values, key)
}

func (s *WebSession) String(key string) string {
	v, ok := s.httpSession.Values[key]
	if !ok {
		return ""
	}
	str, ok := v.(string)
	if !ok {
		return ""
	}
	return str
}

func (s *WebSession) Bool(key string) bool {
	v, ok := s.httpSession.Values[key]
	if !ok {
		return ok
	}

	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}

func (s *WebSession) Int64(key string) int64 {
	v, ok := s.httpSession.Values[key]
	if !ok {
		return 0
	}

	b, ok := v.(int64)
	if !ok {
		return 0
	}
	return b
}

func (s *WebSession) Save(w http.ResponseWriter) error {
	return s.httpSession.Save(s.r, w)
}

func (s *WebSession) Encoded() ([]byte, error) {
	return json.Marshal(s.httpSession.Values)
}
