package oms

const (
	SettingsDataPath                     = "/data"
	SettingsDataMaxSizePath              = "/data/max_size"
	SettingsSecurityPath                 = "/security"
	SettingsSecurityRulesPath            = "/security/rules"
	SettingsAccessSecurityRulesPath      = "/security/rules/access"
	SettingsDataAccessSecurityRulesPath  = "/security/rules/access/data"
	SettingsGraftAccessSecurityRulesPath = "/security/rules/access/grafts"
	SettingsExtendableSecurityRulesPath  = "/security/rules/extendable"
)

var SettingsPathFormats = map[string]string{
	SettingsDataMaxSizePath: "%d",
}

var SettingsPathValueMimes = map[string]string{
	SettingsDataMaxSizePath: "text/plain",
}

const DefaultSettings = `{
    "data": {
        "max_size": 5242880
    },
    "security": {
		"rules": {
			"access": {
				"data": {
					"create": "auth.validated",
					"read": "auth.validated && (auth.group == \"admin\" || auth.uid == data.creator)",
					"write": "auth.validated && (auth.group == \"admin\" || auth.uid == data.creator || acl(auth.uid, data.id).write)",
					"delete": "auth.validated && (auth.group == \"admin\" || auth.uid == data.creator || acl(auth.uid, data.id).delete)",
					"graft": "auth.validated && (auth.group == \"admin\" || auth.uid == data.creator || acl(auth.uid, data.id).graft)",
					"rules": "auth.validated && (auth.group == \"admin\" || auth.uid == data.creator)"
				},
				"grafts": {
					"read": "auth.validated && (graft.creator == auth.uid || graft.creator == auth.uid || acl(auth.uid, graft.did).read)",
					"write": "auth.validated && (graft.creator == auth.uid)",
					"delete": "auth.validated && (graft.creator == auth.uid)"
				}
			},
			"extendable": ["read"]
		}
    }
}`
