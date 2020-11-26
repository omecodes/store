package oms

import "io"

type Patch struct {
	objectID string
	path     string
	content  io.Reader
	length   int64
}

func NewPatch(objectID string, path string) *Patch {
	p := new(Patch)
	p.objectID = objectID
	return p
}

func (p *Patch) SetContent(reader io.Reader, length int64) {
	p.content = reader
	p.length = length
}

func (p *Patch) Size() int64 {
	return p.length
}

func (p *Patch) Path() string {
	return p.path
}

func (p *Patch) GetObjectID() string {
	return p.objectID
}

func (p *Patch) Marshal() ([]byte, error) {
	return nil, nil
}
