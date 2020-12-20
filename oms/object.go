package oms

import (
	"bytes"
	"github.com/omecodes/omestore/pb"
	"io"
)

type Object struct {
	decoded bool
	header  *pb.Header
	content io.Reader
}

func NewObject() *Object {
	o := new(Object)
	o.header = new(pb.Header)
	return o
}

func DecodeObject(encoded string) (*Object, error) {
	o := new(Object)
	o.decoded = true
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

func (o *Object) SetSize(size int64) {
	o.header.Size = size
}

func (o *Object) SetContent(reader io.Reader) {
	o.content = reader
}

func (o *Object) SetHeader(i *pb.Header) {
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

func (o *Object) GetContent() io.Reader {
	return o.content
}

func (o *Object) Header() *pb.Header {
	return o.header
}
