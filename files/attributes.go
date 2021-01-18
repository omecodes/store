package files

import (
	"encoding/json"
	"github.com/omecodes/store/pb"
)

type Attributes map[string]string

type Attribute interface {
	Name() string
	Value() string
}

const (
	AttrPrefix = "store-"

	AttrPathHistory = AttrPrefix + "path-history"
	AttrSize        = AttrPrefix + "size"
	AttrCreatedBy   = AttrPrefix + "created-by"
	AttrCreatedAt   = AttrPrefix + "created-at"
	AttrHash        = AttrPrefix + "hash"
	AttrPermissions = AttrPrefix + "permissions"
)

func NewAttributesHolder() *AttributesHolder {
	return &AttributesHolder{
		Attributes: Attributes{},
	}
}

func HoldAttributes(attrs Attributes) *AttributesHolder {
	return &AttributesHolder{
		Attributes: attrs,
	}
}

type AttributesHolder struct {
	Attributes  Attributes
	permissions *pb.FilePermissions
}

func (h *AttributesHolder) SetPermissions(perms *pb.FilePermissions) error {
	return nil
}

func (h *AttributesHolder) SetEncodedPermissions(encoded string) error {
	return nil
}

func (h *AttributesHolder) AddReadPermission(permission *pb.Permission) {

}

func (h *AttributesHolder) GetPermissions() (*pb.FilePermissions, error) {
	if h.permissions == nil {
		encoded := h.Attributes[AttrPermissions]
		err := json.Unmarshal([]byte(encoded), &h.permissions)
		if err != nil {
			return nil, err
		}
	}
	return h.permissions, nil
}

func (h *AttributesHolder) GetAttributes() (Attributes, error) {
	return nil, nil
}
