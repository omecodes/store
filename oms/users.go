package oms

type users struct{}

func (u *users) GetInfo(ID string) (*JSON, error) {
	return nil, nil
}

func (u *users) SaveInfo(ID string, info *JSON) error {
	return nil
}
