package auth

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"

	ome "github.com/omecodes/libome"
	"github.com/omecodes/libome/errors"
	"github.com/omecodes/store/pb"
)

type ctxAuthentication struct{}

func Get(ctx context.Context) *pb.Auth {
	o := ctx.Value(ctxAuthentication{})
	if o == nil {
		return nil
	}
	return o.(*pb.Auth)
}

func Context(parent context.Context, a *pb.Auth) context.Context {
	return context.WithValue(parent, ctxAuthentication{}, a)
}

func BasicContextUpdater(manager CredentialsManager) ome.GrpcContextUpdaterFunc {
	return func(ctx context.Context) (context.Context, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return ctx, nil
		}

		authorizationParts := strings.SplitN(md.Get("authorization")[0], " ", 2)
		authType := strings.ToLower(authorizationParts[0])
		var authorization string
		if len(authorizationParts) > 1 {
			authorization = authorizationParts[1]
		}

		if authType == "basic" {
			if authorization == "" {
				return ctx, errors.New(errors.CodeBadRequest, "malformed authorization value")
			}
			return updateContextWithBasic(ctx, manager, authorization)
		}
		return ctx, nil
	}
}

func OAuth2ContextUpdater(secret string) ome.GrpcContextUpdaterFunc {
	return func(ctx context.Context) (context.Context, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return ctx, nil
		}

		authorizationParts := strings.SplitN(md.Get("authorization")[0], " ", 2)
		authType := strings.ToLower(authorizationParts[0])
		var authorization string
		if len(authorizationParts) > 1 {
			authorization = authorizationParts[1]
		}

		if authType == "bearer" {
			if authorization == "" {
				return ctx, errors.New(errors.CodeBadRequest, "malformed authorization value")
			}

			return updateContextWithOauth2(ctx, secret, authorization)
		}
		return ctx, nil
	}
}

func UpdateFromMeta(parent context.Context) (context.Context, error) {
	a := FindInMD(parent)
	if a != nil {
		return context.WithValue(parent, ctxAuthentication{}, a), nil
	}
	return parent, nil
}

func DetectBasicMiddleware(manager CredentialsManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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

					ctx, err := updateContextWithBasic(r.Context(), manager, authorization)
					if err != nil {
						if err2, ok := err.(*errors.Error); ok {
							w.WriteHeader(err2.Code)
							return
						}
						w.WriteHeader(http.StatusForbidden)
						return
					}
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func DetectOauth2Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("Authorization")

			if authorization != "" {
				authorizationParts := strings.SplitN(authorization, " ", 2)
				authType := strings.ToLower(authorizationParts[0])
				if len(authorizationParts) > 1 {
					authorization = authorizationParts[1]
				}

				if authType == "bearer" {
					if authorization == "" {
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					ctx, err := updateContextWithOauth2(r.Context(), secret, authorization)
					if err != nil {
						if err2, ok := err.(*errors.Error); ok {
							w.WriteHeader(err2.Code)
							return
						}
						w.WriteHeader(http.StatusForbidden)
						return
					}
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func updateContextWithBasic(ctx context.Context, manager CredentialsManager, authorization string) (context.Context, error) {
	bytes, err := base64.StdEncoding.DecodeString(authorization)
	if err != nil {
		return ctx, errors.New(errors.CodeBadRequest, "authorization value non base64 encoding")
	}

	parts := strings.Split(string(bytes), ":")
	if len(parts) != 2 {
		return ctx, errors.New(errors.CodeBadRequest, "authorization value non base64 encoding")
	}

	authUser := parts[0]
	var pass string
	if len(parts) > 1 {
		pass = parts[1]
	}

	sh := sha512.New()
	_, err = sh.Write([]byte(pass))
	if err != nil {
		return ctx, errors.New(errors.CodeInternal, "password hashing failed")
	}
	hashed := sh.Sum(nil)

	password, err := manager.Get(authUser)
	if err != nil {
		return ctx, err
	}

	if password != hex.EncodeToString(hashed) {
		return ctx, errors.New(errors.CodeForbidden, "authorization value non base64 encoding")
	}

	return context.WithValue(ctx, ctxAuthentication{}, &pb.Auth{
		Uid:    authUser,
		Worker: "admin" != authUser,
	}), nil
}

func updateContextWithOauth2(ctx context.Context, secret string, authorization string) (context.Context, error) {
	jwt, err := ome.ParseJWT(authorization)
	if err != nil {
		return ctx, nil
	}

	signature, err := jwt.SecretBasedSignature(secret)
	if err != nil {
		return ctx, err
	}

	if signature != jwt.Signature {
		return ctx, errors.New(errors.CodeForbidden, "token not signed")
	}

	return context.WithValue(ctx, ctxAuthentication{}, &pb.Auth{
		Uid:    jwt.Claims.Sub,
		Email:  jwt.Claims.Profile.Email,
		Worker: false,
		Scope:  strings.Split(jwt.Claims.Scope, ""),
	}), nil
}
