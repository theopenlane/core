package googledrive

import (
	"context"
	"encoding/json"
	"io"

	"google.golang.org/api/drive/v3"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const exportMIMEType = "text/html"

// DocumentExport holds the configuration and result for a Google Drive document export
type DocumentExport struct {
	// FileID is the Google Drive file identifier to export
	FileID string `json:"fileId"`
	// HTML is the exported document content (populated in the response)
	HTML string `json:"html,omitempty"`
}

// Handle adapts the document export to the generic operation registration boundary
func (d DocumentExport) Handle() types.OperationHandler {
	return providerkit.WithClientConfig(driveClient, documentExportOperation, ErrExportFailed, d.Run)
}

// Run executes the HTML export using the Google Drive API files.export endpoint
func (DocumentExport) Run(ctx context.Context, svc *drive.Service, cfg DocumentExport) (json.RawMessage, error) {
	if cfg.FileID == "" {
		return nil, ErrExportFailed
	}

	resp, err := svc.Files.Export(cfg.FileID, exportMIMEType).Context(ctx).Download()
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("Failed to create HTML export of drive file")

		return nil, ErrExportFailed
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("Failed to parse exported drive file contents")

		return nil, ErrExportFailed
	}

	return providerkit.EncodeResult(DocumentExport{
		FileID: cfg.FileID,
		HTML:   string(body),
	}, ErrResultEncode)
}
