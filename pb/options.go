package pb

// Options
type CollectionOptions struct {
	Name       string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	FullObject bool   `protobuf:"varint,2,opt,name=full_object,json=fullObject,proto3" json:"full_object,omitempty"`
}

type DateRangeOptions struct {
	Before int64 `protobuf:"varint,1,opt,name=before,proto3" json:"before,omitempty"`
	After  int64 `protobuf:"varint,2,opt,name=after,proto3" json:"after,omitempty"`
}

type ListOptions struct {
	CollectionOptions CollectionOptions `protobuf:"bytes,1,opt,name=collection_options,json=collectionOptions,proto3" json:"collection_options,omitempty"`
	Condition         string            `protobuf:"bytes,2,opt,name=condition,proto3" json:"condition,omitempty"`
	At                string            `protobuf:"bytes,3,opt,name=at,proto3" json:"at,omitempty"`
	DateOptions       DateRangeOptions  `protobuf:"bytes,5,opt,name=date_options,json=dateOptions,proto3" json:"date_options,omitempty"`
}

type PutOptions struct{}

type GetOptions struct {
	Info bool   `protobuf:"varint,1,opt,name=info,proto3" json:"info,omitempty"`
	At   string `protobuf:"bytes,2,opt,name=at,proto3" json:"at,omitempty"`
}

type PatchOptions struct{}
