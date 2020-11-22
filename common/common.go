package common

type CredentialsValidator interface {
	Validate(access, secret string) (bool, error)
}

type CredentialsValidateFunc func(string, string) (bool, error)

func (f CredentialsValidateFunc) Validate(access, secret string) (bool, error) {
	return f(access, secret)
}

type ErrorHandler interface {
	Handle(err error)
}

type ErrorHandlerFunc func(err error)

func (f ErrorHandlerFunc) Handle(err error) {
	f(err)
}
