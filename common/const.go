package common

const (
	AdminAuthFile      = "./admin.auth"
	CookiesKeyFilename = "./cookies.key"
)

const (
	HttpHeaderContentType = "Content-Type"

	HttpHeaderAccept                   = "Accept"
	HttpHeaderAccessControlAllowOrigin = "Access-Control-Allow-Origin"
	HttpHeaderUserAuthorization        = "Authorization"
	HttpHeaderAppAuthorization         = "X-STORE-CLIENT-APP-AUTHENTICATION"
)

const (
	ContentTypeJSONStream = "application/stream+json"
	ContentTypeJSON       = "application/json"
)

const (
	ApiDefaultLocation     = "/api"
	ApiObjectsRoutePrefix  = "/api/objects"
	ApiFilesRoutePrefix    = "/api/files"
	ApiAuthRoutePrefix     = "/api/auth"
	ApiAccountsRoutePrefix = "/api/accounts"
	ApiSettingsRoutePrefix = "/api/settings"

	ApiParamName = "name"

	ApiParamUsername = "username"
	ApiParamPassword = "password"

	ApiParamContinueURL = "continue"

	ApiRouteVarId = "{id}"

	ApiRouteVarIdName = "id"

	ApiRouteVarCollectionName = "collection"

	// API Routes
	//

	ApiSetSettingsRoute = "/settings"
	ApiGetSettingsRoute = "/settings"

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

	ApiCreateFileAccess          = "/files/accesses"
	ApiListFileAccesses          = "/files/accesses"
	ApiGetFileAccess             = "/files/accesses/{id}"
	ApiDeleteFileAccess          = "/files/accesses/{id}"
	ApiFileTreeRoutePrefix       = "/files/tree"
	ApiFileAttributesRoutePrefix = "/files/attributes"
	ApiFileDataRoutePrefix       = "/files/data"
)
