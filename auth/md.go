package auth

import (
	"context"
	"github.com/omecodes/store/pb"
	"google.golang.org/grpc/metadata"
)

const (
	uuid   = "auth.user"
	email  = "auth.user"
	worker = "auth.is_worker"
	scope  = "auth.scope"
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

func FindInMD(ctx context.Context) *pb.Auth {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	return FromMD(md)
}

func ToMD(a *pb.Auth) metadata.MD {
	md := metadata.MD{}
	md.Set(uuid, a.Uid)

	if a.Email != "" {
		md.Set(email, a.Email)
	}

	if a.Worker {
		md.Set(worker, "true")
	} else {
		md.Set(worker, "false")
	}

	if len(a.Scope) > 0 {
		md.Set(scope, a.Scope...)
	}

	if a.Group != "" {
		md.Set(group, a.Group)
	}
	return md
}

func FromMD(md metadata.MD) *pb.Auth {
	userValues := md.Get(uuid)
	if len(userValues) == 0 {
		return nil
	}

	a := &pb.Auth{}
	a.Uid = userValues[0]

	emailValues := md.Get(email)
	if len(emailValues) > 0 {
		a.Email = emailValues[0]
	}

	workerValues := md.Get(worker)
	if len(workerValues) > 0 {
		a.Worker = workerValues[0] == "true"
	}

	a.Scope = md.Get(scope)

	groupValues := md.Get(group)
	if len(groupValues) > 0 {
		a.Group = groupValues[0]
	}

	return a
}
