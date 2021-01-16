package files

type EncryptionAlg int

const (
	AESGCM = EncryptionAlg(1)
)

type SourceType int

const (
	TypeLocal     = SourceType(1)
	TypeRemote    = SourceType(2)
	TypeSourceRef = SourceType(3)
)

type EncryptionInfo struct {
	Key []byte        `json:"key,omitempty"`
	Alg EncryptionAlg `json:"alg,omitempty"`
}

type Permissions struct {
	Read  []string `json:"read,omitempty"`
	Write []string `json:"write,omitempty"`
}

type Source struct {
	ID                  string          `json:"id,omitempty"`
	Label               string          `json:"label,omitempty"`
	Description         string          `json:"description,omitempty"`
	Type                SourceType      `json:"type,omitempty"`
	URI                 string          `json:"uri,omitempty"`
	Encryption          *EncryptionInfo `json:"encryption,omitempty"`
	PermissionOverrides *Permissions    `json:"permission_overrides,omitempty"`
}

type SourceManager interface {
	Save(source *Source) (string, error)
	Get(id string) (*Source, error)
	List() ([]*Source, error)
	Delete(id string) error
}
