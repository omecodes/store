package pb

type File struct {
	Name       string            `json:"name,omitempty"`
	IsDir      bool              `json:"is_dir,omitempty"`
	Size       int64             `json:"size,omitempty"`
	CreateTime int64             `json:"create_time,omitempty"`
	ModTime    int64             `json:"mod_time,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
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

type PutFileOptions struct {
	Append      bool             `json:"append,omitempty"`
	Hash        string           `json:"hash,omitempty"`
	Permissions *FilePermissions `json:"permissions,omitempty"`
}

type ContentRange struct {
	Offset int64 `json:"offset,omitempty"`
	Length int64 `json:"length,omitempty"`
}

type GetFileOptions struct {
	Range ContentRange `json:"range,omitempty"`
}

type GetFileInfoOptions struct {
	WithAttrs bool `json:"with_attrs,omitempty"`
}

type DeleteFileOptions struct {
	Recursive    bool `json:"recursive,omitempty"`
	PutInRecycle bool `json:"put_in_recycle"`
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
