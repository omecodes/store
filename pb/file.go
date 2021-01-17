package pb

type File struct {
	Path       string            `json:"path,omitempty"`
	IsDir      bool              `json:"is_dir,omitempty"`
	Size       uint64            `json:"size,omitempty"`
	CreateTime int64             `json:"create_time,omitempty"`
	Hash       string            `json:"hash,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type PutFileOptions struct {
	Append      bool             `json:"append,omitempty"`
	Hash        string           `json:"hash,omitempty"`
	Permissions *FileAccessRules `json:"permissions,omitempty"`
}

type ContentRange struct {
	Offset uint64 `json:"offset,omitempty"`
	Length int64  `json:"length,omitempty"`
}

type GetFileOptions struct {
	Range ContentRange `json:"range,omitempty"`
}

type GetFileInfoOptions struct {
	WithHash bool `json:"with_hash,omitempty"`
	WithMeta bool `json:"with_meta,omitempty"`
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
