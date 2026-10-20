package googledrive

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// storeSelectedDocuments saves the Google Docs selected in the Picker as the installation's selected documents user input
func storeSelectedDocuments(ctx context.Context, req types.OperationRequest, resources []types.SelectedResource) error {
	input, err := userInput.Decode(req.Integration)
	if err != nil {
		return ErrUserInputDecode
	}

	input.SelectedDocuments = resources

	return userInput.Save(ctx, req.DB, req.Integration, input)
}
