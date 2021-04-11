package settings

const (
	DataMaxSizePath        = "data_max_size"
	CreateDataSecurityRule = "create_security_rule"
	ObjectListMaxCount     = "object_list_max_count"
)

var Default = map[string]string{
	DataMaxSizePath:        "5242880",
	CreateDataSecurityRule: "auth.worker",
	ObjectListMaxCount:     "5",
}

type Manager interface {
	Set(string, string) error
	Get(string) (string, error)
	Delete(string) error
	Clear() error
}
