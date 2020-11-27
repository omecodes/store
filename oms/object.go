package oms

import (
	"bytes"
	"encoding/json"
	"io"
)

const (
	infoKey = ".info"
)

type Object struct {
	decoded bool
	info    *Info
	content io.Reader
}

func NewObject() *Object {
	o := new(Object)
	o.info = new(Info)
	return o
}

func DecodeObject(encoded string) (*Object, error) {
	o := new(Object)
	o.decoded = true
	var info Info
	err := json.Unmarshal([]byte(encoded), &info)
	if err != nil {
		return nil, err
	}
	o.content = bytes.NewBufferString(encoded)
	return o, nil
}

func (o *Object) SetID(id string) {
	o.info.Id = id
}

func (o *Object) SetCreatedBy(createdBy string) {
	o.info.CreatedBy = createdBy
}

func (o *Object) SetCreatedAt(createdAt int64) {
	o.info.CreatedAt = createdAt
}

func (o *Object) SetContent(reader io.Reader, length int64) {
	o.content = reader
	o.info.Size = length
}

func (o *Object) ID() string {
	return o.info.Id
}

func (o *Object) Size() int64 {
	return o.info.Size
}

func (o *Object) CreatedBy() string {
	return o.info.CreatedBy
}

func (o *Object) CreatedAt() int64 {
	return o.info.CreatedAt
}

func (o *Object) Content() io.Reader {
	return o.content
}

func (o *Object) Header() *Info {
	return o.info
}
