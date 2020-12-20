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

// HandlerUnitClient is the client API for HandlerUnit service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HandlerUnitClient interface {
	PutObject(ctx context.Context, in *PutObjectRequest, opts ...grpc.CallOption) (*PutObjectResponse, error)
	UpdateObject(ctx context.Context, in *UpdateObjectRequest, opts ...grpc.CallOption) (*UpdateObjectResponse, error)
	GetObject(ctx context.Context, in *GetObjectRequest, opts ...grpc.CallOption) (*GetObjectResponse, error)
	DeleteObject(ctx context.Context, in *DeleteObjectRequest, opts ...grpc.CallOption) (*DeleteObjectResponse, error)
	ObjectInfo(ctx context.Context, in *ObjectInfoRequest, opts ...grpc.CallOption) (*ObjectInfoResponse, error)
	ListObjects(ctx context.Context, in *ListObjectsRequest, opts ...grpc.CallOption) (HandlerUnit_ListObjectsClient, error)
	SearchObjects(ctx context.Context, in *SearchObjectsRequest, opts ...grpc.CallOption) (HandlerUnit_SearchObjectsClient, error)
}

type handlerUnitClient struct {
	cc grpc.ClientConnInterface
}

func NewHandlerUnitClient(cc grpc.ClientConnInterface) HandlerUnitClient {
	return &handlerUnitClient{cc}
}

