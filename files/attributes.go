package files

import (
	"encoding/json"
	pb "github.com/omecodes/store/gen/go/proto"
)

type Attributes map[string]string

type Attribute interface {
	Name() string
	Value() string
}

// AttrPathHistory = AttrPrefix + "path-history"
// AttrSize        = AttrPrefix + "size"
// AttrCreatedBy   = AttrPrefix + "created-by"
// AttrCreatedAt   = AttrPrefix + "created-at"
// AttrHash        = AttrPrefix + "hash"

const (
	AttrPrefix = "store-"

	AttrPermissions = AttrPrefix + "permissions"
)

/*func NewAttributesHolder() *AttributesHolder {
	return &AttributesHolder{
		Attributes: Attributes{},
	}
}

func HoldAttributes(attrs Attributes) *AttributesHolder {
	return &AttributesHolder{
		Attributes: attrs,
	}
} */

type AttributesHolder struct {
	Attributes  Attributes
	permissions *pb.AccessActionRelation
}

func (h *AttributesHolder) SetPermissions(_ *pb.AccessActionRelation) error {
	return nil
}

func (h *AttributesHolder) SetEncodedPermissions(_ string) error {
	return nil
}

func (h *AttributesHolder) AddReadPermissions(_ *pb.AccessActionRelation) {

}

func (h *AttributesHolder) GetPermissions() (*pb.AccessActionRelation, bool, error) {
	if h.permissions == nil {
		encoded := h.Attributes[AttrPermissions]
		if encoded == "" {
			return nil, false, nil
		}

		err := json.Unmarshal([]byte(encoded), &h.permissions)
		if err != nil {
			return nil, false, err
		}
	}
	return h.permissions, true, nil
}

func (h *AttributesHolder) GetAttributes() (Attributes, error) {
	return nil, nil
}
