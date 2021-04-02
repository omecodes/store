package files

import (
	"context"
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

func (h *ServiceClientHandler) CreateSource(ctx context.Context, source *Source) error {
	client, err := NewSourcesServiceClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.CreateSource(ctx, &CreateSourceRequest{Source: source})
	return err
}

func (h *ServiceClientHandler) ListSources(ctx context.Context) ([]*Source, error) {
	client, err := NewSourcesServiceClient(ctx, h.clientType)
	if err != nil {
		return nil, err
	}

	stream, err := client.GetSources(ctx, &GetSourcesRequest{})
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = stream.CloseSend()
	}()

	var sources []*Source

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

func (h *ServiceClientHandler) GetSource(ctx context.Context, sourceID string) (*Source, error) {
	client, err := NewSourcesServiceClient(ctx, h.clientType)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetSource(ctx, &GetSourceRequest{Id: sourceID})
	if err != nil {
		return nil, err
	}

	return rsp.Source, nil
}

func (h *ServiceClientHandler) DeleteSource(ctx context.Context, sourceID string) error {
	client, err := NewSourcesServiceClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	stream, err := client.DeleteSource(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = stream.CloseSend()
	}()

	return stream.Send(&DeleteSourceRequest{SourceId: sourceID})
}

func (h *ServiceClientHandler) CreateDir(ctx context.Context, sourceID string, filename string) error {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.CreateDir(ctx, &CreateDirRequest{
		SourceId: sourceID,
		Path:     filename,
	})
	return err
}

func (h *ServiceClientHandler) WriteFileContent(ctx context.Context, sourceID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	client, err := NewTransfersServiceClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.WriteFile(ctx, &WriteRequest{
		SourceID: sourceID,
		Path:     filename,
		Data:     content,
		Length:   size,
		Append:   opts.Append,
	})

	return err
}

func (h *ServiceClientHandler) ReadFileContent(ctx context.Context, sourceID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	client, err := NewTransfersServiceClient(ctx, h.clientType)
	if err != nil {
		return nil, 0, err
	}

	rsp, err := client.ReadFile(ctx, &ReadRequest{
		SourceID: sourceID,
		Path:     filename,
		Offset:   opts.Range.Offset,
		Length:   opts.Range.Length,
	})
	if err != nil {
		return nil, 0, err
	}

	return rsp.Data, rsp.Length, nil
}

func (h *ServiceClientHandler) ListDir(ctx context.Context, sourceID string, dirname string, opts ListDirOptions) (*DirContent, error) {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return nil, err
	}

	rsp, err := client.ListDir(ctx, &ListDirRequest{
		SourceId: sourceID,
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

func (h *ServiceClientHandler) GetFileInfo(ctx context.Context, sourceID string, filename string, opts GetFileOptions) (*File, error) {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetFile(ctx, &GetFileRequest{
		SourceId:       sourceID,
		Path:           filename,
		WithAttributes: opts.WithAttrs,
	})
	if err != nil {
		return nil, err
	}

	return rsp.File, err
}

func (h *ServiceClientHandler) DeleteFile(ctx context.Context, sourceID string, filename string, opts DeleteFileOptions) error {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.DeleteFile(ctx, &DeleteFileRequest{
		SourceId: sourceID,
		Path:     filename,
	})
	return err
}

func (h *ServiceClientHandler) SetFileAttributes(ctx context.Context, sourceID string, filename string, attrs Attributes) error {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.SetFileAttributes(ctx, &SetFileAttributesRequest{
		SourceId:   sourceID,
		Path:       filename,
		Attributes: attrs,
	})
	return err
}

func (h *ServiceClientHandler) GetFileAttributes(ctx context.Context, sourceID string, filename string, name ...string) (Attributes, error) {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return nil, err
	}

	rsp, err := client.GetFileAttributes(ctx, &GetFileAttributesRequest{
		SourceId: sourceID,
		Path:     filename,
		Names:    name,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Attributes, nil
}

func (h *ServiceClientHandler) RenameFile(ctx context.Context, sourceID string, filename string, newName string) error {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.RenameFile(ctx, &RenameFileRequest{
		SourceId: sourceID,
		Path:     filename,
		NewName:  newName,
	})
	return err
}

func (h *ServiceClientHandler) MoveFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.MoveFile(ctx, &MoveFileRequest{
		SourceId:  sourceID,
		Path:      filename,
		TargetDir: dirname,
	})
	return err
}

func (h *ServiceClientHandler) CopyFile(ctx context.Context, sourceID string, filename string, dirname string) error {
	client, err := NewClient(ctx, h.clientType)
	if err != nil {
		return err
	}

	_, err = client.CopyFile(ctx, &CopyFileRequest{
		SourceId:  sourceID,
		Path:      filename,
		TargetDir: dirname,
	})
	return err
}

func (h *ServiceClientHandler) OpenMultipartSession(ctx context.Context, sourceID string, filename string, info MultipartSessionInfo) (string, error) {
	client, err := NewTransfersServiceClient(ctx, h.clientType)
	if err != nil {
		return "", err
	}

	rsp, err := client.OpenFileMultipartWriteSession(ctx, &OpenMultipartSessionRequest{
		SourceId: sourceID,
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
