package files

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/logs"
	"net/url"
	"strings"
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
	CreatedBy           string                 `json:"created_by,omitempty"`
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
	Delete(ctx context.Context, id string) error
	UserSources(ctx context.Context, username string) ([]*Source, error)
}

func ContextWithSourceManager(parent context.Context, manager SourceManager) context.Context {
	return context.WithValue(parent, ctxSourceManager{}, manager)
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
	if sourcesManager == nil {
		return nil, errors.Internal("missing source manager in context")
	}
	source, err := sourcesManager.Get(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	resolvedSource := source
	sourceChain := []string{sourceID}
	for resolvedSource.Type == TypeReference {
		u, err := url.Parse(source.URI)
		if err != nil {
			return nil, errors.Internal("could not resolve source uri", errors.Details{Key: "source uri", Value: err})
		}

		refSourceID := u.Host
		resolvedSource, err = sourcesManager.Get(ctx, refSourceID)
		if err != nil {
			logs.Error("could not load source", logs.Details("source", refSourceID), logs.Err(err))
			return nil, err
		}

		for _, src := range sourceChain {
			if src == refSourceID {
				return nil, errors.Internal("source cycle references")
			}
		}
		sourceChain = append(sourceChain, refSourceID)
		resolvedSource.URI = strings.TrimSuffix(resolvedSource.URI, "/") + u.Path

		logs.Info("resolved source", logs.Details("uri", resolvedSource.URI))
	}

	return resolvedSource, nil
}
