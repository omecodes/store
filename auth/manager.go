package auth

type CredentialsManager interface {
	Get(user string) (string, error)
}

type CredentialsMangerFunc func(string) (string, error)

func (f CredentialsMangerFunc) Get(user string) (string, error) {
	return f(user)
}
