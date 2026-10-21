package googledrive

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// storeSelectedDocuments saves the Google Docs selected in the Picker as the installation's selected documents metadata
func storeSelectedDocuments(ctx context.Context, req types.OperationRequest, resources []types.SelectedResource) error {
	metadata, err := installation.Decode(req.Integration)
	if err != nil {
		return ErrInstallationMetadataDecode
	}

	metadata.SelectedDocuments = resources

	return installation.Save(ctx, req.DB, req.Integration, metadata)
}
