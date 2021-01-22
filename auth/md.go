package auth

import (
	"context"
	"google.golang.org/grpc/metadata"
)

const (
	uuid   = "user.name"
	access = "user.access"
	group  = "auth.group"
)

func SetMetaWithExisting(ctx context.Context) context.Context {
	a := Get(ctx)
	if a != nil {
		md := ToMD(a)
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

func FindInMD(ctx context.Context) *User {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	return FromMD(md)
}

func ToMD(a *User) metadata.MD {
	md := metadata.MD{}
	md.Set(uuid, a.Name)

	if a.Access != "" {
		md.Set(access, a.Access)
	}

	if a.Group != "" {
		md.Set(group, a.Group)
	}
	return md
}

func FromMD(md metadata.MD) *User {
	userValues := md.Get(uuid)
	if len(userValues) == 0 {
		return nil
	}

	a := &User{}
	a.Name = userValues[0]

	emailValues := md.Get(access)
	if len(emailValues) > 0 {
		a.Access = emailValues[0]
	}

	groupValues := md.Get(group)
	if len(groupValues) > 0 {
		a.Group = groupValues[0]
	}

	return a
}
