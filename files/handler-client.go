package files

import (
	"context"
	pb "github.com/omecodes/store/gen/go/proto"
	"io"
)

func NewHandlerServiceClient(clientType uint32) Handler {
	return &ServiceClientHandler{
		clientType: clientType,
	}
}

type ServiceClientHandler struct {
	BaseHandler
	clientType uint32
}

func (h *ServiceClientHandler) CreateAccess(ctx context.Context, source *pb.Access) error {
	client, err := NewSourcesServiceClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.CreateAccess(ctx, &pb.CreateAccessRequest{Access: source})
	return err
}

func (h *ServiceClientHandler) GetAccessList(ctx context.Context) ([]*pb.Access, error) {
	client, err := NewSourcesServiceClient(ctx, h.clientType)
	if err != nil {
		return nil, err
	}

	stream, err := client.GetAccessList(ctx, &pb.GetAccessListRequest{})
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = stream.CloseSend()
	}()

	var sources []*pb.Access

	done := false
	for !done {
		source, err := stream.Recv()
		if err != nil {
			if done = err == io.EOF; !done {
				return nil, err
			}
		}
		sources = append(sources, source)
	}

	return sources, nil
}

func (h *ServiceClientHandler) GetAccess(ctx context.Context, accessID string) (*pb.Access, error) {
	client, err := NewSourcesServiceClient(ctx, h.clientType)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetAccess(ctx, &pb.GetAccessRequest{Id: accessID})
	if err != nil {
		return nil, err
	}

	return rsp.Access, nil
}

func (h *ServiceClientHandler) DeleteAccess(ctx context.Context, accessID string) error {
	client, err := NewSourcesServiceClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	stream, err := client.DeleteAccess(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = stream.CloseSend()
	}()

	return stream.Send(&pb.DeleteAccessRequest{AccessId: accessID})
}

func (h *ServiceClientHandler) CreateDir(ctx context.Context, accessID string, filename string) error {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.CreateDir(ctx, &pb.CreateDirRequest{
		AccessId: accessID,
		Path:     filename,
	})
	return err
}

func (h *ServiceClientHandler) WriteFileContent(ctx context.Context, accessID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	client, err := NewTransfersServiceClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.WriteFile(ctx, &WriteRequest{
		AccessID: accessID,
		Path:     filename,
		Data:     content,
		Length:   size,
		Append:   opts.Append,
	})

	return err
}

func (h *ServiceClientHandler) ReadFileContent(ctx context.Context, accessID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	client, err := NewTransfersServiceClient(ctx, h.clientType)
	if err != nil {
		return nil, 0, err
	}

	rsp, err := client.ReadFile(ctx, &ReadRequest{
		AccessID: accessID,
		Path:     filename,
		Offset:   opts.Range.Offset,
		Length:   opts.Range.Length,
	})
	if err != nil {
		return nil, 0, err
	}

	return rsp.Data, rsp.Length, nil
}

func (h *ServiceClientHandler) ListDir(ctx context.Context, accessID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return nil, err
	}

	rsp, err := client.ListDir(ctx, &pb.ListDirRequest{
		AccessId: accessID,
		Path:     dirname,
	})
	if err != nil {
		return nil, err
	}

	return &DirContent{
		Files:  rsp.Files,
		Total:  int(rsp.Total),
		Offset: int(rsp.Offset),
	}, err
}

func (h *ServiceClientHandler) GetFileInfo(ctx context.Context, accessID string, filename string, opts GetFileOptions) (*pb.File, error) {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetFile(ctx, &pb.GetFileRequest{
		AccessId:       accessID,
		Path:           filename,
		WithAttributes: opts.WithAttrs,
	})
	if err != nil {
		return nil, err
	}

	return rsp.File, err
}

func (h *ServiceClientHandler) DeleteFile(ctx context.Context, accessID string, filename string, opts DeleteFileOptions) error {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.DeleteFile(ctx, &pb.DeleteFileRequest{
		AccessId: accessID,
		Path:     filename,
	})
	return err
}

func (h *ServiceClientHandler) SetFileAttributes(ctx context.Context, accessID string, filename string, attrs Attributes) error {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.SetFileAttributes(ctx, &pb.SetFileAttributesRequest{
		AccessId:   accessID,
		Path:       filename,
		Attributes: attrs,
	})
	return err
}

func (h *ServiceClientHandler) GetFileAttributes(ctx context.Context, accessID string, filename string, name ...string) (Attributes, error) {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetFileAttributes(ctx, &pb.GetFileAttributesRequest{
		AccessId: accessID,
		Path:     filename,
		Names:    name,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Attributes, nil
}

func (h *ServiceClientHandler) RenameFile(ctx context.Context, accessID string, filename string, newName string) error {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.RenameFile(ctx, &pb.RenameFileRequest{
		AccessId: accessID,
		Path:     filename,
		NewName:  newName,
	})
	return err
}

func (h *ServiceClientHandler) MoveFile(ctx context.Context, accessID string, filename string, dirname string) error {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.MoveFile(ctx, &pb.MoveFileRequest{
		AccessId:  accessID,
		Path:      filename,
		TargetDir: dirname,
	})
	return err
}

func (h *ServiceClientHandler) CopyFile(ctx context.Context, accessID string, filename string, dirname string) error {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.CopyFile(ctx, &pb.CopyFileRequest{
		AccessId:  accessID,
		Path:      filename,
		TargetDir: dirname,
	})
	return err
}

func (h *ServiceClientHandler) OpenMultipartSession(ctx context.Context, accessID string, filename string, info MultipartSessionInfo) (string, error) {
	client, err := NewTransfersServiceClient(ctx, h.clientType)
	if err != nil {
		return "", err
	}

	rsp, err := client.OpenFileMultipartWriteSession(ctx, &OpenMultipartSessionRequest{
		AccessID: accessID,
		Path:     filename,
	})

	if err != nil {
		return "", err
	}
	return rsp.SessionId, nil
}

func (h *ServiceClientHandler) WriteFilePart(ctx context.Context, sessionID string, content io.Reader, size int64, info ContentPartInfo) (int64, error) {
	client, err := NewTransfersServiceClient(ctx, h.clientType)
	if err != nil {
		return 0, err
	}

	rsp, err := client.WriteFilePart(ctx, &WriteFilePartRequest{
		SessionId: sessionID,
		Data:      content,
		Length:    size,
	})
	if err != nil {
		return 0, err
	}

	return rsp.Written, nil
}

func (h *ServiceClientHandler) CloseMultipartSession(ctx context.Context, sessionId string) error {
	client, err := NewTransfersServiceClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.CloseFileMultipartWriteSession(ctx, &CloseMultipartWriteSessionRequest{
		SessionId: sessionId,
	})
	return err
}
