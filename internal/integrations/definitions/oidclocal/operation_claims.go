package oidclocal

import (
	"context"
	"encoding/json"
	"maps"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Handle adapts claim inspection to the generic operation registration boundary
func (ci ClaimsInspect) Handle() types.OperationHandler {
	return func(_ context.Context, request types.OperationRequest) (json.RawMessage, error) {
		return ci.Run(request.Credentials)
	}
}

// Run returns the stored OIDC claims for local inspection
func (ClaimsInspect) Run(bindings types.CredentialBindings) (json.RawMessage, error) {
	cred, ok, err := oidcCredential.Resolve(bindings)
	if err != nil {
		return nil, ErrCredentialDecode
	}

	if !ok {
		return providerkit.EncodeResult(ClaimsInspect{Claims: map[string]any{}}, ErrResultEncode)
	}

	if len(cred.Claims) == 0 {
		return providerkit.EncodeResult(ClaimsInspect{Claims: map[string]any{}}, ErrResultEncode)
	}

	return providerkit.EncodeResult(ClaimsInspect{Claims: maps.Clone(cred.Claims)}, ErrResultEncode)
}
