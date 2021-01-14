package pb

// Options
type CollectionOptions struct {
	Name       string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	FullObject bool   `protobuf:"varint,2,opt,name=full_object,json=fullObject,proto3" json:"full_object,omitempty"`
}

type RangeOptions struct {
	Offset int64  `protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	Count  uint64 `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
}

type ListOptions struct {
	At           string       `protobuf:"bytes,3,opt,name=at,proto3" json:"at,omitempty"`
	RangeOptions RangeOptions `protobuf:"bytes,5,opt,name=range_options,json=dateOptions,proto3" json:"range_options,omitempty"`
}

type PutOptions struct{}

type GetOptions struct {
	Info bool   `protobuf:"varint,1,opt,name=info,proto3" json:"info,omitempty"`
	At   string `protobuf:"bytes,2,opt,name=at,proto3" json:"at,omitempty"`
}

type PatchOptions struct{}

type MoveOptions struct {
	NewSecurity *PathAccessRules
}
