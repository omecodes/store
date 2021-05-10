package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/accounts"
	"github.com/omecodes/store/auth"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/files"
	pb "github.com/omecodes/store/gen/go/proto"
	"github.com/omecodes/store/objects"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func New(host string, opts ...Option) *Client {
	c := &Client{
		host: host,
	}
	c.options.apiLocation = common.ApiDefaultLocation
	c.options.port = 443
	for _, opt := range opts {
		opt(&c.options)
	}

	c.httpClient = &http.Client{
		Transport: &http.Transport{TLSClientConfig: c.options.tlsConfig},
	}
	return c
}

type Client struct {
	host       string
	options    options
	httpClient *http.Client
}

func (c *Client) fullAPILocation() string {
	if c.options.noTLS {
		return fmt.Sprintf("http://%s:%d%s", c.host, c.options.port, path.Join("/", c.options.apiLocation))
	}
	return fmt.Sprintf("https://%s:%d%s", c.host, c.options.port, path.Join("/", c.options.apiLocation))
}

func (c *Client) request(method string, endpoint string, headers http.Header, body io.Reader) (*http.Response, error) {
	if !strings.HasPrefix(endpoint, "http") {
		if c.options.noTLS {
			endpoint = "http://" + endpoint
		} else {
			endpoint = "https://" + endpoint
		}
	}

	fmt.Println(method + " " + endpoint)

	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}

	if c.options.userAuth != nil {
		req.Header.Set(c.options.userAuth.HeaderKey(), c.options.userAuth.HeaderValue())
	}
	if c.options.appAuth != nil {
		req.Header.Set(c.options.appAuth.HeaderKey(), c.options.appAuth.HeaderValue())
	}

	if headers != nil {
		for k := range headers {
			req.Header.Set(k, headers.Get(k))
		}
	}
	return c.httpClient.Do(req)
}

func (c *Client) CreateObjectsCollection(collection *pb.Collection) error {
	endpoint := fmt.Sprintf(c.fullAPILocation() + common.ApiCreateCollectionRoute)

	encoded, err := json.Marshal(collection)
	if err != nil {
		return err
	}

	rsp, err := c.request(http.MethodPut, endpoint, nil, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) ListCollections() ([]*pb.Collection, error) {
	endpoint := fmt.Sprintf(c.fullAPILocation() + common.ApiListCollectionRoute)

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var collections []*pb.Collection
	err = json.NewDecoder(rsp.Body).Decode(&collections)
	return collections, err
}

func (c *Client) GetCollection(collectionId string) (*pb.Collection, error) {
	endpoint := c.fullAPILocation() + common.ApiGetCollectionRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, collectionId, 1)

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	col := new(pb.Collection)
	return col, jsonpb.Unmarshal(rsp.Body, col)
}

func (c *Client) DeleteCollection(collectionId string) error {
	endpoint := c.fullAPILocation() + common.ApiDeleteCollectionRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, collectionId, 1)

	rsp, err := c.request(http.MethodDelete, endpoint, nil, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()
	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) PutObject(collectionId string, object *pb.Object, accessRules *pb.PathAccessRules, indexes ...*pb.TextIndex) error {
	endpoint := c.fullAPILocation() + common.ApiPutObjectRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarCollection, collectionId, 1)

	buff := bytes.NewBuffer(nil)
	encoder := jsonpb.Marshaler{EnumsAsInts: true}
	err := encoder.Marshal(buff, &pb.PutObjectRequest{
		Collection:          collectionId,
		Object:              object,
		Indexes:             indexes,
		AccessSecurityRules: accessRules,
	})
	if err != nil {
		return err
	}

	rsp, err := c.request(http.MethodPut, endpoint, nil, buff)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) GetObject(collectionId string, objectId string, opts objects.GetOptions) (*pb.Object, error) {
	endpoint := c.fullAPILocation() + common.ApiGetObjectRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarCollection, collectionId, 1)
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, objectId, 1)

	if opts.Info || opts.At != "" {
		values := url.Values{}
		values.Set(common.ApiParamHeader, fmt.Sprintf("%v", opts.Info))
		values.Set(common.ApiParamAt, opts.At)
		endpoint = fmt.Sprintf("%s?%s", endpoint, values.Encode())
	}

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	object := new(pb.Object)

	err = jsonpb.Unmarshal(rsp.Body, object)
	return object, err
}

