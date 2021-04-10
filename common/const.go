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
	HttpHeaderContentType   = "Content-Type"
	HttpHeaderContentLength = "Content-Length"
	HttpHeaderAccept        = "Accept"
)

const (
	ContentTypeJSONStream = "application/stream+json"
	ContentTypeJSON       = "application/json"
	AllJSONContentTypes   = ContentTypeJSONStream + "," + ContentTypeJSON
)

const (
	ApiQueryParamOffset = "offset"
	ApiQueryParamAt     = "at"
	ApiQueryParamPath   = "path"
	ApiQueryParamHeader = "header"

	ApiSetSettingsRoute = "/objects/settings"
	ApiGetSettingsRoute = "/objects/settings"

	ApiRouteVarId         = "{id}"
	ApiRouteVarSource     = "{source}"
	ApiRouteVarCollection = "{collection}"

	ApiRouteVarIdName         = "id"
	ApiRouteVarSourceName     = "source"
	ApiRouteVarCollectionName = "collection"

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
