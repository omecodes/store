package files

import (
	"encoding/json"
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
	permissions *Permissions
}

func (h *AttributesHolder) SetPermissions(perms *Permissions) error {
	return nil
}

func (h *AttributesHolder) SetEncodedPermissions(encoded string) error {
	return nil
}

func (h *AttributesHolder) AddReadPermissions(permission *Permissions) {

}

func (h *AttributesHolder) GetPermissions() (*Permissions, bool, error) {
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
