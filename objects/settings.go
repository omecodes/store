package objects

const (
	SettingsDataMaxSizePath        = "data_max_size"
	SettingsCreateDataSecurityRule = "create_security_rule"
)

var SettingsPathFormats = map[string]string{
	SettingsDataMaxSizePath:        "%d",
	SettingsCreateDataSecurityRule: "%s",
}

var SettingsPathValueMimes = map[string]string{
	SettingsDataMaxSizePath: "text/plain",
}

var DefaultSettings = map[string]string{
	SettingsDataMaxSizePath:        "5242880",
	SettingsCreateDataSecurityRule: "auth.worker",
}

type SettingsManager interface {
	Set(string, string) error
	Get(string) (string, error)
	Delete(string) error
	Clear() error
}
