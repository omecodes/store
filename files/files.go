package files

import (
	pb "github.com/omecodes/store/gen/go/proto"
	"strings"
)

type CreateAccessOptions struct{}

type GetAccessListOptions struct{}

type GetAccessOptions struct {
	Resolved bool `json:"resolved,omitempty"`
}

type DeleteAccessOptions struct{}

type CreateDirOptions struct{}

type SetFileAttributesOptions struct{}

type GetFileAttributesOptions struct{}

type RenameFileOptions struct{}

type MoveFileOptions struct{}

type CopyFileOptions struct{}

type OpenMultipartSessionOptions struct{}

type WriteFilePartOptions struct{}

type CloseMultipartSessionOptions struct{}

type GetFSAccessOptions struct {
	Resolved bool `json:"resolved,omitempty"`
}

type FileLocation struct {
	AccessID string `json:"access_id,omitempty"`
	Filename string `json:"filename,omitempty"`
}

type TreePatchInfo struct {
	Rename bool   `json:"rename,omitempty"`
	Value  string `json:"value,omitempty"`
}

type DirContent struct {
	Files  []*pb.File `json:"files,omitempty"`
	Total  int        `json:"total"`
	Offset int        `json:"offset"`
}

type ListDirOptions struct {
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

type WriteOptions struct {
	Append bool   `json:"append,omitempty"`
	Hash   string `json:"hash,omitempty"`
}

type ContentRange struct {
	Offset int64 `json:"offset,omitempty"`
	Length int64 `json:"length,omitempty"`
}

type ReadOptions struct {
	Range ContentRange `json:"range,omitempty"`
}

type ShareOptions struct{}

type GetSharesOptions struct{}

type DeleteSharesOptions struct{}

type GetFileOptions struct {
	WithAttrs bool `json:"with_attrs,omitempty"`
}

type DeleteFileOptions struct {
	Recursive bool `json:"recursive,omitempty"`
}

type MultipartSessionInfo struct{}

type ContentPartInfo struct {
	ID          string `json:"id,omitempty"`
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
