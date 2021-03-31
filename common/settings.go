package common

const (
	SettingsDataMaxSizePath        = "data_max_size"
	SettingsCreateDataSecurityRule = "create_security_rule"
	SettingsObjectListMaxCount     = "object_list_max_count"
)

var DefaultSettings = map[string]string{
	SettingsDataMaxSizePath:        "5242880",
	SettingsCreateDataSecurityRule: "auth.worker",
	SettingsObjectListMaxCount:     "5",
}

type SettingsManager interface {
	Set(string, string) error
	Get(string) (string, error)
	Delete(string) error
	Clear() error
}
