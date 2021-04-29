// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// ObjectsClient is the client API for Objects service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ObjectsClient interface {
	CreateCollection(ctx context.Context, in *CreateCollectionRequest, opts ...grpc.CallOption) (*CreateCollectionResponse, error)
	GetCollection(ctx context.Context, in *GetCollectionRequest, opts ...grpc.CallOption) (*GetCollectionResponse, error)
	ListCollections(ctx context.Context, in *ListCollectionsRequest, opts ...grpc.CallOption) (*ListCollectionsResponse, error)
	DeleteCollection(ctx context.Context, in *DeleteCollectionRequest, opts ...grpc.CallOption) (*DeleteCollectionResponse, error)
	PutObject(ctx context.Context, in *PutObjectRequest, opts ...grpc.CallOption) (*PutObjectResponse, error)
	PatchObject(ctx context.Context, in *PatchObjectRequest, opts ...grpc.CallOption) (*PatchObjectResponse, error)
	MoveObject(ctx context.Context, in *MoveObjectRequest, opts ...grpc.CallOption) (*MoveObjectResponse, error)
	GetObject(ctx context.Context, in *GetObjectRequest, opts ...grpc.CallOption) (*GetObjectResponse, error)
	DeleteObject(ctx context.Context, in *DeleteObjectRequest, opts ...grpc.CallOption) (*DeleteObjectResponse, error)
	ObjectInfo(ctx context.Context, in *ObjectInfoRequest, opts ...grpc.CallOption) (*ObjectInfoResponse, error)
	ListObjects(ctx context.Context, in *ListObjectsRequest, opts ...grpc.CallOption) (Objects_ListObjectsClient, error)
	SearchObjects(ctx context.Context, in *SearchObjectsRequest, opts ...grpc.CallOption) (Objects_SearchObjectsClient, error)
}

type objectsClient struct {
	cc grpc.ClientConnInterface
}

func NewObjectsClient(cc grpc.ClientConnInterface) ObjectsClient {
	return &objectsClient{cc}
}

