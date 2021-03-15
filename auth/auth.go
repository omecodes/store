package auth

import (
	"context"
	"encoding/base64"
	"github.com/omecodes/libome/logs"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/omecodes/errors"
	ome "github.com/omecodes/libome"
)

type InitClientAppSessionRequest struct {
	ClientApp *ClientApp `json:"client,omitempty"`
}

func BasicContextUpdater(ctx context.Context) (context.Context, error) {
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
			return ctx, errors.BadRequest("malformed authorization value")
		}
		return updateContextWithBasic(ctx, authorization)
	}
	return ctx, nil
}

func OAuth2ContextUpdater(ctx context.Context) (context.Context, error) {
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
			return ctx, errors.BadRequest("malformed authorization value")
		}
		return updateContextWithOauth2(ctx, authorization)
	}
	return ctx, nil
}

func UpdateFromMeta(parent context.Context) (context.Context, error) {
	a := FindInMD(parent)
	if a != nil {
		return context.WithValue(parent, ctxUser{}, a), nil
	}
	return parent, nil
}

func updateContextWithBasic(ctx context.Context, authorization string) (context.Context, error) {
	bytes, err := base64.StdEncoding.DecodeString(authorization)
	if err != nil {
		return ctx, errors.BadRequest("authorization wrong encoding")
	}

	parts := strings.Split(string(bytes), ":")
	if len(parts) != 2 {
		return ctx, errors.BadRequest("wrong basic authentication")
	}

	authUser := parts[0]
	if authUser == "admin" {

		var pass string
		if len(parts) > 1 {
			pass = parts[1]
		}

		manager := GetCredentialsManager(ctx)
		if manager == nil {
			return ctx, errors.Forbidden("No manager basic authentication is not supported")
		}

		err = manager.ValidateAdminAccess(pass)
		if err != nil {
			logs.Error("verifying admin authentication", logs.Err(err))
			return ctx, errors.Forbidden("admin authentication failed")
		}

		return context.WithValue(ctx, ctxUser{}, &User{
			Name: "admin",
		}), nil

	} else {
		var pass string
		if len(parts) > 1 {
			pass = parts[1]
		}

		manager := GetCredentialsManager(ctx)
		if manager == nil {
			return ctx, errors.Forbidden("No manager basic authentication is not supported")
		}

		clientApp, err := manager.GetClientApp(authUser)
		if err != nil {
			logs.Error("client access not found", logs.Details("access", authUser), logs.Err(err))
			return ctx, errors.Forbidden("client access not found")
		}

		if clientApp.Secret == pass {
			return context.WithValue(ctx, ctxUser{}, &User{
				Name: "admin",
			}), nil
		}

		return nil, errors.Forbidden("authentication failed")
	}
}

func updateContextWithClientAppInfo(ctx context.Context, authorization string) (context.Context, error) {
	bytes, err := base64.StdEncoding.DecodeString(authorization)
	if err != nil {
		return ctx, errors.BadRequest("authorization value non base64 encoding")
	}

	parts := strings.Split(string(bytes), ":")
	if len(parts) != 2 {
		return ctx, errors.BadRequest("authorization value non base64 encoding")
	}

	authUser := parts[0]
	var pass string
	if len(parts) > 1 {
		pass = parts[1]
	}

	manager := GetCredentialsManager(ctx)
	if manager == nil {
		return ctx, errors.Forbidden("No manager basic authentication is not supported")
	}

	clientApp, err := manager.GetClientApp(authUser)
	if err != nil {
		return ctx, err
	}

	if clientApp.Secret != pass {
		return ctx, errors.Forbidden("authorization value non base64 encoding")
	}

	return context.WithValue(ctx, ctxApp{}, access), nil
}

func updateContextWithOauth2(ctx context.Context, authorization string) (context.Context, error) {
	jwt, err := ome.ParseJWT(authorization)
	if err != nil {
		return ctx, nil
	}

	providers := GetProviders(ctx)
	if providers == nil {
		return ctx, errors.Forbidden("token not signed")
	}

	provider, err := providers.Get(jwt.Claims.Iss)
	if err != nil {
		return ctx, err
	}

	signature, err := jwt.SecretBasedSignature(provider.Config.ClientSecret)
	if err != nil {
		return ctx, err
	}

	if signature != jwt.Signature {
		return ctx, errors.Forbidden("token not signed")
	}

	ctx = context.WithValue(ctx, ctxJWt{}, jwt)
	o := ctx.Value(ctxUser{})
	if o != nil {
		user := o.(*User)
		user.Name = jwt.Claims.Sub
		return ctx, nil

	} else {
		return context.WithValue(ctx, ctxUser{}, &User{
			Name: jwt.Claims.Sub,
		}), nil
	}

}
