package objectstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

// Handle adapts the object write to the generic operation registration boundary
func (WriteObject) Handle() types.OperationHandler {
	return providerkit.WithClientConfig(storageClient, writeObjectOperation, ErrOperationConfigInvalid, func(ctx context.Context, client *Client, cfg WriteObject) (json.RawMessage, error) {
		result, err := cfg.Run(ctx, client.Provider)
		if err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(result, ErrResultEncode)
	})
}

// Run encodes the document and uploads it to the bucket under the configured key
func (w WriteObject) Run(ctx context.Context, client storage.Provider) (WriteObjectResult, error) {
	body, err := w.encode()
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("key", w.Key).Msg("objectstore: document could not be encoded")

		return WriteObjectResult{}, err
	}

	contentType := w.ContentType
	if contentType == "" {
		contentType = defaultContentType
	}

	uploaded, err := client.Upload(ctx, bytes.NewReader(body), &storagetypes.UploadFileOptions{
		FileName:    w.Key,
		ContentType: contentType,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("key", w.Key).Msg("objectstore: object upload failed")

		return WriteObjectResult{}, fmt.Errorf("%w: %w", ErrObjectUploadFailed, err)
	}

	return WriteObjectResult{
		Bucket:      uploaded.Bucket,
		Key:         uploaded.Key,
		ContentType: uploaded.ContentType,
		Size:        uploaded.Size,
		URI:         uploaded.FullURI,
	}, nil
}

// encode returns the document bytes, writing strings verbatim and JSON encoding any other value
func (w WriteObject) encode() ([]byte, error) {
	switch v := w.Content.(type) {
	case string:
		return []byte(v), nil
	default:
		body, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrContentEncode, err)
		}

		return body, nil
	}
}
