package operations

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

type DocClient interface {
	Export(ctx context.Context, cfg *DocumentExport) error
}

// DocumentExport holds the configuration and result for a integration document export
type DocumentExport struct {
	// FileID is the external file identifier to export
	FileID string `json:"fileId"`
	// HTML is the exported document content as an embeddable iframe string (populated in the response)
	HTML string `json:"html,omitempty"`
	// PDF is the exported document content as raw PDF bytes (populated in the response)
	PDF []byte `json:"pdf,omitempty"`
	// MimeType is the content type of the downloaded file
	MimeType string `json:"mimeType,omitempty"`
	// Name is the file name without extension
	Name string `json:"name,omitempty"`
}

// Handle adapts the document export to the generic operation registration boundary
func Handle[C DocClient](ref types.ClientRef[C], op types.OperationRef[DocumentExport]) types.OperationHandler {
	return providerkit.WithClientConfig(ref, op, ErrExportFailed, func(ctx context.Context, svc C, cfg DocumentExport) (json.RawMessage, error) {
		return Run(ctx, svc, &cfg)
	})
}

// Run executes the HTML export using the Google Drive API files.export endpoint
func Run(ctx context.Context, svc DocClient, cfg *DocumentExport) (json.RawMessage, error) {
	if cfg == nil || cfg.FileID == "" {
		return nil, ErrExportFailed
	}

	err := svc.Export(ctx, cfg)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("Failed to create HTML export of file")

		return nil, ErrExportFailed
	}

	return providerkit.EncodeResult(*cfg, ErrResultEncode)
}
