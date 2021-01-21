package files

import (
	"context"
	"github.com/omecodes/errors"
	"net/url"
	"path"
)

const activeUserVar = "{user}"

const (
	SchemeFS     = "files"
	SchemeSource = "ref"
	SchemeHTTP   = "http"
	SchemeHTTPS  = "https"
	SchemeAWS    = "aws"
)

type SourceType int

type ctxSourceManager struct{}

const (
	TypeDisk      = SourceType(1)
	TypeActive    = SourceType(2)
	TypePartition = SourceType(3)
	TypeObjects   = SourceType(4)
	TypeReference = SourceType(5)
)

type Source struct {
	ID                  string                 `json:"id,omitempty"`
	Label               string                 `json:"label,omitempty"`
	Description         string                 `json:"description,omitempty"`
	Type                SourceType             `json:"type,omitempty"`
	URI                 string                 `json:"uri,omitempty"`
	Encryption          *EncryptionInfo        `json:"encryption,omitempty"`
	PermissionOverrides *Permissions           `json:"permission_overrides,omitempty"`
	ExpireTime          int64                  `json:"expire_time,omitempty"`
	Info                map[string]interface{} `json:"info,omitempty"`
}

type SourceManager interface {
	Save(ctx context.Context, source *Source) (string, error)
	Get(ctx context.Context, id string) (*Source, error)
	List(ctx context.Context) ([]*Source, error)
	Delete(ctx context.Context, id string) error
}

func getSourceManager(ctx context.Context) SourceManager {
	o := ctx.Value(ctxSourceManager{})
	if o == nil {
		return nil
	}
	return o.(SourceManager)
}

func resolveSource(ctx context.Context, sourceID string) (*Source, error) {
	sourcesManager := getSourceManager(ctx)
	source, err := sourcesManager.Get(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	sourceChain := []string{sourceID}
	sourceType := source.Type
	for sourceType == TypeReference {
		u, err := url.Parse(source.URI)
		if err != nil {
			return nil, errors.Create(errors.Internal, "could not resolve source uri", errors.Info{Name: "source uri", Details: err.Error()})
		}

		refSourceID := u.Host
		refSource, err := sourcesManager.Get(ctx, refSourceID)
		if err != nil {
			return nil, err
		}

		for _, src := range sourceChain {
			if src == refSourceID {
				return nil, errors.Create(errors.Internal, "source cycle references")
			}
		}
		sourceChain = append(sourceChain, refSourceID)

		source.URI = path.Join(refSource.URI, u.Path)
		sourceType = source.Type
	}

	return source, nil
}
