package files

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/pb"
	"net/url"
	"path"
	"strings"
)

type EncryptionAlg int

const (
	AESGCM = EncryptionAlg(1)
)

const (
	SchemeFS     = "files"
	SchemeSource = "ref"
	SchemeHTTP   = "http"
	SchemeHTTPS  = "https"
	SchemeAWS    = "aws"
)

type SourceType int

const (
	TypeFS        = SourceType(1)
	TypeRemote    = SourceType(2)
	TypeSourceRef = SourceType(3)
)

type EncryptionInfo struct {
	Key []byte        `json:"key,omitempty"`
	Alg EncryptionAlg `json:"alg,omitempty"`
}

type Source struct {
	ID                  string              `json:"id,omitempty"`
	Label               string              `json:"label,omitempty"`
	Description         string              `json:"description,omitempty"`
	Type                SourceType          `json:"type,omitempty"`
	URI                 string              `json:"uri,omitempty"`
	Encryption          *EncryptionInfo     `json:"encryption,omitempty"`
	PermissionOverrides *pb.FilePermissions `json:"permission_overrides,omitempty"`
	ExpireTime          int64               `json:"expire_time,omitempty"`
}

type SourceManager interface {
	Save(source *Source) (string, error)
	Get(id string) (*Source, error)
	List() ([]*Source, error)
	Delete(id string) error
}

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

func ResolveSource(ctx context.Context, sourceID string) (*Source, error) {
	sourcesManager := GetSourceManager(ctx)
	source, err := sourcesManager.Get(sourceID)
	if err != nil {
		return nil, err
	}

	sourceChain := []string{sourceID}
	sourceType := source.Type
	for sourceType == TypeSourceRef {
		u, err := url.Parse(source.URI)
		if err != nil {
			return nil, errors.Create(errors.Internal, "could not resolve source uri", errors.Info{Name: "source uri", Details: err.Error()})
		}

		refSourceID := u.Host
		refSource, err := sourcesManager.Get(refSourceID)
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
