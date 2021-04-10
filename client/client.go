package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/omecodes/errors"
	"github.com/omecodes/store/common"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/objects"
	se "github.com/omecodes/store/search-engine"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func New(server string, opts ...Option) *Client {
	c := &Client{
		Server: server,
	}
	c.options.apiLocation = "/api"
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
	Server     string
	options    options
	httpClient *http.Client
}

func (c *Client) fullAPILocation() string {
	if c.options.noTLS {
		return fmt.Sprintf("http://localhost:%d%s", c.options.port, path.Join("/", c.options.apiLocation))
	}
	return fmt.Sprintf("https://localhost:%d%s", c.options.port, path.Join("/", c.options.apiLocation))
}

func (c *Client) request(method string, endpoint string, headers http.Header, body io.Reader) (*http.Response, error) {
	if !strings.HasPrefix(endpoint, "http") {
		if c.options.noTLS {
			endpoint = "http://" + endpoint
		} else {
			endpoint = "https://" + endpoint
		}
	}

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

func (c *Client) CreateObjectsCollection(collection *objects.Collection) error {
	endpoint := fmt.Sprintf(c.fullAPILocation() + common.ApiCreateCollectionRoute)

	encoded, err := json.Marshal(collection)
	if err != nil {
		return err
	}

	headers := http.Header{}
	headers.Set(common.HttpHeaderAccept, common.AllJSONContentTypes)
	rsp, err := c.request(http.MethodPut, endpoint, headers, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	return common.ErrorFromHttpResponse(rsp)
}

func (c *Client) ListCollections() ([]*objects.Collection, error) {
	endpoint := fmt.Sprintf(c.fullAPILocation() + common.ApiListCollectionRoute)

	headers := http.Header{}
	headers.Set(common.HttpHeaderAccept, common.AllJSONContentTypes)

	rsp, err := c.request(http.MethodGet, endpoint, headers, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()
	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var collections []*objects.Collection

	contentType := rsp.Header.Get(common.HttpHeaderContentType)
	if strings.HasPrefix(contentType, common.ContentTypeJSONStream) {
		col := new(objects.Collection)
		for {
			err = jsonpb.Unmarshal(rsp.Body, col)
			if err != nil {
				if err == io.EOF {
					return collections, nil
				}
				return nil, errors.Unsupported("response encoding")
			}
			collections = append(collections, col)
		}
	}

	return collections, json.NewDecoder(rsp.Body).Decode(&collections)
}

func (c *Client) GetCollection(collectionId string) (*objects.Collection, error) {
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

	col := new(objects.Collection)
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

func (c *Client) PutObject(collectionId string, object *objects.Object, accessRules *objects.PathAccessRules, indexes ...*se.TextIndex) error {
	endpoint := c.fullAPILocation() + common.ApiPutObjectRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarCollection, collectionId, 1)

	buff := bytes.NewBuffer(nil)
	encoder := jsonpb.Marshaler{EnumsAsInts: true}
	err := encoder.Marshal(buff, &objects.PutObjectRequest{
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

func (c *Client) GetObject(collectionId string, objectId string, opts objects.GetOptions) (*objects.Object, error) {
	endpoint := c.fullAPILocation() + common.ApiGetObjectRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarCollection, collectionId, 1)
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, objectId, 1)

	if opts.Info || opts.At != "" {
		values := url.Values{}
		values.Set(common.ApiQueryParamHeader, fmt.Sprintf("%v", opts.Info))
		values.Set(common.ApiQueryParamAt, opts.At)
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

	object := new(objects.Object)

	err = jsonpb.Unmarshal(rsp.Body, object)
	return object, err
}

func (c *Client) PatchObject(collectionId string, objectId string, patch *objects.Patch) error {
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

func (c *Client) MoveObject(collectionId string, objectId string, tartCollectionId string, accessRules *objects.PathAccessRules) error {
	endpoint := c.fullAPILocation() + common.ApiMoveObjectRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarCollection, collectionId, 1)
	endpoint = strings.Replace(endpoint, common.ApiRouteVarId, objectId, 1)

	buff := bytes.NewBuffer(nil)
	encoder := jsonpb.Marshaler{EnumsAsInts: true}
	err := encoder.Marshal(buff, &objects.MoveObjectRequest{
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

func (c *Client) ListObjects(collectionId string, opts objects.ListOptions) ([]*objects.Object, error) {
	endpoint := c.fullAPILocation() + common.ApiListObjectsRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarCollection, collectionId, 1)
	if opts.Offset > 0 || opts.At != "" {
		values := url.Values{}
		values.Set(common.ApiQueryParamHeader, fmt.Sprintf("%d", opts.Offset))
		values.Set(common.ApiQueryParamAt, opts.At)
		endpoint = fmt.Sprintf("%s?%s", endpoint, values.Encode())
	}

	headers := http.Header{}
	headers.Set(common.HttpHeaderAccept, common.AllJSONContentTypes)

	rsp, err := c.request(http.MethodGet, endpoint, headers, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var list []*objects.Object

	contentType := rsp.Header.Get(common.HttpHeaderContentType)
	if strings.HasPrefix(contentType, common.ContentTypeJSONStream) {
		o := new(objects.Object)
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

func (c *Client) SearchObjects(collectionId string, query *se.SearchQuery) ([]*objects.Object, error) {
	endpoint := c.fullAPILocation() + common.ApiSearchObjectsRoute
	endpoint = strings.Replace(endpoint, common.ApiRouteVarCollection, collectionId, 1)

	headers := http.Header{}
	headers.Set(common.HttpHeaderAccept, common.AllJSONContentTypes)

	buff := bytes.NewBuffer(nil)
	encoder := jsonpb.Marshaler{EnumsAsInts: true}
	err := encoder.Marshal(buff, query)
	if err != nil {
		return nil, err
	}

	rsp, err := c.request(http.MethodPost, endpoint, headers, buff)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var list []*objects.Object

	contentType := rsp.Header.Get(common.HttpHeaderContentType)
	if strings.HasPrefix(contentType, common.ContentTypeJSONStream) {
		o := new(objects.Object)
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

func (c *Client) CreateFileSource(source *files.Source) error {
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

func (c *Client) ListFileSources() ([]*files.Source, error) {
	endpoint := c.fullAPILocation() + common.ApiListFileSources

	headers := http.Header{}
	headers.Set(common.HttpHeaderAccept, common.AllJSONContentTypes)

	rsp, err := c.request(http.MethodGet, endpoint, headers, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if err = common.ErrorFromHttpResponse(rsp); err != nil {
		return nil, err
	}

	var sources []*files.Source
	contentType := rsp.Header.Get(common.HttpHeaderContentType)

	if strings.HasPrefix(contentType, common.ContentTypeJSONStream) {
		source := new(files.Source)
		for {
			err = jsonpb.Unmarshal(rsp.Body, source)
			if err != nil {
				if err == io.EOF {
					return sources, nil
				}
				return nil, errors.Unsupported("response encoding")
			}
			sources = append(sources, source)
		}
	}
	return sources, json.NewDecoder(rsp.Body).Decode(&sources)
}

func (c *Client) GetFileSource(sourceId string) (*files.Source, error) {
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

	source := new(files.Source)
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

func (c *Client) CreateFile(sourceId string, file *files.File) error {
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

	headers := http.Header{}
	headers.Set(common.HttpHeaderAccept, common.ContentTypeJSON)

	rsp, err := c.request(http.MethodGet, endpoint, headers, nil)
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

func (c *Client) GetFile(sourceId string, filename string) (*files.File, error) {
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

	file := new(files.File)
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
