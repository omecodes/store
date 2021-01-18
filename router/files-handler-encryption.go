package router

import (
	"context"
	"github.com/omecodes/errors"
	"github.com/omecodes/libome/crypt"
	"github.com/omecodes/store/files"
	"github.com/omecodes/store/pb"
	"io"
)

type FilesEncryptionHandler struct {
	FilesBaseHandler
}

func (h *FilesEncryptionHandler) WriteFileContent(ctx context.Context, filename string, content io.Reader, size int64, opts pb.PutFileOptions) error {
	source := files.GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	if source.Encryption == nil {
		return h.next.WriteFileContent(ctx, filename, content, size, opts)
	}

	encryptStream := crypt.NewEncryptWrapper(nil, crypt.WithBlockSize(4096))
	return h.next.WriteFileContent(ctx, filename, encryptStream.WrapReader(content), size, opts)
}

func (h *FilesEncryptionHandler) ReadFileContent(ctx context.Context, filename string, opts pb.GetFileOptions) (io.ReadCloser, int64, error) {
	source := files.GetSource(ctx)
	if source == nil {
		return nil, 0, errors.Create(errors.Internal, "missing source in context")
	}

	if source.Encryption == nil {
		return h.next.ReadFileContent(ctx, filename, opts)
	}

	readCloser, size, err := h.next.ReadFileContent(ctx, filename, opts)
	if err != nil {
		return nil, 0, err
	}

	decryptStream := crypt.NewDecryptWrapper(nil, crypt.WithLimit(opts.Range.Length), crypt.WithOffset(opts.Range.Offset))
	return decryptStream.WrapReadCloser(readCloser), size, nil
}

func (h *FilesEncryptionHandler) AddContentPart(ctx context.Context, sessionID string, content io.Reader, size int64, info *pb.ContentPartInfo) error {
	source := files.GetSource(ctx)
	if source == nil {
		return errors.Create(errors.Internal, "missing source in context")
	}

	if source.Encryption == nil {
		return h.next.AddContentPart(ctx, sessionID, content, size, info)
	}

	encryptStream := crypt.NewEncryptWrapper(nil, crypt.WithBlockSize(4096))
	return h.next.AddContentPart(ctx, sessionID, encryptStream.WrapReader(content), size, info)
}
