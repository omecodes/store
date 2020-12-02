// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.13.0
// source: pb.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type AllowedTo int32

const (
	AllowedTo_doNothing  AllowedTo = 0
	AllowedTo_create     AllowedTo = 1
	AllowedTo_read       AllowedTo = 2
	AllowedTo_write      AllowedTo = 4
	AllowedTo_delete     AllowedTo = 8
	AllowedTo_editAcl    AllowedTo = 16
	AllowedTo_graft      AllowedTo = 32
	AllowedTo_doAnything AllowedTo = 63
)

// Enum value maps for AllowedTo.
var (
	AllowedTo_name = map[int32]string{
		0:  "doNothing",
		1:  "create",
		2:  "read",
		4:  "write",
		8:  "delete",
		16: "editAcl",
		32: "graft",
		63: "doAnything",
	}
	AllowedTo_value = map[string]int32{
		"doNothing":  0,
		"create":     1,
		"read":       2,
		"write":      4,
		"delete":     8,
		"editAcl":    16,
		"graft":      32,
		"doAnything": 63,
	}
)

func (x AllowedTo) Enum() *AllowedTo {
	p := new(AllowedTo)
	*p = x
	return p
}

func (x AllowedTo) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AllowedTo) Descriptor() protoreflect.EnumDescriptor {
	return file_pb_proto_enumTypes[0].Descriptor()
}

func (AllowedTo) Type() protoreflect.EnumType {
	return &file_pb_proto_enumTypes[0]
}

func (x AllowedTo) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AllowedTo.Descriptor instead.
func (AllowedTo) EnumDescriptor() ([]byte, []int) {
	return file_pb_proto_rawDescGZIP(), []int{0}
}

type RelationDef struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name        string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Equivalence bool   `protobuf:"varint,2,opt,name=equivalence,proto3" json:"equivalence,omitempty"`
}

func (x *RelationDef) Reset() {
	*x = RelationDef{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RelationDef) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RelationDef) ProtoMessage() {}

func (x *RelationDef) ProtoReflect() protoreflect.Message {
	mi := &file_pb_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RelationDef.ProtoReflect.Descriptor instead.
func (*RelationDef) Descriptor() ([]byte, []int) {
	return file_pb_proto_rawDescGZIP(), []int{0}
}

func (x *RelationDef) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *RelationDef) GetEquivalence() bool {
	if x != nil {
		return x.Equivalence
	}
	return false
}

type AccessRules struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Read   []string `protobuf:"bytes,1,rep,name=read,proto3" json:"read,omitempty"`
	Write  []string `protobuf:"bytes,2,rep,name=write,proto3" json:"write,omitempty"`
	Delete []string `protobuf:"bytes,3,rep,name=delete,proto3" json:"delete,omitempty"`
}

func (x *AccessRules) Reset() {
	*x = AccessRules{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AccessRules) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AccessRules) ProtoMessage() {}

func (x *AccessRules) ProtoReflect() protoreflect.Message {
	mi := &file_pb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AccessRules.ProtoReflect.Descriptor instead.
func (*AccessRules) Descriptor() ([]byte, []int) {
	return file_pb_proto_rawDescGZIP(), []int{1}
}

func (x *AccessRules) GetRead() []string {
	if x != nil {
		return x.Read
	}
	return nil
}

func (x *AccessRules) GetWrite() []string {
	if x != nil {
		return x.Write
	}
	return nil
}

func (x *AccessRules) GetDelete() []string {
	if x != nil {
		return x.Delete
	}
	return nil
}

type PathAccessRules struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AccessRules map[string]*AccessRules `protobuf:"bytes,5,rep,name=access_rules,json=accessRules,proto3" json:"access_rules,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *PathAccessRules) Reset() {
	*x = PathAccessRules{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PathAccessRules) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PathAccessRules) ProtoMessage() {}

func (x *PathAccessRules) ProtoReflect() protoreflect.Message {
	mi := &file_pb_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PathAccessRules.ProtoReflect.Descriptor instead.
func (*PathAccessRules) Descriptor() ([]byte, []int) {
	return file_pb_proto_rawDescGZIP(), []int{2}
}

func (x *PathAccessRules) GetAccessRules() map[string]*AccessRules {
	if x != nil {
		return x.AccessRules
	}
	return nil
}

type Header struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	CreatedBy string `protobuf:"bytes,2,opt,name=created_by,json=createdBy,proto3" json:"created_by,omitempty"`
	CreatedAt int64  `protobuf:"varint,3,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	Size      int64  `protobuf:"varint,4,opt,name=size,proto3" json:"size,omitempty"`
}

