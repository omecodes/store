package files

import (
	"context"
	"io"

	"github.com/omecodes/libome/crypt"
)

type EncryptionHandler struct {
	BaseHandler
}

func (h *EncryptionHandler) WriteFileContent(ctx context.Context, sourceID string, filename string, content io.Reader, size int64, opts WriteOptions) error {
	source, err := h.next.GetAccess(ctx, sourceID)
	if err != nil {
		return err
	}

	if source.Encryption == nil {
		return h.next.WriteFileContent(ctx, sourceID, filename, content, size, opts)
	}

	encryptStream := crypt.NewEncryptWrapper(nil, crypt.WithBlockSize(4096))
	return h.next.WriteFileContent(ctx, sourceID, filename, encryptStream.WrapReader(content), size, opts)
}

func (h *EncryptionHandler) ReadFileContent(ctx context.Context, sourceID string, filename string, opts ReadOptions) (io.ReadCloser, int64, error) {
	source, err := h.next.GetAccess(ctx, sourceID)
	if err != nil {
		return nil, 0, err
	}

	if source.Encryption == nil {
		return h.next.ReadFileContent(ctx, sourceID, filename, opts)
	}

	readCloser, size, err := h.next.ReadFileContent(ctx, sourceID, filename, opts)
	if err != nil {
		return nil, 0, err
	}

	decryptStream := crypt.NewDecryptWrapper(nil, crypt.WithLimit(opts.Range.Length), crypt.WithOffset(opts.Range.Offset))
	return decryptStream.WrapReadCloser(readCloser), size, nil
}

func (h *EncryptionHandler) WriteFilePart(ctx context.Context, sessionID string, content io.Reader, size int64, info ContentPartInfo) (int64, error) {
	source, err := h.next.GetAccess(ctx, sessionID)
	if err != nil {
		return 0, err
	}

	if source.Encryption == nil {
		return h.next.WriteFilePart(ctx, sessionID, content, size, info)
	}

	encryptStream := crypt.NewEncryptWrapper(nil, crypt.WithBlockSize(4096))
	return h.next.WriteFilePart(ctx, sessionID, encryptStream.WrapReader(content), size, info)
}
