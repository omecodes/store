package files

import (
	"strings"
)

type FileLocation struct {
	Source   string `json:"source,omitempty"`
	Filename string `json:"filename,omitempty"`
}

type TreePatchInfo struct {
	Rename bool   `json:"rename,omitempty"`
	Value  string `json:"value,omitempty"`
}

type DirContent struct {
	Files  []*File `json:"files,omitempty"`
	Total  int     `json:"total"`
	Offset int     `json:"offset"`
}

type ListDirOptions struct {
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

type WriteOptions struct {
	Append      bool         `json:"append,omitempty"`
	Hash        string       `json:"hash,omitempty"`
	Permissions *Permissions `json:"permissions,omitempty"`
}

type ContentRange struct {
	Offset int64 `json:"offset,omitempty"`
	Length int64 `json:"length,omitempty"`
}

type ReadOptions struct {
	Range ContentRange `json:"range,omitempty"`
}

type GetFileOptions struct {
	WithAttrs bool `json:"with_attrs,omitempty"`
}

type DeleteFileOptions struct {
	Recursive bool `json:"recursive,omitempty"`
}

type MultipartSessionInfo struct {
	ID          string `json:"id,omitempty"`
	User        string `json:"user,omitempty"`
	PartCount   int    `json:"part_count,omitempty"`
	ContentHash string `json:"content_hash"`
}

type ContentPartInfo struct {
	ID          string `json:"id,omitempty"`
	User        string `json:"user,omitempty"`
	PartNumber  int    `json:"part_number,omitempty"`
	ContentHash string `json:"content_hash"`
}

type AddContentPartOptions struct{}

func Split(filename string) (string, string) {
	if filename == "" || filename == "/" {
		return "", ""
	}

	pathComponents := strings.Split(strings.TrimPrefix(filename, "/"), "/")
	if len(pathComponents) < 2 {
		return "", ""
	}

	return pathComponents[0], "/" + strings.Join(pathComponents[1:], "/")
}