func (c *Client) PatchObject(collectionId string, objectId string, patch *pb.Patch) error {
	endpoint := c.fullAPILocation() + common.ApiPatchObjectRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarCollection, collectionId, 1)
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, objectId, 1)

	buff := bytes.NewBuffer(nil)
	encoder := jsonpb.Marshaler{EnumsAsInts: true}
	err := encoder.Marshal(buff, patch)
	if err != nil {
		return err
	}

	rsp, err := c.request(http.MethodPatch, endpoint, nil, buff)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) MoveObject(collectionId string, objectId string, tartCollectionId string, accessRules *pb.PathAccessRules) error {
	endpoint := c.fullAPILocation() + common.ApiMoveObjectRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarCollection, collectionId, 1)
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, objectId, 1)

	buff := bytes.NewBuffer(nil)
	encoder := jsonpb.Marshaler{EnumsAsInts: true}
	err := encoder.Marshal(buff, &pb.MoveObjectRequest{
		TargetCollection:    tartCollectionId,
		AccessSecurityRules: accessRules,
	})
	if err != nil {
		return err
	}

	rsp, err := c.request(http.MethodPost, endpoint, nil, buff)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) ListObjects(collectionId string, opts objects.ListOptions) ([]*pb.Object, error) {
	endpoint := c.fullAPILocation() + common.ApiListObjectsRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarCollection, collectionId, 1)
	if opts.Offset > 0 || opts.At != "" {
		values := url.Values{}
		values.Set(common.ApiParamHeader, fmt.Sprintf("%d", opts.Offset))
		values.Set(common.ApiParamAt, opts.At)
		endpoint = fmt.Sprintf("%s?%s", endpoint, values.Encode())
	}

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var list []*pb.Object

	contentType := rsp.Header.Get(common.HttpHeaderContentType)
	if strings.HasPrefix(contentType, common.ContentTypeJSONStream) {
		o := new(pb.Object)
		for {
			err = jsonpb.Unmarshal(rsp.Body, o)
			if err != nil {
				if err == io.EOF {
					return list, nil
				}
				return nil, errors.Unsupported("response encoding")
			}
			list = append(list, o)
		}
	}
	return list, json.NewDecoder(rsp.Body).Decode(&list)
}

func (c *Client) SearchObjects(collectionId string, query *pb.SearchQuery) ([]*pb.Object, error) {
	endpoint := c.fullAPILocation() + common.ApiSearchObjectsRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarCollection, collectionId, 1)

	buff := bytes.NewBuffer(nil)
	encoder := jsonpb.Marshaler{EnumsAsInts: true}
	err := encoder.Marshal(buff, query)
	if err != nil {
		return nil, err
	}

	rsp, err := c.request(http.MethodPost, endpoint, nil, buff)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var list []*pb.Object
	return list, json.NewDecoder(rsp.Body).Decode(&list)
}

func (c *Client) DeleteObject(collectionId string, objectId string) error {
	endpoint := c.fullAPILocation() + common.ApiDeleteObjectRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarCollection, collectionId, 1)
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, objectId, 1)

	rsp, err := c.request(http.MethodDelete, endpoint, nil, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) CreateFileAccess(source *pb.Access) error {
	endpoint := c.fullAPILocation() + common.ApiCreateFileSource

	encoded, err := json.Marshal(source)
	if err != nil {
		return err
	}

	rsp, err := c.request(http.MethodPut, endpoint, nil, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()
	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) ListFileSources() ([]*pb.Access, error) {
	endpoint := c.fullAPILocation() + common.ApiListFileSources

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var sources []*pb.Access
	return sources, json.NewDecoder(rsp.Body).Decode(&sources)
}

func (c *Client) GetFileSource(sourceId string) (*pb.Access, error) {
	endpoint := c.fullAPILocation() + common.ApiGetFileSource
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, sourceId, 1)

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	source := new(pb.Access)
	return source, jsonpb.Unmarshal(rsp.Body, source)
}

