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
	p.path = path
	return p
}

func (p *Patch) SetContent(reader io.Reader) {
	p.content = reader
}

func (p *Patch) SetSize(size int64) {
	p.length = size
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

func (p *Patch) GetContent() io.Reader {
	return p.content
}