func (c *handlerUnitClient) PutObject(ctx context.Context, in *PutObjectRequest, opts ...grpc.CallOption) (*PutObjectResponse, error) {
	out := new(PutObjectResponse)
	err := c.cc.Invoke(ctx, "/HandlerUnit/PutObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *handlerUnitClient) UpdateObject(ctx context.Context, in *UpdateObjectRequest, opts ...grpc.CallOption) (*UpdateObjectResponse, error) {
	out := new(UpdateObjectResponse)
	err := c.cc.Invoke(ctx, "/HandlerUnit/UpdateObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *handlerUnitClient) GetObject(ctx context.Context, in *GetObjectRequest, opts ...grpc.CallOption) (*GetObjectResponse, error) {
	out := new(GetObjectResponse)
	err := c.cc.Invoke(ctx, "/HandlerUnit/GetObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *handlerUnitClient) DeleteObject(ctx context.Context, in *DeleteObjectRequest, opts ...grpc.CallOption) (*DeleteObjectResponse, error) {
	out := new(DeleteObjectResponse)
	err := c.cc.Invoke(ctx, "/HandlerUnit/DeleteObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *handlerUnitClient) ObjectInfo(ctx context.Context, in *ObjectInfoRequest, opts ...grpc.CallOption) (*ObjectInfoResponse, error) {
	out := new(ObjectInfoResponse)
	err := c.cc.Invoke(ctx, "/HandlerUnit/ObjectInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *handlerUnitClient) ListObjects(ctx context.Context, in *ListObjectsRequest, opts ...grpc.CallOption) (HandlerUnit_ListObjectsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_HandlerUnit_serviceDesc.Streams[0], "/HandlerUnit/ListObjects", opts...)
	if err != nil {
		return nil, err
	}
	x := &handlerUnitListObjectsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type HandlerUnit_ListObjectsClient interface {
	Recv() (*DataObject, error)
	grpc.ClientStream
}

type handlerUnitListObjectsClient struct {
	grpc.ClientStream
}

func (x *handlerUnitListObjectsClient) Recv() (*DataObject, error) {
	m := new(DataObject)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *handlerUnitClient) SearchObjects(ctx context.Context, in *SearchObjectsRequest, opts ...grpc.CallOption) (HandlerUnit_SearchObjectsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_HandlerUnit_serviceDesc.Streams[1], "/HandlerUnit/SearchObjects", opts...)
	if err != nil {
		return nil, err
	}
	x := &handlerUnitSearchObjectsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type HandlerUnit_SearchObjectsClient interface {
	Recv() (*DataObject, error)
	grpc.ClientStream
}

type handlerUnitSearchObjectsClient struct {
	grpc.ClientStream
}

func (x *handlerUnitSearchObjectsClient) Recv() (*DataObject, error) {
	m := new(DataObject)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// HandlerUnitServer is the server API for HandlerUnit service.
// All implementations must embed UnimplementedHandlerUnitServer
// for forward compatibility
type HandlerUnitServer interface {
	PutObject(context.Context, *PutObjectRequest) (*PutObjectResponse, error)
	UpdateObject(context.Context, *UpdateObjectRequest) (*UpdateObjectResponse, error)
	GetObject(context.Context, *GetObjectRequest) (*GetObjectResponse, error)
	DeleteObject(context.Context, *DeleteObjectRequest) (*DeleteObjectResponse, error)
	ObjectInfo(context.Context, *ObjectInfoRequest) (*ObjectInfoResponse, error)
	ListObjects(*ListObjectsRequest, HandlerUnit_ListObjectsServer) error
	SearchObjects(*SearchObjectsRequest, HandlerUnit_SearchObjectsServer) error
	mustEmbedUnimplementedHandlerUnitServer()
}

// UnimplementedHandlerUnitServer must be embedded to have forward compatible implementations.
type UnimplementedHandlerUnitServer struct {
}

func (UnimplementedHandlerUnitServer) PutObject(context.Context, *PutObjectRequest) (*PutObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutObject not implemented")
}
func (UnimplementedHandlerUnitServer) UpdateObject(context.Context, *UpdateObjectRequest) (*UpdateObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateObject not implemented")
}
func (UnimplementedHandlerUnitServer) GetObject(context.Context, *GetObjectRequest) (*GetObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetObject not implemented")
}
func (UnimplementedHandlerUnitServer) DeleteObject(context.Context, *DeleteObjectRequest) (*DeleteObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteObject not implemented")
}
func (UnimplementedHandlerUnitServer) ObjectInfo(context.Context, *ObjectInfoRequest) (*ObjectInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ObjectInfo not implemented")
}
func (UnimplementedHandlerUnitServer) ListObjects(*ListObjectsRequest, HandlerUnit_ListObjectsServer) error {
	return status.Errorf(codes.Unimplemented, "method ListObjects not implemented")
}
func (UnimplementedHandlerUnitServer) SearchObjects(*SearchObjectsRequest, HandlerUnit_SearchObjectsServer) error {
	return status.Errorf(codes.Unimplemented, "method SearchObjects not implemented")
}
func (UnimplementedHandlerUnitServer) mustEmbedUnimplementedHandlerUnitServer() {}

// UnsafeHandlerUnitServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HandlerUnitServer will
// result in compilation errors.
type UnsafeHandlerUnitServer interface {
	mustEmbedUnimplementedHandlerUnitServer()
}

func RegisterHandlerUnitServer(s grpc.ServiceRegistrar, srv HandlerUnitServer) {
	s.RegisterService(&_HandlerUnit_serviceDesc, srv)
}

func _HandlerUnit_PutObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HandlerUnitServer).PutObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/HandlerUnit/PutObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HandlerUnitServer).PutObject(ctx, req.(*PutObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HandlerUnit_UpdateObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HandlerUnitServer).UpdateObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/HandlerUnit/UpdateObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HandlerUnitServer).UpdateObject(ctx, req.(*UpdateObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HandlerUnit_GetObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HandlerUnitServer).GetObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/HandlerUnit/GetObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HandlerUnitServer).GetObject(ctx, req.(*GetObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HandlerUnit_DeleteObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HandlerUnitServer).DeleteObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/HandlerUnit/DeleteObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HandlerUnitServer).DeleteObject(ctx, req.(*DeleteObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HandlerUnit_ObjectInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ObjectInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HandlerUnitServer).ObjectInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/HandlerUnit/ObjectInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HandlerUnitServer).ObjectInfo(ctx, req.(*ObjectInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HandlerUnit_ListObjects_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListObjectsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(HandlerUnitServer).ListObjects(m, &handlerUnitListObjectsServer{stream})
}

type HandlerUnit_ListObjectsServer interface {
	Send(*DataObject) error
	grpc.ServerStream
}

type handlerUnitListObjectsServer struct {
	grpc.ServerStream
}

func (x *handlerUnitListObjectsServer) Send(m *DataObject) error {
	return x.ServerStream.SendMsg(m)
}

func _HandlerUnit_SearchObjects_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SearchObjectsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(HandlerUnitServer).SearchObjects(m, &handlerUnitSearchObjectsServer{stream})
}

type HandlerUnit_SearchObjectsServer interface {
	Send(*DataObject) error
	grpc.ServerStream
}

type handlerUnitSearchObjectsServer struct {
	grpc.ServerStream
}

func (x *handlerUnitSearchObjectsServer) Send(m *DataObject) error {
	return x.ServerStream.SendMsg(m)
}

var _HandlerUnit_serviceDesc = grpc.ServiceDesc{
	ServiceName: "HandlerUnit",
	HandlerType: (*HandlerUnitServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PutObject",
			Handler:    _HandlerUnit_PutObject_Handler,
		},
		{
			MethodName: "UpdateObject",
			Handler:    _HandlerUnit_UpdateObject_Handler,
		},
		{
			MethodName: "GetObject",
			Handler:    _HandlerUnit_GetObject_Handler,
		},
		{
			MethodName: "DeleteObject",
			Handler:    _HandlerUnit_DeleteObject_Handler,
		},
		{
			MethodName: "ObjectInfo",
			Handler:    _HandlerUnit_ObjectInfo_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ListObjects",
			Handler:       _HandlerUnit_ListObjects_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "SearchObjects",
			Handler:       _HandlerUnit_SearchObjects_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pb.proto",
}

// ACLClient is the client API for ACL service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ACLClient interface {
	PutRules(ctx context.Context, in *PutRulesRequest, opts ...grpc.CallOption) (*PutRulesResponse, error)
	GetRules(ctx context.Context, in *GetRulesRequest, opts ...grpc.CallOption) (*GetRulesResponse, error)
	GetRulesForPath(ctx context.Context, in *GetRulesForPathRequest, opts ...grpc.CallOption) (*GetRulesForPathResponse, error)
	DeleteRules(ctx context.Context, in *DeleteRulesRequest, opts ...grpc.CallOption) (*DeleteRulesResponse, error)
	DeleteRulesForPath(ctx context.Context, in *DeleteRulesForPathRequest, opts ...grpc.CallOption) (*DeleteRulesForPathResponse, error)
}

type aCLClient struct {
	cc grpc.ClientConnInterface
}

func NewACLClient(cc grpc.ClientConnInterface) ACLClient {
	return &aCLClient{cc}
}

func (c *aCLClient) PutRules(ctx context.Context, in *PutRulesRequest, opts ...grpc.CallOption) (*PutRulesResponse, error) {
	out := new(PutRulesResponse)
	err := c.cc.Invoke(ctx, "/ACL/PutRules", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aCLClient) GetRules(ctx context.Context, in *GetRulesRequest, opts ...grpc.CallOption) (*GetRulesResponse, error) {
	out := new(GetRulesResponse)
	err := c.cc.Invoke(ctx, "/ACL/GetRules", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aCLClient) GetRulesForPath(ctx context.Context, in *GetRulesForPathRequest, opts ...grpc.CallOption) (*GetRulesForPathResponse, error) {
	out := new(GetRulesForPathResponse)
	err := c.cc.Invoke(ctx, "/ACL/GetRulesForPath", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aCLClient) DeleteRules(ctx context.Context, in *DeleteRulesRequest, opts ...grpc.CallOption) (*DeleteRulesResponse, error) {
	out := new(DeleteRulesResponse)
	err := c.cc.Invoke(ctx, "/ACL/DeleteRules", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aCLClient) DeleteRulesForPath(ctx context.Context, in *DeleteRulesForPathRequest, opts ...grpc.CallOption) (*DeleteRulesForPathResponse, error) {
	out := new(DeleteRulesForPathResponse)
	err := c.cc.Invoke(ctx, "/ACL/DeleteRulesForPath", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ACLServer is the server API for ACL service.
// All implementations must embed UnimplementedACLServer
// for forward compatibility
type ACLServer interface {
	PutRules(context.Context, *PutRulesRequest) (*PutRulesResponse, error)
	GetRules(context.Context, *GetRulesRequest) (*GetRulesResponse, error)
	GetRulesForPath(context.Context, *GetRulesForPathRequest) (*GetRulesForPathResponse, error)
	DeleteRules(context.Context, *DeleteRulesRequest) (*DeleteRulesResponse, error)
	DeleteRulesForPath(context.Context, *DeleteRulesForPathRequest) (*DeleteRulesForPathResponse, error)
	mustEmbedUnimplementedACLServer()
}

// UnimplementedACLServer must be embedded to have forward compatible implementations.
type UnimplementedACLServer struct {
}

func (UnimplementedACLServer) PutRules(context.Context, *PutRulesRequest) (*PutRulesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutRules not implemented")
}
func (UnimplementedACLServer) GetRules(context.Context, *GetRulesRequest) (*GetRulesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRules not implemented")
}
func (UnimplementedACLServer) GetRulesForPath(context.Context, *GetRulesForPathRequest) (*GetRulesForPathResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRulesForPath not implemented")
}
func (UnimplementedACLServer) DeleteRules(context.Context, *DeleteRulesRequest) (*DeleteRulesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteRules not implemented")
}
func (UnimplementedACLServer) DeleteRulesForPath(context.Context, *DeleteRulesForPathRequest) (*DeleteRulesForPathResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteRulesForPath not implemented")
}
func (UnimplementedACLServer) mustEmbedUnimplementedACLServer() {}

// UnsafeACLServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ACLServer will
// result in compilation errors.
type UnsafeACLServer interface {
	mustEmbedUnimplementedACLServer()
}

func RegisterACLServer(s grpc.ServiceRegistrar, srv ACLServer) {
	s.RegisterService(&_ACL_serviceDesc, srv)
}

func _ACL_PutRules_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRulesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ACLServer).PutRules(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ACL/PutRules",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ACLServer).PutRules(ctx, req.(*PutRulesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ACL_GetRules_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRulesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ACLServer).GetRules(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ACL/GetRules",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ACLServer).GetRules(ctx, req.(*GetRulesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ACL_GetRulesForPath_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRulesForPathRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ACLServer).GetRulesForPath(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ACL/GetRulesForPath",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ACLServer).GetRulesForPath(ctx, req.(*GetRulesForPathRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ACL_DeleteRules_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRulesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ACLServer).DeleteRules(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ACL/DeleteRules",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ACLServer).DeleteRules(ctx, req.(*DeleteRulesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ACL_DeleteRulesForPath_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRulesForPathRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ACLServer).DeleteRulesForPath(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ACL/DeleteRulesForPath",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ACLServer).DeleteRulesForPath(ctx, req.(*DeleteRulesForPathRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ACL_serviceDesc = grpc.ServiceDesc{
	ServiceName: "ACL",
	HandlerType: (*ACLServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PutRules",
			Handler:    _ACL_PutRules_Handler,
		},
		{
			MethodName: "GetRules",
			Handler:    _ACL_GetRules_Handler,
		},
		{
			MethodName: "GetRulesForPath",
			Handler:    _ACL_GetRulesForPath_Handler,
		},
		{
			MethodName: "DeleteRules",
			Handler:    _ACL_DeleteRules_Handler,
		},
		{
			MethodName: "DeleteRulesForPath",
			Handler:    _ACL_DeleteRulesForPath_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb.proto",
}

// SettingsClient is the client API for Settings service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SettingsClient interface {
	SetSettings(ctx context.Context, in *SetSettingsRequest, opts ...grpc.CallOption) (*SetSettingsResponse, error)
	GetSettings(ctx context.Context, in *GetSettingsRequest, opts ...grpc.CallOption) (*GetSettingsResponse, error)
}

type settingsClient struct {
	cc grpc.ClientConnInterface
}

func NewSettingsClient(cc grpc.ClientConnInterface) SettingsClient {
	return &settingsClient{cc}
}

func (c *settingsClient) SetSettings(ctx context.Context, in *SetSettingsRequest, opts ...grpc.CallOption) (*SetSettingsResponse, error) {
	out := new(SetSettingsResponse)
	err := c.cc.Invoke(ctx, "/Settings/SetSettings", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *settingsClient) GetSettings(ctx context.Context, in *GetSettingsRequest, opts ...grpc.CallOption) (*GetSettingsResponse, error) {
	out := new(GetSettingsResponse)
	err := c.cc.Invoke(ctx, "/Settings/GetSettings", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SettingsServer is the server API for Settings service.
// All implementations must embed UnimplementedSettingsServer
// for forward compatibility
type SettingsServer interface {
	SetSettings(context.Context, *SetSettingsRequest) (*SetSettingsResponse, error)
	GetSettings(context.Context, *GetSettingsRequest) (*GetSettingsResponse, error)
	mustEmbedUnimplementedSettingsServer()
}

// UnimplementedSettingsServer must be embedded to have forward compatible implementations.
type UnimplementedSettingsServer struct {
}

func (UnimplementedSettingsServer) SetSettings(context.Context, *SetSettingsRequest) (*SetSettingsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetSettings not implemented")
}
func (UnimplementedSettingsServer) GetSettings(context.Context, *GetSettingsRequest) (*GetSettingsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSettings not implemented")
}
func (UnimplementedSettingsServer) mustEmbedUnimplementedSettingsServer() {}

// UnsafeSettingsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SettingsServer will
// result in compilation errors.
type UnsafeSettingsServer interface {
	mustEmbedUnimplementedSettingsServer()
}

func RegisterSettingsServer(s grpc.ServiceRegistrar, srv SettingsServer) {
	s.RegisterService(&_Settings_serviceDesc, srv)
}

func _Settings_SetSettings_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetSettingsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettingsServer).SetSettings(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Settings/SetSettings",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettingsServer).SetSettings(ctx, req.(*SetSettingsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Settings_GetSettings_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSettingsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SettingsServer).GetSettings(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Settings/GetSettings",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SettingsServer).GetSettings(ctx, req.(*GetSettingsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Settings_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Settings",
	HandlerType: (*SettingsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetSettings",
			Handler:    _Settings_SetSettings_Handler,
		},
		{
			MethodName: "GetSettings",
			Handler:    _Settings_GetSettings_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb.proto",
}