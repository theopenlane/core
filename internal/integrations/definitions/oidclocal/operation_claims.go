package oidclocal

import (
	"context"
	"encoding/json"
	"maps"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// inspectClaims returns the stored OIDC claims for local inspection
func inspectClaims(_ context.Context, req types.OperationRequest, _ ClaimsInspect) (json.RawMessage, error) {
	cred, ok, err := oidcCredential.Resolve(req.Credentials)
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
