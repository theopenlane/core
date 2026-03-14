package scim

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

type scimDirectorySyncDetails struct {
	Message string `json:"message"`
}

// buildSCIMClient returns nil since SCIM is push-based (no outbound client)
func buildSCIMClient(_ context.Context, _ types.ClientBuildRequest) (any, error) {
	return nil, nil
}

// runDirectorySyncOperation is a no-op placeholder for push-based SCIM sync
func runDirectorySyncOperation(_ context.Context, _ *generated.Integration, _ types.CredentialSet, _ any, _ json.RawMessage) (json.RawMessage, error) {
	return jsonx.ToRawMessage(scimDirectorySyncDetails{
		Message: "scim is push-based; sync is triggered by the external identity provider",
	})
}

// verifyWebhook verifies an inbound SCIM webhook request
func verifyWebhook(_ context.Context, _ *http.Request) error {
	return ErrWebhookNotImplemented
}

// resolveWebhook resolves an inbound SCIM webhook to an integration
func resolveWebhook(_ context.Context, _ *http.Request) (*generated.Integration, error) {
	return nil, ErrWebhookNotImplemented
}

// handleWebhook handles an inbound SCIM webhook request
func handleWebhook(_ context.Context, _ *http.Request, _ *generated.Integration) error {
	return ErrWebhookNotImplemented
}
