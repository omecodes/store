package store

const (
	settingsDataPath                     = "/data"
	settingsDataMaxSizePath              = "/data/max_size"
	settingsSecurityPath                 = "/security"
	settingsSecurityRulesPath            = "/security/rules"
	settingsAccessSecurityRulesPath      = "/security/rules/access"
	settingsDataAccessSecurityRulesPath  = "/security/rules/access/data"
	settingsGraftAccessSecurityRulesPath = "/security/rules/access/grafts"
	settingsExtendableSecurityRulesPath  = "/security/rules/extendable"
)

var settingsPathFormats = map[string]string{
	settingsDataMaxSizePath: "%d",
}

var settingsPathValueMimes = map[string]string{
	settingsDataMaxSizePath: "text/plain",
}

const defaultSettings = `{
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
