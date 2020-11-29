package oms

import (
	"bytes"
	"encoding/json"
	"io"
)

type Object struct {
	decoded bool
	header  *Header
	content io.Reader
}

func NewObject() *Object {
	o := new(Object)
	o.header = new(Header)
	return o
}

func DecodeObject(encoded string) (*Object, error) {
	o := new(Object)
	o.decoded = true
	var info Header
	err := json.Unmarshal([]byte(encoded), &info)
	if err != nil {
		return nil, err
	}
	o.content = bytes.NewBufferString(encoded)
	return o, nil
}

func (o *Object) SetID(id string) {
	o.header.Id = id
}

func (o *Object) SetCreatedBy(createdBy string) {
	o.header.CreatedBy = createdBy
}

func (o *Object) SetCreatedAt(createdAt int64) {
	o.header.CreatedAt = createdAt
}

func (o *Object) SetContent(reader io.Reader, length int64) {
	o.content = reader
	if o.header == nil {
		o.header = new(Header)
	}
	o.header.Size = length
}

func (o *Object) SetHeader(i *Header) {
	o.header = i
}

func (o *Object) ID() string {
	return o.header.Id
}

func (o *Object) Size() int64 {
	return o.header.Size
}

func (o *Object) CreatedBy() string {
	return o.header.CreatedBy
}

func (o *Object) CreatedAt() int64 {
	return o.header.CreatedAt
}

func (o *Object) Content() io.Reader {
	return o.content
}

func (o *Object) Header() *Header {
	return o.header
}
