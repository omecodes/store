package auth

import (
	"context"
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc/metadata"
)

const (
	mdUser = "auth.user"
	mdApp  = "auth.app"
)

func ContextWithMeta(parent context.Context) (context.Context, error) {
	md := metadata.MD{}
	marshaler := &jsonpb.Marshaler{EnumsAsInts: true}

	user := Get(parent)
	if user != nil {
		encoded, err := marshaler.MarshalToString(user)
		if err != nil {
			return nil, err
		}
		md.Set(mdUser, encoded)
	}

	client := App(parent)
	if client != nil {
		encoded, err := marshaler.MarshalToString(user)
		if err != nil {
			return nil, err
		}
		md.Set(mdApp, encoded)
	}
	return metadata.NewOutgoingContext(parent, md), nil
}

func ParseMetaInNewContext(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		user := &User{}
		userValues := md.Get(mdUser)
		err := jsonpb.UnmarshalString(userValues[0], user)
		if err != nil {
			return nil, err
		}

		clientApp := &ClientApp{}
		userValues = md.Get(mdApp)
		err = jsonpb.UnmarshalString(userValues[0], clientApp)
		if err != nil {
			return nil, err
		}

		newCtx := context.WithValue(ctx, ctxUser{}, user)
		return context.WithValue(newCtx, ctxApp{}, clientApp), nil
	}
	return ctx, nil
}
