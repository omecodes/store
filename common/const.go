package common

const (
	UserAuthenticationHeaderName = "Authorization"
	AppAuthenticationHeaderName  = "X-STORE-CLIENT-APP-AUTHENTICATION"
)

const (
	AccessInfoEncryptedSecret     = "encrypted_secret"
	AccessInfoSecretEncryptParams = "encrypted_secret_params"
)

const (
	AdminAuthFile         = "./admin.auth"
	CookiesKeyFilename    = "./cookies.key"
	ServiceAuthSecretFile = "./services.auth"
)

const (
	CACertificateFilename = "service-ca.crt"
	CAKeyFilename         = "service-ca.key"
)

const (
	HttpHeaderContentType              = "Content-Type"
	HttpHeaderContentLength            = "Content-Length"
	HttpHeaderAccept                   = "Accept"
	HttpHeaderAccessControlAllowOrigin = "Access-Control-Allow-Origin"
)

const (
	ContentTypeJSONStream = "application/stream+json"
	ContentTypeJSON       = "application/json"
	AllJSONContentTypes   = ContentTypeJSONStream + "," + ContentTypeJSON
)

const (
	ApiDefaultLocation     = "/api"
	ApiObjectsRoutePrefix  = "/api/objects"
	ApiFilesRoutePrefix    = "/api/files"
	ApiAuthRoutePrefix     = "/api/auth"
	ApiAccountsRoutePrefix = "/api/accounts"
	ApiSettingsRoutePrefix = "/api/settings"

	ApiQueryParamOffset = "offset"
	ApiParamAt          = "at"
	ApiParamName        = "name"
	ApiParamQuery       = "q"
	ApiParamUsername    = "username"
	ApiParamPassword    = "password"
	ApiQueryParamPath   = "path"
	ApiParamHeader      = "header"
	ApiParamContinueURL = "continue"

	ApiSetSettingsRoute = "/objects/settings"
	ApiGetSettingsRoute = "/objects/settings"

	ApiRouteVarId         = "{id}"
	ApiRouteVarKey        = "{key}"
	ApiRouteVarSource     = "{source}"
	ApiRouteVarName       = "{name}"
	ApiRouteVarCollection = "{collection}"

	ApiRouteVarIdName         = "id"
	ApiRouteVarKeyName        = "key"
	ApiRouteVarSourceName     = "source"
	ApiRouteVarNameName       = "name"
	ApiRouteVarCollectionName = "collection"

	// API Routes
	//

	ApiLoginRoute = "/login"

	ApiGetAccountRoute    = "/accounts/{id}"
	ApiCreateAccountRoute = "/accounts/{id}"
	ApiFindAccountRoute   = "/accounts/{id}"

	ApiSaveAuthProviderRoute   = "/auth/providers"
	ApiGetAuthProviderRoute    = "/auth/providers/{id}"
	ApiDeleteAuthProviderRoute = "/auth/providers/{id}"
	ApiListAuthProvidersRoute  = "/auth/providers"

	ApiSaveClientAppRoute    = "/auth/apps"
	ApiListClientAppsRoute   = "/auth/apps"
	ApiGetClientAppRoute     = "/auth/apps/{id}"
	ApiDeleteClientAppRoute  = "/auth/apps/{id}"
	ApiCreateAppSessionRoute = "/auth/sessions/client-app"
	ApiCreateUserRoute       = "/auth/users"
	ApiSearchUsersRoute      = "/auth/users"

	ApiCreateCollectionRoute = "/objects/collections"
	ApiListCollectionRoute   = "/objects/collections"
	ApiGetCollectionRoute    = "/objects/collections/{id}"
	ApiDeleteCollectionRoute = "/objects/collections/{id}"
	ApiPutObjectRoute        = "/objects/data/{collection}"
	ApiPatchObjectRoute      = "/objects/data/{collection}/{id}"
	ApiMoveObjectRoute       = "/objects/data/{collection}/{id}"
	ApiGetObjectRoute        = "/objects/data/{collection}/{id}"
	ApiDeleteObjectRoute     = "/objects/data/{collection}/{id}"
	ApiListObjectsRoute      = "/objects/data/{collection}"
	ApiSearchObjectsRoute    = "/objects/data/{collection}"

	ApiCreateFileSource          = "/files/sources"
	ApiListFileSources           = "/files/sources"
	ApiGetFileSource             = "/files/sources/{id}"
	ApiDeleteFileSource          = "/files/sources/{id}"
	ApiFileTreeRoutePrefix       = "/files/tree"
	ApiFileAttributesRoutePrefix = "/files/attributes"
	ApiFileDataRoutePrefix       = "/files/data"
)
