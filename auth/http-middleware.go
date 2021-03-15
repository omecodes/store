package auth

import (
	"context"
	"encoding/json"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/session"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/omecodes/errors"
)

type middlewareOptions struct {
	credentials CredentialsManager
	providers   ProviderManager
}

type MiddlewareOption func(*middlewareOptions)

func MiddlewareWithProviderManager(manager ProviderManager) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.providers = manager
	}
}

func MiddlewareWithCredentials(manager CredentialsManager) MiddlewareOption {
	return func(options *middlewareOptions) {
		options.credentials = manager
	}
}

func Middleware(opts ...MiddlewareOption) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		next = detectUserAuthentication(next)
		next = detectClientAppAuthentication(next)

		var options middlewareOptions
		for _, opt := range opts {
			opt(&options)
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			updatedContext := r.Context()
			if options.credentials != nil {
				updatedContext = context.WithValue(updatedContext, ctxCredentialsManager{}, options.credentials)
			}

			if options.providers != nil {
				updatedContext = context.WithValue(updatedContext, ctxAuthenticationProviders{}, options.providers)
			}

			user := Get(updatedContext)
			if user == nil {
				user = &User{
					Name:  "",
					Group: "",
				}
				updatedContext = ContextWithUser(updatedContext, user)
			}

			next.ServeHTTP(w, r.WithContext(updatedContext))
		})
	}
}

func detectUserAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		if authorization != "" {
			authorizationParts := strings.SplitN(authorization, " ", 2)
			authType := strings.ToLower(authorizationParts[0])
			if len(authorizationParts) > 1 {
				authorization = authorizationParts[1]
			}

			if authType == "basic" {
				if authorization == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				ctx := r.Context()
				ctx, err := updateContextWithBasic(ctx, authorization)
				if err != nil {
					w.WriteHeader(errors.HTTPStatus(err))
					return
				}
				r = r.WithContext(ctx)
			} else if authType == "bearer" {
				if authorization == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				ctx, err := updateContextWithOauth2(r.Context(), authorization)
				if err != nil {
					w.WriteHeader(errors.HTTPStatus(err))
					return
				}
				r = r.WithContext(ctx)
			}

		} else {
			userSession, err := session.GetWebSession(session.UserSession, r)
			if err != nil {
				logs.Error("could not get web session", logs.Err(err))
				w.WriteHeader(errors.HTTPStatus(err))
				return
			}

			if username := userSession.String(session.KeyUsername); username != "" {
				logs.Info("detected user authentication", logs.Details("user", username))
				ctx := context.WithValue(r.Context(), ctxUser{}, &User{
					Name: username,
				})
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func detectClientAppAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("X-STORE-API-Authorization")

		if authorization != "" {
			authorizationParts := strings.SplitN(authorization, " ", 2)
			authType := strings.ToLower(authorizationParts[0])
			if len(authorizationParts) > 1 {
				authorization = authorizationParts[1]
			}

			if authType == "basic" {
				if authorization == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				ctx := r.Context()
				ctx, err := updateContextWithClientAppInfo(ctx, authorization)
				if err != nil {
					w.WriteHeader(errors.HTTPStatus(err))
					return
				}
				r = r.WithContext(ctx)
			}
		} else {
			webSession, err := session.GetWebSession(session.ClientAppSession, r)
			if err != nil {
				logs.Error("could not get web session", logs.Err(err))
				w.WriteHeader(errors.HTTPStatus(err))
				return
			}

			if accessType := webSession.String(session.KeyAccessType); accessType != "" {
				logs.Info("detected client app session")

				clientApp := &ClientApp{
					Key:  webSession.String(session.KeyAccessKey),
					Type: ClientType(ClientType_value[webSession.String(session.KeyAccessType)]),
				}

				infoInterface := webSession.String(session.KeyAccessInfo)
				if infoInterface != "" {
					err = json.Unmarshal([]byte(infoInterface), &clientApp.Info)
					if err != nil {
						logs.Error("could not decode client app info", logs.Err(err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}

				ctx := context.WithValue(r.Context(), ctxApp{}, clientApp)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}
