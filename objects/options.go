package objects

import pb "github.com/omecodes/store/gen/go/proto"

type CreateCollectionOptions struct{}

type GetCollectionOptions struct{}

type ListCollectionOptions struct{}

type DeleteCollectionOptions struct{}

type GetHeaderOptions struct{}

type DeleteObjectOptions struct{}

type CollectionOptions struct {
	Name       string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	FullObject bool   `protobuf:"varint,2,opt,name=full_object,json=fullObject,proto3" json:"full_object,omitempty"`
}

type ListOptions struct {
	At     string `protobuf:"bytes,3,opt,name=at,proto3" json:"at,omitempty"`
	Offset int64  `protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
}

type PutOptions struct{}

type GetObjectOptions struct {
	Info bool   `protobuf:"varint,1,opt,name=info,proto3" json:"info,omitempty"`
	At   string `protobuf:"bytes,2,opt,name=at,proto3" json:"at,omitempty"`
}

type PatchOptions struct{}

type MoveOptions struct {
	NewSecurity *pb.PathAccessRules
}

type SearchObjectsOptions struct{}