func (x *Header) Reset() {
	*x = Header{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Header) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Header) ProtoMessage() {}

func (x *Header) ProtoReflect() protoreflect.Message {
	mi := &file_pb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Header.ProtoReflect.Descriptor instead.
func (*Header) Descriptor() ([]byte, []int) {
	return file_pb_proto_rawDescGZIP(), []int{3}
}

func (x *Header) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Header) GetCreatedBy() string {
	if x != nil {
		return x.CreatedBy
	}
	return ""
}

func (x *Header) GetCreatedAt() int64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *Header) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

type Auth struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid       string   `protobuf:"bytes,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Email     string   `protobuf:"bytes,2,opt,name=email,proto3" json:"email,omitempty"`
	Worker    bool     `protobuf:"varint,3,opt,name=worker,proto3" json:"worker,omitempty"`
	Validated bool     `protobuf:"varint,4,opt,name=validated,proto3" json:"validated,omitempty"`
	Scope     []string `protobuf:"bytes,5,rep,name=scope,proto3" json:"scope,omitempty"`
	Group     string   `protobuf:"bytes,6,opt,name=group,proto3" json:"group,omitempty"`
}

func (x *Auth) Reset() {
	*x = Auth{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Auth) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Auth) ProtoMessage() {}

func (x *Auth) ProtoReflect() protoreflect.Message {
	mi := &file_pb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Auth.ProtoReflect.Descriptor instead.
func (*Auth) Descriptor() ([]byte, []int) {
	return file_pb_proto_rawDescGZIP(), []int{4}
}

func (x *Auth) GetUid() string {
	if x != nil {
		return x.Uid
	}
	return ""
}

func (x *Auth) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *Auth) GetWorker() bool {
	if x != nil {
		return x.Worker
	}
	return false
}

func (x *Auth) GetValidated() bool {
	if x != nil {
		return x.Validated
	}
	return false
}

func (x *Auth) GetScope() []string {
	if x != nil {
		return x.Scope
	}
	return nil
}

func (x *Auth) GetGroup() string {
	if x != nil {
		return x.Group
	}
	return ""
}

var File_pb_proto protoreflect.FileDescriptor

var file_pb_proto_rawDesc = []byte{
	0x0a, 0x08, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x43, 0x0a, 0x0b, 0x52, 0x65,
	0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x65, 0x66, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a,
	0x0b, 0x65, 0x71, 0x75, 0x69, 0x76, 0x61, 0x6c, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x0b, 0x65, 0x71, 0x75, 0x69, 0x76, 0x61, 0x6c, 0x65, 0x6e, 0x63, 0x65, 0x22,
	0x4f, 0x0a, 0x0b, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x52, 0x75, 0x6c, 0x65, 0x73, 0x12, 0x12,
	0x0a, 0x04, 0x72, 0x65, 0x61, 0x64, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x72, 0x65,
	0x61, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x77, 0x72, 0x69, 0x74, 0x65, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x05, 0x77, 0x72, 0x69, 0x74, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x22, 0xa5, 0x01, 0x0a, 0x0f, 0x50, 0x61, 0x74, 0x68, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x52,
	0x75, 0x6c, 0x65, 0x73, 0x12, 0x44, 0x0a, 0x0c, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x72,
	0x75, 0x6c, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x50, 0x61, 0x74,
	0x68, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x52, 0x75, 0x6c, 0x65, 0x73, 0x2e, 0x41, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x52, 0x75, 0x6c, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0b, 0x61,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x52, 0x75, 0x6c, 0x65, 0x73, 0x1a, 0x4c, 0x0a, 0x10, 0x41, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x52, 0x75, 0x6c, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x22, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0c, 0x2e, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x52, 0x75, 0x6c, 0x65, 0x73, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x6a, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x62, 0x79,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x42,
	0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74,
	0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04,
	0x73, 0x69, 0x7a, 0x65, 0x22, 0x90, 0x01, 0x0a, 0x04, 0x41, 0x75, 0x74, 0x68, 0x12, 0x10, 0x0a,
	0x03, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12,
	0x14, 0x0a, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x65, 0x6d, 0x61, 0x69, 0x6c, 0x12, 0x16, 0x0a, 0x06, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x12, 0x1c, 0x0a,
	0x09, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x09, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x73,
	0x63, 0x6f, 0x70, 0x65, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x73, 0x63, 0x6f, 0x70,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2a, 0x6f, 0x0a, 0x09, 0x41, 0x6c, 0x6c, 0x6f, 0x77,
	0x65, 0x64, 0x54, 0x6f, 0x12, 0x0d, 0x0a, 0x09, 0x64, 0x6f, 0x4e, 0x6f, 0x74, 0x68, 0x69, 0x6e,
	0x67, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x10, 0x01, 0x12,
	0x08, 0x0a, 0x04, 0x72, 0x65, 0x61, 0x64, 0x10, 0x02, 0x12, 0x09, 0x0a, 0x05, 0x77, 0x72, 0x69,
	0x74, 0x65, 0x10, 0x04, 0x12, 0x0a, 0x0a, 0x06, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x10, 0x08,
	0x12, 0x0b, 0x0a, 0x07, 0x65, 0x64, 0x69, 0x74, 0x41, 0x63, 0x6c, 0x10, 0x10, 0x12, 0x09, 0x0a,
	0x05, 0x67, 0x72, 0x61, 0x66, 0x74, 0x10, 0x20, 0x12, 0x0e, 0x0a, 0x0a, 0x64, 0x6f, 0x41, 0x6e,
	0x79, 0x74, 0x68, 0x69, 0x6e, 0x67, 0x10, 0x3f, 0x42, 0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6f, 0x6d, 0x65, 0x63, 0x6f, 0x64, 0x65, 0x73, 0x2f,
	0x6f, 0x6d, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2f, 0x70, 0x62, 0x2f, 0x70, 0x62, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x3b, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pb_proto_rawDescOnce sync.Once
	file_pb_proto_rawDescData = file_pb_proto_rawDesc
)

func file_pb_proto_rawDescGZIP() []byte {
	file_pb_proto_rawDescOnce.Do(func() {
		file_pb_proto_rawDescData = protoimpl.X.CompressGZIP(file_pb_proto_rawDescData)
	})
	return file_pb_proto_rawDescData
}

var file_pb_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_pb_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_pb_proto_goTypes = []interface{}{
	(AllowedTo)(0),          // 0: AllowedTo
	(*RelationDef)(nil),     // 1: RelationDef
	(*AccessRules)(nil),     // 2: AccessRules
	(*PathAccessRules)(nil), // 3: PathAccessRules
	(*Header)(nil),          // 4: Header
	(*Auth)(nil),            // 5: Auth
	nil,                     // 6: PathAccessRules.AccessRulesEntry
}
var file_pb_proto_depIdxs = []int32{
	6, // 0: PathAccessRules.access_rules:type_name -> PathAccessRules.AccessRulesEntry
	2, // 1: PathAccessRules.AccessRulesEntry.value:type_name -> AccessRules
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_pb_proto_init() }
func file_pb_proto_init() {
	if File_pb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RelationDef); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AccessRules); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PathAccessRules); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Header); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Auth); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pb_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pb_proto_goTypes,
		DependencyIndexes: file_pb_proto_depIdxs,
		EnumInfos:         file_pb_proto_enumTypes,
		MessageInfos:      file_pb_proto_msgTypes,
	}.Build()
	File_pb_proto = out.File
	file_pb_proto_rawDesc = nil
	file_pb_proto_goTypes = nil
	file_pb_proto_depIdxs = nil
}