func (c *objectsClient) CreateCollection(ctx context.Context, in *CreateCollectionRequest, opts ...grpc.CallOption) (*CreateCollectionResponse, error) {
	out := new(CreateCollectionResponse)
	err := c.cc.Invoke(ctx, "/Objects/CreateCollection", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectsClient) GetCollection(ctx context.Context, in *GetCollectionRequest, opts ...grpc.CallOption) (*GetCollectionResponse, error) {
	out := new(GetCollectionResponse)
	err := c.cc.Invoke(ctx, "/Objects/GetCollection", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectsClient) ListCollections(ctx context.Context, in *ListCollectionsRequest, opts ...grpc.CallOption) (*ListCollectionsResponse, error) {
	out := new(ListCollectionsResponse)
	err := c.cc.Invoke(ctx, "/Objects/ListCollections", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectsClient) DeleteCollection(ctx context.Context, in *DeleteCollectionRequest, opts ...grpc.CallOption) (*DeleteCollectionResponse, error) {
	out := new(DeleteCollectionResponse)
	err := c.cc.Invoke(ctx, "/Objects/DeleteCollection", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectsClient) PutObject(ctx context.Context, in *PutObjectRequest, opts ...grpc.CallOption) (*PutObjectResponse, error) {
	out := new(PutObjectResponse)
	err := c.cc.Invoke(ctx, "/Objects/PutObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectsClient) PatchObject(ctx context.Context, in *PatchObjectRequest, opts ...grpc.CallOption) (*PatchObjectResponse, error) {
	out := new(PatchObjectResponse)
	err := c.cc.Invoke(ctx, "/Objects/PatchObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectsClient) MoveObject(ctx context.Context, in *MoveObjectRequest, opts ...grpc.CallOption) (*MoveObjectResponse, error) {
	out := new(MoveObjectResponse)
	err := c.cc.Invoke(ctx, "/Objects/MoveObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectsClient) GetObject(ctx context.Context, in *GetObjectRequest, opts ...grpc.CallOption) (*GetObjectResponse, error) {
	out := new(GetObjectResponse)
	err := c.cc.Invoke(ctx, "/Objects/GetObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectsClient) DeleteObject(ctx context.Context, in *DeleteObjectRequest, opts ...grpc.CallOption) (*DeleteObjectResponse, error) {
	out := new(DeleteObjectResponse)
	err := c.cc.Invoke(ctx, "/Objects/DeleteObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectsClient) ObjectInfo(ctx context.Context, in *ObjectInfoRequest, opts ...grpc.CallOption) (*ObjectInfoResponse, error) {
	out := new(ObjectInfoResponse)
	err := c.cc.Invoke(ctx, "/Objects/ObjectInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectsClient) ListObjects(ctx context.Context, in *ListObjectsRequest, opts ...grpc.CallOption) (Objects_ListObjectsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Objects_serviceDesc.Streams[0], "/Objects/ListObjects", opts...)
	if err != nil {
		return nil, err
	}
	x := &objectsListObjectsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Objects_ListObjectsClient interface {
	Recv() (*Object, error)
	grpc.ClientStream
}

type objectsListObjectsClient struct {
	grpc.ClientStream
}

func (x *objectsListObjectsClient) Recv() (*Object, error) {
	m := new(Object)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *objectsClient) SearchObjects(ctx context.Context, in *SearchObjectsRequest, opts ...grpc.CallOption) (Objects_SearchObjectsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Objects_serviceDesc.Streams[1], "/Objects/SearchObjects", opts...)
	if err != nil {
		return nil, err
	}
	x := &objectsSearchObjectsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Objects_SearchObjectsClient interface {
	Recv() (*Object, error)
	grpc.ClientStream
}

type objectsSearchObjectsClient struct {
	grpc.ClientStream
}

func (x *objectsSearchObjectsClient) Recv() (*Object, error) {
	m := new(Object)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ObjectsServer is the server API for Objects service.
// All implementations must embed UnimplementedObjectsServer
// for forward compatibility
type ObjectsServer interface {
	CreateCollection(context.Context, *CreateCollectionRequest) (*CreateCollectionResponse, error)
	GetCollection(context.Context, *GetCollectionRequest) (*GetCollectionResponse, error)
	ListCollections(context.Context, *ListCollectionsRequest) (*ListCollectionsResponse, error)
	DeleteCollection(context.Context, *DeleteCollectionRequest) (*DeleteCollectionResponse, error)
	PutObject(context.Context, *PutObjectRequest) (*PutObjectResponse, error)
	PatchObject(context.Context, *PatchObjectRequest) (*PatchObjectResponse, error)
	MoveObject(context.Context, *MoveObjectRequest) (*MoveObjectResponse, error)
	GetObject(context.Context, *GetObjectRequest) (*GetObjectResponse, error)
	DeleteObject(context.Context, *DeleteObjectRequest) (*DeleteObjectResponse, error)
	ObjectInfo(context.Context, *ObjectInfoRequest) (*ObjectInfoResponse, error)
	ListObjects(*ListObjectsRequest, Objects_ListObjectsServer) error
	SearchObjects(*SearchObjectsRequest, Objects_SearchObjectsServer) error
	mustEmbedUnimplementedObjectsServer()
}

// UnimplementedObjectsServer must be embedded to have forward compatible implementations.
type UnimplementedObjectsServer struct {
}

func (UnimplementedObjectsServer) CreateCollection(context.Context, *CreateCollectionRequest) (*CreateCollectionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateCollection not implemented")
}
func (UnimplementedObjectsServer) GetCollection(context.Context, *GetCollectionRequest) (*GetCollectionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCollection not implemented")
}
func (UnimplementedObjectsServer) ListCollections(context.Context, *ListCollectionsRequest) (*ListCollectionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListCollections not implemented")
}
func (UnimplementedObjectsServer) DeleteCollection(context.Context, *DeleteCollectionRequest) (*DeleteCollectionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteCollection not implemented")
}
func (UnimplementedObjectsServer) PutObject(context.Context, *PutObjectRequest) (*PutObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutObject not implemented")
}
func (UnimplementedObjectsServer) PatchObject(context.Context, *PatchObjectRequest) (*PatchObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PatchObject not implemented")
}
func (UnimplementedObjectsServer) MoveObject(context.Context, *MoveObjectRequest) (*MoveObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MoveObject not implemented")
}
func (UnimplementedObjectsServer) GetObject(context.Context, *GetObjectRequest) (*GetObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetObject not implemented")
}
func (UnimplementedObjectsServer) DeleteObject(context.Context, *DeleteObjectRequest) (*DeleteObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteObject not implemented")
}
func (UnimplementedObjectsServer) ObjectInfo(context.Context, *ObjectInfoRequest) (*ObjectInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ObjectInfo not implemented")
}
func (UnimplementedObjectsServer) ListObjects(*ListObjectsRequest, Objects_ListObjectsServer) error {
	return status.Errorf(codes.Unimplemented, "method ListObjects not implemented")
}
func (UnimplementedObjectsServer) SearchObjects(*SearchObjectsRequest, Objects_SearchObjectsServer) error {
	return status.Errorf(codes.Unimplemented, "method SearchObjects not implemented")
}
func (UnimplementedObjectsServer) mustEmbedUnimplementedObjectsServer() {}

// UnsafeObjectsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ObjectsServer will
// result in compilation errors.
type UnsafeObjectsServer interface {
	mustEmbedUnimplementedObjectsServer()
}

func RegisterObjectsServer(s *grpc.Server, srv ObjectsServer) {
	s.RegisterService(&_Objects_serviceDesc, srv)
}

func _Objects_CreateCollection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateCollectionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectsServer).CreateCollection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Objects/CreateCollection",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectsServer).CreateCollection(ctx, req.(*CreateCollectionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Objects_GetCollection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCollectionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectsServer).GetCollection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Objects/GetCollection",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectsServer).GetCollection(ctx, req.(*GetCollectionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Objects_ListCollections_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListCollectionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectsServer).ListCollections(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Objects/ListCollections",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectsServer).ListCollections(ctx, req.(*ListCollectionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Objects_DeleteCollection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteCollectionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectsServer).DeleteCollection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Objects/DeleteCollection",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectsServer).DeleteCollection(ctx, req.(*DeleteCollectionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Objects_PutObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectsServer).PutObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Objects/PutObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectsServer).PutObject(ctx, req.(*PutObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Objects_PatchObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PatchObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectsServer).PatchObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Objects/PatchObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectsServer).PatchObject(ctx, req.(*PatchObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Objects_MoveObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MoveObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectsServer).MoveObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Objects/MoveObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectsServer).MoveObject(ctx, req.(*MoveObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Objects_GetObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectsServer).GetObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Objects/GetObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectsServer).GetObject(ctx, req.(*GetObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Objects_DeleteObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectsServer).DeleteObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Objects/DeleteObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectsServer).DeleteObject(ctx, req.(*DeleteObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Objects_ObjectInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ObjectInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectsServer).ObjectInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Objects/ObjectInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectsServer).ObjectInfo(ctx, req.(*ObjectInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Objects_ListObjects_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListObjectsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ObjectsServer).ListObjects(m, &objectsListObjectsServer{stream})
}

type Objects_ListObjectsServer interface {
	Send(*Object) error
	grpc.ServerStream
}

type objectsListObjectsServer struct {
	grpc.ServerStream
}

func (x *objectsListObjectsServer) Send(m *Object) error {
	return x.ServerStream.SendMsg(m)
}

func _Objects_SearchObjects_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SearchObjectsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ObjectsServer).SearchObjects(m, &objectsSearchObjectsServer{stream})
}

type Objects_SearchObjectsServer interface {
	Send(*Object) error
	grpc.ServerStream
}

type objectsSearchObjectsServer struct {
	grpc.ServerStream
}

func (x *objectsSearchObjectsServer) Send(m *Object) error {
	return x.ServerStream.SendMsg(m)
}

var _Objects_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Objects",
	HandlerType: (*ObjectsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateCollection",
			Handler:    _Objects_CreateCollection_Handler,
		},
		{
			MethodName: "GetCollection",
			Handler:    _Objects_GetCollection_Handler,
		},
		{
			MethodName: "ListCollections",
			Handler:    _Objects_ListCollections_Handler,
		},
		{
			MethodName: "DeleteCollection",
			Handler:    _Objects_DeleteCollection_Handler,
		},
		{
			MethodName: "PutObject",
			Handler:    _Objects_PutObject_Handler,
		},
		{
			MethodName: "PatchObject",
			Handler:    _Objects_PatchObject_Handler,
		},
		{
			MethodName: "MoveObject",
			Handler:    _Objects_MoveObject_Handler,
		},
		{
			MethodName: "GetObject",
			Handler:    _Objects_GetObject_Handler,
		},
		{
			MethodName: "DeleteObject",
			Handler:    _Objects_DeleteObject_Handler,
		},
		{
			MethodName: "ObjectInfo",
			Handler:    _Objects_ObjectInfo_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ListObjects",
			Handler:       _Objects_ListObjects_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "SearchObjects",
			Handler:       _Objects_SearchObjects_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/objects.proto",
}

// AccessRuleStoreClient is the client API for AccessRuleStore service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AccessRuleStoreClient interface {
	PutRules(ctx context.Context, in *PutRulesRequest, opts ...grpc.CallOption) (*PutRulesResponse, error)
	GetRules(ctx context.Context, in *GetRulesRequest, opts ...grpc.CallOption) (*GetRulesResponse, error)
	GetRulesForPath(ctx context.Context, in *GetRulesForPathRequest, opts ...grpc.CallOption) (*GetRulesForPathResponse, error)
	DeleteRules(ctx context.Context, in *DeleteRulesRequest, opts ...grpc.CallOption) (*DeleteRulesResponse, error)
	DeleteRulesForPath(ctx context.Context, in *DeleteRulesForPathRequest, opts ...grpc.CallOption) (*DeleteRulesForPathResponse, error)
}

type accessRuleStoreClient struct {
	cc grpc.ClientConnInterface
}

func NewAccessRuleStoreClient(cc grpc.ClientConnInterface) AccessRuleStoreClient {
	return &accessRuleStoreClient{cc}
}

func (c *accessRuleStoreClient) PutRules(ctx context.Context, in *PutRulesRequest, opts ...grpc.CallOption) (*PutRulesResponse, error) {
	out := new(PutRulesResponse)
	err := c.cc.Invoke(ctx, "/AccessRuleStore/PutRules", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accessRuleStoreClient) GetRules(ctx context.Context, in *GetRulesRequest, opts ...grpc.CallOption) (*GetRulesResponse, error) {
	out := new(GetRulesResponse)
	err := c.cc.Invoke(ctx, "/AccessRuleStore/GetRules", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accessRuleStoreClient) GetRulesForPath(ctx context.Context, in *GetRulesForPathRequest, opts ...grpc.CallOption) (*GetRulesForPathResponse, error) {
	out := new(GetRulesForPathResponse)
	err := c.cc.Invoke(ctx, "/AccessRuleStore/GetRulesForPath", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accessRuleStoreClient) DeleteRules(ctx context.Context, in *DeleteRulesRequest, opts ...grpc.CallOption) (*DeleteRulesResponse, error) {
	out := new(DeleteRulesResponse)
	err := c.cc.Invoke(ctx, "/AccessRuleStore/DeleteRules", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accessRuleStoreClient) DeleteRulesForPath(ctx context.Context, in *DeleteRulesForPathRequest, opts ...grpc.CallOption) (*DeleteRulesForPathResponse, error) {
	out := new(DeleteRulesForPathResponse)
	err := c.cc.Invoke(ctx, "/AccessRuleStore/DeleteRulesForPath", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AccessRuleStoreServer is the server API for AccessRuleStore service.
// All implementations must embed UnimplementedAccessRuleStoreServer
// for forward compatibility
type AccessRuleStoreServer interface {
	PutRules(context.Context, *PutRulesRequest) (*PutRulesResponse, error)
	GetRules(context.Context, *GetRulesRequest) (*GetRulesResponse, error)
	GetRulesForPath(context.Context, *GetRulesForPathRequest) (*GetRulesForPathResponse, error)
	DeleteRules(context.Context, *DeleteRulesRequest) (*DeleteRulesResponse, error)
	DeleteRulesForPath(context.Context, *DeleteRulesForPathRequest) (*DeleteRulesForPathResponse, error)
	mustEmbedUnimplementedAccessRuleStoreServer()
}

// UnimplementedAccessRuleStoreServer must be embedded to have forward compatible implementations.
type UnimplementedAccessRuleStoreServer struct {
}

func (UnimplementedAccessRuleStoreServer) PutRules(context.Context, *PutRulesRequest) (*PutRulesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutRules not implemented")
}
func (UnimplementedAccessRuleStoreServer) GetRules(context.Context, *GetRulesRequest) (*GetRulesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRules not implemented")
}
func (UnimplementedAccessRuleStoreServer) GetRulesForPath(context.Context, *GetRulesForPathRequest) (*GetRulesForPathResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRulesForPath not implemented")
}
func (UnimplementedAccessRuleStoreServer) DeleteRules(context.Context, *DeleteRulesRequest) (*DeleteRulesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteRules not implemented")
}
func (UnimplementedAccessRuleStoreServer) DeleteRulesForPath(context.Context, *DeleteRulesForPathRequest) (*DeleteRulesForPathResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteRulesForPath not implemented")
}
func (UnimplementedAccessRuleStoreServer) mustEmbedUnimplementedAccessRuleStoreServer() {}

// UnsafeAccessRuleStoreServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AccessRuleStoreServer will
// result in compilation errors.
type UnsafeAccessRuleStoreServer interface {
	mustEmbedUnimplementedAccessRuleStoreServer()
}

func RegisterAccessRuleStoreServer(s *grpc.Server, srv AccessRuleStoreServer) {
	s.RegisterService(&_AccessRuleStore_serviceDesc, srv)
}

func _AccessRuleStore_PutRules_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRulesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccessRuleStoreServer).PutRules(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/AccessRuleStore/PutRules",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccessRuleStoreServer).PutRules(ctx, req.(*PutRulesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccessRuleStore_GetRules_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRulesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccessRuleStoreServer).GetRules(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/AccessRuleStore/GetRules",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccessRuleStoreServer).GetRules(ctx, req.(*GetRulesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccessRuleStore_GetRulesForPath_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRulesForPathRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccessRuleStoreServer).GetRulesForPath(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/AccessRuleStore/GetRulesForPath",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccessRuleStoreServer).GetRulesForPath(ctx, req.(*GetRulesForPathRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccessRuleStore_DeleteRules_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRulesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccessRuleStoreServer).DeleteRules(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/AccessRuleStore/DeleteRules",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccessRuleStoreServer).DeleteRules(ctx, req.(*DeleteRulesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AccessRuleStore_DeleteRulesForPath_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRulesForPathRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccessRuleStoreServer).DeleteRulesForPath(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/AccessRuleStore/DeleteRulesForPath",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccessRuleStoreServer).DeleteRulesForPath(ctx, req.(*DeleteRulesForPathRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _AccessRuleStore_serviceDesc = grpc.ServiceDesc{
	ServiceName: "AccessRuleStore",
	HandlerType: (*AccessRuleStoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PutRules",
			Handler:    _AccessRuleStore_PutRules_Handler,
		},
		{
			MethodName: "GetRules",
			Handler:    _AccessRuleStore_GetRules_Handler,
		},
		{
			MethodName: "GetRulesForPath",
			Handler:    _AccessRuleStore_GetRulesForPath_Handler,
		},
		{
			MethodName: "DeleteRules",
			Handler:    _AccessRuleStore_DeleteRules_Handler,
		},
		{
			MethodName: "DeleteRulesForPath",
			Handler:    _AccessRuleStore_DeleteRulesForPath_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/objects.proto",
}