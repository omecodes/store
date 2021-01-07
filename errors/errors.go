package errors

type ErrCause int

const (
	ACLNotFound    = ErrCause(1)
	ObjectNotFound = ErrCause(2)
)

type Error struct {
	Cause int
}

func (e Error) Error() string {
	panic("implement me")
}
