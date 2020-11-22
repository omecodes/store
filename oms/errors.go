package oms

type Error int

const (
	SchemaNotFound         = Error(1)
	ValueNotMatchingSchema = Error(2)
)

func (e Error) Error() string {
	switch e {
	case SchemaNotFound:
		return "schema not found"

	case ValueNotMatchingSchema:
		return "value not matching schema"

	default:
		return ""
	}
}
