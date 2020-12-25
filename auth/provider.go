package auth

type AuthenticationProvider interface {
	GetName() string
	Secret() string
	Verify()
}
