package files

import (
	"context"
	"github.com/omecodes/service"
	"github.com/omecodes/store/micro"
)

type ReadRequest struct{}
type ReadResponse struct{}

type WriteRequest struct{}
type WriteResponse struct{}

type OpenMultipartSessionRequest struct{}
type OpenMultipartSessionResponse struct{}

type WritePartRequest struct{}
type WritePartResponse struct{}

type CloseMultipartWriteSessionRequest struct{}
type CloseMultipartWriteSessionResponse struct{}

type TransferNodeClient interface {
	ReadFile(ctx context.Context, request *ReadRequest) (*ReadResponse, error)
	WriteFile(ctx context.Context, request *WriteRequest) (*WriteResponse, error)
	OpenFileMultipartWriteSession(ctx context.Context, request *WriteRequest) (*WriteResponse, error)
	WriteFilePart(ctx context.Context, request *WriteRequest) (*WriteResponse, error)
	CloseFileMultipartWriteSession(ctx context.Context, request *CloseMultipartWriteSessionRequest) (*CloseMultipartWriteSessionResponse, error)
}

type NodeClient interface {
	FilesClient
	SourcesClient
	TransferNodeClient
}

type ServiceNodeClient struct{}

type ServiceNodeServer struct {
	params *micro.ServiceConfig
	box    *service.Box
}
