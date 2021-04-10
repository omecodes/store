package client

import (
	"crypto/tls"
	"net/http"
)

type options struct {
	userAuth    Authentication
	appAuth     Authentication
	apiLocation string
	noTLS       bool
	headers     http.Header
	port        int
	tlsConfig   *tls.Config
}

type Option func(*options)

func WithAppAuthentication(key, secret string) Option {
	return func(o *options) {}
}

func WithUserBasicAuthentication(username, password string) Option {
	return func(o *options) {
		o.userAuth = &userBasicAuthentication{username: username, password: password}
	}
}

func WithUserBearerTokenAuthentication(token string) Option {
	return func(o *options) {
		o.userAuth = &bearerTokenAuthentication{token: token}
	}
}

func WithAPILocation(location string) Option {
	return func(o *options) {
		o.apiLocation = location
	}
}

func WithoutTLS() Option {
	return func(o *options) {
		o.noTLS = true
	}
}

func WithTLS(tc *tls.Config) Option {
	return func(o *options) {
		o.tlsConfig = tc
	}
}

func WithHeader(name, value string) Option {
	return func(o *options) {
		if o.headers == nil {
			o.headers = http.Header{}
		}
		o.headers.Set(name, value)
	}
}
