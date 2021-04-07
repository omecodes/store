package files

import (
	"context"
	"fmt"
	"github.com/omecodes/errors"
	ome "github.com/omecodes/libome"
	"github.com/omecodes/libome/logs"
	"github.com/omecodes/service"
	"github.com/omecodes/store/common"
	"io"
	"net/http"
	"path"
	"sync"
)

// NewTransfersServiceClient is a source service client constructor
func NewTransfersServiceClient(ctx context.Context, serviceType uint32) (TransferClient, error) {
	provider := GetTransfersServiceClientProvider(ctx)
	if provider == nil {
		return nil, errors.ServiceUnavailable("no service available", errors.Details{Key: "type", Value: "service"}, errors.Details{Key: "service-type", Value: serviceType})
	}
	return provider.GetClient(ctx, serviceType)
}

type TransferClient interface {
	ReadFile(ctx context.Context, request *ReadRequest) (*ReadResponse, error)
	WriteFile(ctx context.Context, request *WriteRequest) (*WriteResponse, error)
	OpenFileMultipartWriteSession(ctx context.Context, request *OpenMultipartSessionRequest) (*OpenMultipartSessionResponse, error)
	WriteFilePart(ctx context.Context, request *WriteFilePartRequest) (*WritePartResponse, error)
	CloseFileMultipartWriteSession(ctx context.Context, request *CloseMultipartWriteSessionRequest) (*CloseMultipartWriteSessionResponse, error)
}

type ReadRequest struct {
	SourceID string
	Path     string
	Offset   int64
	Length   int64
}
type ReadResponse struct {
	Data   io.ReadCloser
	Length int64
}

type WriteRequest struct {
	SourceID string
	Path     string
	Data     io.Reader
	Length   int64
	Append   bool
}
type WriteResponse struct {
	Written int64
}

type OpenMultipartSessionRequest struct {
	SourceId string
	Path     string
}
type OpenMultipartSessionResponse struct {
	SessionId string
}

type WriteFilePartRequest struct {
	SessionId string
	Data      io.Reader
	Length    int64
}
type WritePartResponse struct {
	Written int64
}

type CloseMultipartWriteSessionRequest struct {
	SessionId string
}
type CloseMultipartWriteSessionResponse struct{}

type TransfersServiceClientProvider interface {
	GetClient(ctx context.Context, serviceType uint32) (TransferClient, error)
}

type DefaultTransfersServiceClientProvider struct {
	sync.RWMutex
	balanceIndex int
}

func (p *DefaultTransfersServiceClientProvider) incrementBalanceIndex() {
	p.Lock()
	defer p.Unlock()
	p.balanceIndex++
}

func (p *DefaultTransfersServiceClientProvider) getBalanceIndex() int {
	p.RLock()
	defer p.Unlock()
	return p.balanceIndex
}

func (p *DefaultTransfersServiceClientProvider) GetClient(ctx context.Context, serviceType uint32) (TransferClient, error) {
	infoList, err := service.GetRegistry(ctx).GetOfType(serviceType)
	if err != nil {
		return nil, err
	}

	if len(infoList) == 0 {
		return nil, errors.ServiceUnavailable("could not find service ", errors.Details{Key: "type", Value: serviceType})
	}

	if len(infoList) == 1 {
		return p.getNodeClient(ctx, infoList[0])
	}

	defer p.incrementBalanceIndex()
	balanceIndex := p.getBalanceIndex()
	lastBalanceIndex := balanceIndex % len(infoList)

	for i := p.balanceIndex + 1; i%len(infoList) != lastBalanceIndex; i++ {
		info := infoList[p.balanceIndex%len(infoList)]
		client, err := p.getNodeClient(ctx, info)
		if err != nil {
			logs.Error("could not connect to service", logs.Err(err))
			continue
		}
		return client, nil
	}
	return nil, errors.ServiceUnavailable("could not find service ", errors.Details{Key: "type", Value: serviceType})
}

func (p *DefaultTransfersServiceClientProvider) getNodeClient(ctx context.Context, info *ome.ServiceInfo) (TransferClient, error) {
	for _, node := range info.Nodes {
		if node.Protocol == ome.Protocol_Http {
			client := &defaultTransfersServiceClient{}

			tlsConfig, err := service.GetClientTLSConfig(ctx)
			if err != nil {
				return nil, err
			}

			hc := &http.Client{Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			}}

			client.client = hc
			client.apiPathPrefix, _ = node.Meta[common.ServiceMetaAPIPath]
			client.address = node.Address
			return client, nil
		}
	}
	return nil, errors.ServiceUnavailable("could not find service for protocol", errors.Details{Key: "protocol", Value: "HTTP"})
}

type defaultTransfersServiceClient struct {
	client        *http.Client
	apiPathPrefix string
	address       string
}

func (d *defaultTransfersServiceClient) ReadFile(ctx context.Context, request *ReadRequest) (*ReadResponse, error) {
	downloadURL := fmt.Sprintf("https://%s/%s", d.address, path.Join(d.apiPathPrefix, "/data", request.SourceID, request.Path))
	rsp, err := d.client.Get(downloadURL)
	if err != nil {
		return nil, err
	}

	rr := &ReadResponse{
		Data:   rsp.Body,
		Length: rsp.ContentLength,
	}
	return rr, nil
}

func (d *defaultTransfersServiceClient) WriteFile(ctx context.Context, request *WriteRequest) (*WriteResponse, error) {
	uploadURL := fmt.Sprintf("https://%s/%s", d.address, path.Join(d.apiPathPrefix, "/data", request.SourceID, request.Path))

	req, err := http.NewRequest(http.MethodPut, uploadURL, request.Data)
	if err != nil {
		return nil, err
	}

	_, err = d.client.Do(req)

	return &WriteResponse{
		Written: req.ContentLength,
	}, err
}

func (d *defaultTransfersServiceClient) OpenFileMultipartWriteSession(ctx context.Context, request *OpenMultipartSessionRequest) (*OpenMultipartSessionResponse, error) {
	return nil, errors.UnImplemented("")
}

func (d *defaultTransfersServiceClient) WriteFilePart(ctx context.Context, request *WriteFilePartRequest) (*WritePartResponse, error) {
	return nil, errors.UnImplemented("")
}

func (d *defaultTransfersServiceClient) CloseFileMultipartWriteSession(ctx context.Context, request *CloseMultipartWriteSessionRequest) (*CloseMultipartWriteSessionResponse, error) {
	return nil, errors.UnImplemented("")
}
