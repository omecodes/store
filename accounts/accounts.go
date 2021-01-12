package accounts

type Session struct {
	Token     string `json:"token,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UserAgent string `json:"client,omitempty"`
	Device    string `json:"device,omitempty"`
}

type Account struct {
	Name        string              `json:"source,omitempty"`
	FullName    string              `json:"source_name,omitempty"`
	FirstName   string              `json:"first_name,omitempty"`
	LastName    string              `json:"last_name,omitempty"`
	Email       string              `json:"email,omitempty"`
	Login       string              `json:"login,omitempty"`
	CreatedAt   int64               `json:"created_at,omitempty"`
	Sessions    map[string]*Session `json:"sessions,omitempty"`
	Preferences map[string]string   `json:"preferences,omitempty"`
}

type Manager interface {
	Create(account *Account) error
	Get(username string) (*Account, error)
	Find(provider string, originalName string) (*Account, error)
	Search(pattern string) ([]string, error)
}
