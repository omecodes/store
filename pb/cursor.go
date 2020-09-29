package pb

type DataCursor interface {
	Walk() (bool, error)
	Get() *Data
	Close() error
}

type GraftCursor interface {
	Walk() (bool, error)
	Get() *Graft
	Close() error
}
