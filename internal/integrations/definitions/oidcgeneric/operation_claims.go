package oidcgeneric

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// ClaimsInspect returns stored OIDC claims for inspection
type ClaimsInspect struct {
	// Claims is the full set of OIDC claims
	Claims map[string]any `json:"claims"`
}

// Handle adapts claims inspection to the generic operation registration boundary
func (ci ClaimsInspect) Handle() types.OperationHandler {
	return func(_ context.Context, request types.OperationRequest) (json.RawMessage, error) {
		return ci.Run(request.Credential)
	}
}

// Run returns the stored OIDC claims
func (ClaimsInspect) Run(credential types.CredentialSet) (json.RawMessage, error) {
	if credential.Claims == nil {
		return providerkit.EncodeResult(ClaimsInspect{Claims: map[string]any{}}, ErrResultEncode)
	}

	return providerkit.EncodeResult(ClaimsInspect{Claims: credential.Claims}, ErrResultEncode)
}
