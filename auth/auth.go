package auth

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"github.com/omecodes/libome/logs"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/omecodes/errors"
	ome "github.com/omecodes/libome"
)

type User struct {
	Name   string `json:"name,omitempty"`
	Access string `json:"access,omitempty"`
	Group  string `json:"group,omitempty"`
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
			return ctx, errors.Create(errors.BadRequest, "malformed authorization value")
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
			return ctx, errors.Create(errors.BadRequest, "malformed authorization value")
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
		return ctx, errors.Create(errors.BadRequest, "authorization wrong encoding")
	}

	parts := strings.Split(string(bytes), ":")
	if len(parts) != 2 {
		return ctx, errors.Create(errors.BadRequest, "wrong basic authentication")
	}

	authUser := parts[0]
	if authUser != "admin" {
		return nil, errors.Create(errors.Forbidden, "forbidden")
	}

	var pass string
	if len(parts) > 1 {
		pass = parts[1]
	}

	manager := GetCredentialsManager(ctx)
	if manager == nil {
		return ctx, errors.Create(errors.Forbidden, "No manager basic authentication is not supported")
	}

	err = manager.ValidateAdminAccess(pass)
	if err != nil {
		logs.Error("verifying admin authentication", logs.Err(err))
		return ctx, errors.Create(errors.Forbidden, "admin authentication failed")
	}

	return context.WithValue(ctx, ctxUser{}, &User{
		Name: "admin",
	}), nil
}

func updateContextWithProxyBasic(ctx context.Context, authorization string) (context.Context, error) {
	bytes, err := base64.StdEncoding.DecodeString(authorization)
	if err != nil {
		return ctx, errors.Create(errors.BadRequest, "authorization value non base64 encoding")
	}

	parts := strings.Split(string(bytes), ":")
	if len(parts) != 2 {
		return ctx, errors.Create(errors.BadRequest, "authorization value non base64 encoding")
	}

	authUser := parts[0]
	var pass string
	if len(parts) > 1 {
		pass = parts[1]
	}

	manager := GetCredentialsManager(ctx)
	if manager == nil {
		return ctx, errors.Create(errors.Forbidden, "No manager basic authentication is not supported")
	}

	sh := sha512.New()
	_, err = sh.Write([]byte(pass))
	if err != nil {
		return ctx, errors.Create(errors.Internal, "password hashing failed")
	}
	hashed := sh.Sum(nil)

	access, err := manager.GetAccess(authUser)
	if err != nil {
		return ctx, err
	}

	if access.Secret != hex.EncodeToString(hashed) {
		return ctx, errors.Create(errors.Forbidden, "authorization value non base64 encoding")
	}

	return context.WithValue(ctx, ctxUser{}, &User{
		Access: access.Type,
	}), nil
}

func updateContextWithOauth2(ctx context.Context, authorization string) (context.Context, error) {
	jwt, err := ome.ParseJWT(authorization)
	if err != nil {
		return ctx, nil
	}

	providers := GetProviders(ctx)
	if providers == nil {
		return ctx, errors.Create(errors.Forbidden, "token not signed")
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
		return ctx, errors.Create(errors.Forbidden, "token not signed")
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
