package googledrive

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/logx"
)

// storeSelectedDocuments saves the Google Docs selected in the Picker as the installation's selected documents metadata and moves the queued folder sync up to run now
func storeSelectedDocuments(ctx context.Context, req types.OperationRequest, resources []types.SelectedResource) error {
	metadata, err := installation.Decode(req.Integration)
	if err != nil {
		return ErrInstallationMetadataDecode
	}

	metadata.SelectedDocuments = resources

	if err := installation.Save(ctx, req.DB, req.Integration, metadata); err != nil {
		return err
	}

	if _, err := req.Services.TriggerReconcile(ctx, req.Integration.ID, folderSyncOperation.Name()); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("googledrive: failed to trigger folder sync after document selection")
	}

	return nil
}
