package objects

import "github.com/omecodes/store/pb"

type DateRangeOptions struct {
	Before int64 `json:"before"`
	After  int64 `json:"after"`
}

type PutDataOptions struct {
	Indexes []*pb.Index
}

type PatchOptions struct{}

type DeleteOptions struct {
	Path string `json:"path"`
}

type DataOptions struct {
	Path string `json:"path"`
}

type ListOptions struct {
	Filter     ObjectFilter
	Collection string `json:"collection"`
	Path       string `json:"path"`
	Before     int64  `json:"before"`
	After      int64  `json:"after"`
	Count      int    `json:"count"`
}

type SearchParams struct {
	Collection string `json:"collection"`
	Condition  string `json:"condition"`
}

type SearchOptions struct {
	Filter ObjectFilter
	Path   string `json:"path"`
	Before int64  `json:"before"`
	After  int64  `json:"after"`
	Count  int    `json:"count"`
}

type SettingsOptions struct{}

type UserOptions struct {
	WithAccessList  bool
	WithPermissions bool
	WithGroups      bool
	WithPassword    bool
}

type GetObjectOptions struct {
	Path string `json:"path"`
	Info bool   `json:"header"`
}

type DatedRef struct {
	Date int64  `json:"date"`
	ID   string `json:"id"`
}
