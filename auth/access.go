package auth

const (
	AccessTypeAuthorities = "authorities"
	AccessTypeWorker      = "worker"
	AccessTypeUserApp     = "client"
)

type APIAccess struct {
	Key    string            `json:"key,omitempty"`
	Secret string            `json:"secret,omitempty"`
	Type   string            `json:"type,omitempty"`
	Info   map[string]string `json:"info,omitempty"`
}
