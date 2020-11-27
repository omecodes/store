package oms

const (
	SettingsDataMaxSizePath        = "/data_max_size"
	SettingsCreateDataSecurityRule = "/create_security_rule"
)

var SettingsPathFormats = map[string]string{
	SettingsDataMaxSizePath: "%d",
}

var SettingsPathValueMimes = map[string]string{
	SettingsDataMaxSizePath: "text/plain",
}

const DefaultSettings = `{
    "data_max_size": 5242880
	"create_security_rule": "auth.validated"
}`
