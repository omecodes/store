package oms

import (
	"encoding/json"
	"github.com/omecodes/bome"
)

type PermissionsStore interface {
	Save(perm *Permission) error
	GetForUser(user string) ([]*Permission, error)
	GetForResource()
	Get(collection string, dataID string, user string) (*Permission, error)
	Delete(collection string, dataID string, user string) error
}

type permissionStore struct {
	bome.DoubleMap
}

func (p *permissionStore) Save(perm *Permission) error {
	data, err := json.Marshal(perm)
	if err != nil {
		return err
	}

	entry := &bome.DoubleMapEntry{
		FirstKey:  perm.Collection,
		SecondKey: perm.DataId,
		Value:     string(data),
	}
	return p.DoubleMap.Save(entry)
}

func (p *permissionStore) GetForUser(user string) ([]*Permission, error) {
	return nil, nil
}

func (p *permissionStore) GetForResource() {
	panic("implement me")
}

func (p *permissionStore) Get(collection string, dataID string, user string) (*Permission, error) {
	panic("implement me")
}

func (p *permissionStore) Delete(collection string, dataID string, user string) error {
	panic("implement me")
}
