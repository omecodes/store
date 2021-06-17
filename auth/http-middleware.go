package auth

import (
	"context"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/store/common"
	pb "github.com/omecodes/store/gen/go/proto"
	"github.com/omecodes/store/session"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/omecodes/errors"
)

const (
	UserHeader = "X-User"
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
	var options middlewareOptions
	for _, opt := range opts {
		opt(&options)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			updatedContext := r.Context()

			updatedContext = context.WithValue(updatedContext, ctxCredentialsManager{}, options.credentials)
			updatedContext = context.WithValue(updatedContext, ctxAuthenticationProviders{}, options.providers)

			var err error
			updatedContext, err = userContext(r.WithContext(updatedContext))
			if err != nil {
				w.WriteHeader(errors.HTTPStatus(err))
				return
			}

			updatedContext, err = clientAppContext(r.WithContext(updatedContext))
			if err != nil {
				w.WriteHeader(errors.HTTPStatus(err))
				return
			}
			next.ServeHTTP(w, r.WithContext(updatedContext))
		})
	}
}

func userContext(r *http.Request) (context.Context, error) {
	ctx := r.Context()

	authorization := r.Header.Get(common.HttpHeaderUserAuthorization)
	if authorization != "" {
		authorizationParts := strings.SplitN(authorization, " ", 2)
		authType := strings.ToLower(authorizationParts[0])
		if len(authorizationParts) > 1 {
			authorization = authorizationParts[1]
		}

		if authType == "basic" {
			if authorization == "" {
				return nil, errors.Forbidden("malformed authentication")
			}
			return updateContextWithBasic(ctx, authorization)

		} else if authType == "bearer" {
			if authorization == "" {
				return nil, errors.Forbidden("malformed authentication")
			}
			return updateContextWithOauth2(r.Context(), authorization)
		}

	} else {
		userSession, err := session.GetWebSession(session.UserSession, r)
		if err != nil {
			logs.Error("could not get web session", logs.Err(err))
			return nil, err
		}

		if username := userSession.String(session.KeyUsername); username != "" {
			logs.Info("detected user authentication", logs.Details("user", username))
			return context.WithValue(r.Context(), ctxUser{}, &pb.User{
				Name: username,
			}), nil
		}
	}
	return ctx, nil
}

func clientAppContext(r *http.Request) (context.Context, error) {
	ctx := r.Context()

	authorization := r.Header.Get(common.HttpHeaderAppAuthorization)
	if authorization != "" {
		authorizationParts := strings.SplitN(authorization, " ", 2)
		authType := strings.ToLower(authorizationParts[0])
		if len(authorizationParts) > 1 {
			authorization = authorizationParts[1]
		}
		if authType == "basic" {
			if authorization == "" {
				return ctx, errors.Forbidden("malformed authentication")
			}
			return updateContextWithClientAppInfo(ctx, authorization)
		}

	} else {
		webSession, err := session.GetWebSession(session.ClientAppSession, r)
		if err != nil {
			return ctx, err
		}

		if accessType := webSession.String(session.KeyAccessType); accessType != "" {
			logs.Info("detected client app session")

			clientApp := &pb.ClientApp{
				Key:  webSession.String(session.KeyAccessKey),
				Type: pb.ClientType(pb.ClientType_value[webSession.String(session.KeyAccessType)]),
			}

			infoInterface := webSession.String(session.KeyAccessInfo)
			if infoInterface != "" {
				err = json.Unmarshal([]byte(infoInterface), &clientApp.Info)
				if err != nil {
					logs.Error("could not decode client app info", logs.Err(err))
					return ctx, err
				}
			}
			return context.WithValue(r.Context(), ctxApp{}, clientApp), nil
		}
	}
	return ctx, nil
}

func ServiceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		updatedContext := r.Context()
		encoded := r.Header.Get(UserHeader)
		if encoded != "" {
			user := &pb.User{}
			err := jsonpb.UnmarshalString(encoded, user)
			if err != nil {
				logs.Error("could not decode user from custom header", logs.Err(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			updatedContext = ContextWithUser(updatedContext, user)
		}

		encoded = r.Header.Get(UserHeader)
		if encoded != "" {
			app := &pb.ClientApp{}
			err := jsonpb.UnmarshalString(encoded, app)
			if err != nil {
				logs.Error("could not decode client app from custom header", logs.Err(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			updatedContext = ContextWithApp(updatedContext, app)
		}

		next.ServeHTTP(w, r.WithContext(updatedContext))
	})
}
