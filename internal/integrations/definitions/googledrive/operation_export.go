package googledrive

import (
	"context"
	"io"

	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const exportMIMEType = "text/html"

// Handle adapts the document export to the generic operation registration boundary
func Handle() types.OperationHandler {
	return operations.Handle(driveClient, documentExportOperation)
}

// Export fetches OneDrive item metadata and returns either an iframe embed or PDF bytes
func (c DriveClient) Export(ctx context.Context, cfg *operations.DocumentExport) error {
	resp, err := c.Svc.Files.Export(cfg.FileID, exportMIMEType).Context(ctx).Download()
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("Failed to create HTML export of drive file")

		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("Failed to parse exported drive file contents")

		return err
	}

	cfg.HTML = string(body)

	return nil
}
