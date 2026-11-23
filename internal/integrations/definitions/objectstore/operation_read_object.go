package objectstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/logx"
)

// Run reads the object under the key, reporting a missing object as not found rather than as an error
func (r ReadObject) Run(ctx context.Context, client storage.Provider) (ReadObjectResult, error) {
	exists, err := client.Exists(ctx, &storagetypes.File{Key: r.Key})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("key", r.Key).Msg("objectstore: object exists check failed")

		return ReadObjectResult{}, err
	}

	if !exists {
		return ReadObjectResult{}, nil
	}

	// the object can still be removed between the exists check and the download
	downloaded, err := client.Download(ctx, &storagetypes.File{Key: r.Key}, &storagetypes.DownloadFileOptions{})
	if err != nil {
		if objectMissing(err) {
			return ReadObjectResult{}, nil
		}

		logx.FromContext(ctx).Error().Err(err).Str("key", r.Key).Msg("objectstore: object download failed")

		return ReadObjectResult{}, fmt.Errorf("%w: %s: %w", ErrObjectDownloadFailed, r.Key, err)
	}

	document := downloaded.File

	if !json.Valid(document) {
		content, err := json.Marshal(string(document))
		if err != nil {
			return ReadObjectResult{}, err
		}

		return ReadObjectResult{Found: true, Content: content}, nil
	}

	return ReadObjectResult{Found: true, Content: document}, nil
}

// runReadObject reads the document under the configured key from the installation's bucket
func runReadObject(ctx context.Context, _ types.OperationRequest, client *Client, cfg ReadObject) (json.RawMessage, error) {
	result, err := cfg.Run(ctx, client.Provider)
	if err != nil {
		return nil, err
	}

	return providerkit.EncodeResult(result, ErrResultEncode)
}

// objectMissing reports whether a provider error means the object does not exist
func objectMissing(err error) bool {
	return errors.Is(err, storage.ErrObjectNotFound) || errors.Is(err, fs.ErrNotExist)
}