func (c *Client) DeleteFilesSource(sourceId string) error {
	endpoint := c.fullAPILocation() + common.ApiDeleteFileSource
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, sourceId, 1)

	rsp, err := c.request(http.MethodDelete, endpoint, nil, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) CreateFile(sourceId string, file *pb.File) error {
	endpoint := c.fullAPILocation() + path.Join(common.ApiFileTreeRoutePrefix, sourceId, file.Name)

	encoded, err := json.Marshal(file)
	if err != nil {
		return err
	}

	rsp, err := c.request(http.MethodPut, endpoint, nil, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()
	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) Ls(sourceId string, dirname string, opts files.ListDirOptions) (*files.DirContent, error) {
	endpoint := c.fullAPILocation() + path.Join(common.ApiFileTreeRoutePrefix, sourceId, dirname)

	buff := bytes.NewBuffer(nil)
	err := json.NewEncoder(buff).Encode(&opts)
	if err != nil {
		return nil, err
	}

	rsp, err := c.request(http.MethodPost, endpoint, nil, buff)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var dirContent *files.DirContent
	return dirContent, json.NewDecoder(rsp.Body).Decode(&dirContent)
}

func (c *Client) GetFile(sourceId string, filename string) (*pb.File, error) {
	endpoint := c.fullAPILocation() + path.Join(common.ApiFileTreeRoutePrefix, sourceId, filename)

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	file := new(pb.File)
	return file, jsonpb.Unmarshal(rsp.Body, file)
}

func (c *Client) DeleteFile(sourceId string, filename string) error {
	endpoint := c.fullAPILocation() + path.Join(common.ApiFileTreeRoutePrefix, sourceId, filename)

	rsp, err := c.request(http.MethodDelete, endpoint, nil, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()
	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) RenameFile(sourceId string, filename string, newName string) error {
	endpoint := c.fullAPILocation() + path.Join(common.ApiFileTreeRoutePrefix, sourceId, filename)

	encoded, err := json.Marshal(&files.TreePatchInfo{Rename: true, Value: newName})
	if err != nil {
		return err
	}

	rsp, err := c.request(http.MethodPatch, endpoint, nil, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()
	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) MoveFile(sourceId string, filename string, dirname string) error {
	endpoint := c.fullAPILocation() + path.Join(common.ApiFileTreeRoutePrefix, sourceId, filename)

	encoded, err := json.Marshal(&files.TreePatchInfo{Rename: false, Value: dirname})
	if err != nil {
		return err
	}

	rsp, err := c.request(http.MethodPatch, endpoint, nil, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()
	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) DownloadFile(sourceId string, filename string) (int64, io.ReadCloser, error) {
	endpoint := c.fullAPILocation() + path.Join(common.ApiFileDataRoutePrefix, sourceId, filename)
	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return 0, nil, err
	}

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return 0, nil, err
	}
	return rsp.ContentLength, rsp.Body, nil
}

func (c *Client) UploadFile(sourceId string, filename string, content io.Reader, length int64) error {
	endpoint := c.fullAPILocation() + path.Join(common.ApiFileDataRoutePrefix, sourceId, filename)

	headers := http.Header{}
	headers.Set(common.HttpHeaderContentLength, fmt.Sprintf("%d", length))
	rsp, err := c.request(http.MethodPatch, endpoint, headers, content)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()
	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) SetFileAttributes(sourceId string, filename string, attributes files.Attributes) error {
	endpoint := c.fullAPILocation() + path.Join(common.ApiFileAttributesRoutePrefix, sourceId, filename)

	buff := bytes.NewBuffer(nil)
	err := json.NewEncoder(buff).Encode(attributes)
	if err != nil {
		return err
	}

	rsp, err := c.request(http.MethodPost, endpoint, nil, buff)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) GetFileAttributes(sourceId string, filename string) (files.Attributes, error) {
	endpoint := c.fullAPILocation() + path.Join(common.ApiFileAttributesRoutePrefix, sourceId, filename)

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var attrs files.Attributes
	err = json.NewDecoder(rsp.Body).Decode(&attrs)
	return attrs, err
}

func (c *Client) SaveAuthenticationProvider(provider *auth.Provider) error {
	endpoint := c.fullAPILocation() + common.ApiSaveAuthProviderRoute

	encoded, err := json.Marshal(provider)
	if err != nil {
		return err
	}

	rsp, err := c.request(http.MethodPut, endpoint, nil, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	defer func() {
		_ = rsp.Body.Close()
	}()
	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) GetAuthenticationProvider(providerId string) (*auth.Provider, error) {
	endpoint := c.fullAPILocation() + common.ApiGetAuthProviderRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, providerId, 1)

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var provider *auth.Provider
	return provider, json.NewDecoder(rsp.Body).Decode(&provider)
}

func (c *Client) ListAuthenticationProvider() ([]*auth.Provider, error) {
	endpoint := c.fullAPILocation() + common.ApiListAuthProvidersRoute

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var list []*auth.Provider
	return list, json.NewDecoder(rsp.Body).Decode(&list)
}

func (c *Client) DeleteAuthProvider(providerId string) error {
	endpoint := c.fullAPILocation() + common.ApiDeleteAuthProviderRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, providerId, 1)

	rsp, err := c.request(http.MethodDelete, endpoint, nil, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) CreateUserCredentials(credentials *pb.UserCredentials) error {
	endpoint := c.fullAPILocation() + common.ApiSaveAuthProviderRoute

	encoded, err := json.Marshal(credentials)
	if err != nil {
		return err
	}

	rsp, err := c.request(http.MethodPut, endpoint, nil, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	defer func() {
		_ = rsp.Body.Close()
	}()
	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) SearchUsers(pattern string) ([]string, error) {
	values := url.Values{}
	values.Set(common.ApiParamQuery, pattern)

	endpoint := c.fullAPILocation() + common.ApiSaveAuthProviderRoute
	endpoint = endpoint + "?" + values.Encode()

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var list []string
	return list, json.NewDecoder(rsp.Body).Decode(&list)
}

func (c *Client) SaveClientApplicationInfo(clientApp *pb.ClientApp) error {
	endpoint := c.fullAPILocation() + common.ApiSaveClientAppRoute

	encoded, err := json.Marshal(clientApp)
	if err != nil {
		return err
	}

	rsp, err := c.request(http.MethodPut, endpoint, nil, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	defer func() {
		_ = rsp.Body.Close()
	}()
	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) GetClientApplicationInfo(appID string) (*pb.ClientApp, error) {
	endpoint := c.fullAPILocation() + common.ApiListClientAppsRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, appID, 1)

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var clientApp *pb.ClientApp
	return clientApp, json.NewDecoder(rsp.Body).Decode(&clientApp)
}

func (c *Client) ListClientApplications() ([]*pb.ClientApp, error) {
	endpoint := c.fullAPILocation() + common.ApiListClientAppsRoute

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var list []*pb.ClientApp
	return list, json.NewDecoder(rsp.Body).Decode(&list)
}

func (c *Client) DeleteClientApplication(appID string) error {
	endpoint := c.fullAPILocation() + common.ApiDeleteClientAppRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, appID, 1)

	rsp, err := c.request(http.MethodDelete, endpoint, nil, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) CreateAccount(a *accounts.Account) error {
	endpoint := c.fullAPILocation() + common.ApiCreateAccountRoute

	encoded, err := json.Marshal(a)
	if err != nil {
		return err
	}

	rsp, err := c.request(http.MethodPut, endpoint, nil, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	defer func() {
		_ = rsp.Body.Close()
	}()
	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) GetAccount(id string) (*accounts.Account, error) {
	endpoint := c.fullAPILocation() + common.ApiGetAccountRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, id, 1)

	rsp, err := c.request(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var a *accounts.Account
	return a, json.NewDecoder(rsp.Body).Decode(&a)
}

func (c *Client) SaveSettings(name string, value string) error {
	values := url.Values{}
	values.Set(common.ApiParamName, name)

	endpoint := c.fullAPILocation() + common.ApiSetSettingsRoute
	endpoint = endpoint + "?" + values.Encode()

	rsp, err := c.request(http.MethodPut, endpoint, nil, strings.NewReader(value))
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) GetSettings(name string) (string, error) {
	values := url.Values{}
	values.Set(common.ApiParamName, name)

	endpoint := c.fullAPILocation() + common.ApiSetSettingsRoute
	endpoint = endpoint + "?" + values.Encode()

	rsp, err := c.request(http.MethodPut, endpoint, nil, nil)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return "", err
	}

	valueBytes, err := io.ReadAll(rsp.Body)
	return string(valueBytes), err
}
