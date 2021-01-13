package accounts

import "context"

type Source struct {
	Provider  string `json:"provider,omitempty"`
	Name      string `json:"name,omitempty"`
	FullName  string `json:"source_name,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
}

type Session struct {
	Token     string `json:"token,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UserAgent string `json:"client,omitempty"`
	Device    string `json:"device,omitempty"`
}

type Account struct {
	Login       string              `json:"login,omitempty"`
	Source      *Source             `json:"source,omitempty"`
	CreatedAt   int64               `json:"created_at,omitempty"`
	Sessions    map[string]*Session `json:"sessions,omitempty"`
	Preferences map[string]string   `json:"preferences,omitempty"`
}

type Manager interface {
	Create(ctx context.Context, account *Account) error
	Get(ctx context.Context, username string) (*Account, error)
	Find(ctx context.Context, provider string, originalName string) (*Account, error)
	Search(ctx context.Context, pattern string) ([]string, error)
}
