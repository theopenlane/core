package oidclocal

import (
	"context"
	"encoding/json"
	"maps"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// inspectClaims returns the stored OIDC claims for local inspection
func inspectClaims(_ context.Context, _ types.OperationRequest, c Client, _ ClaimsInspect) (json.RawMessage, error) {
	if len(c.Credential.Claims) == 0 {
		return providerkit.EncodeResult(ClaimsInspect{Claims: map[string]any{}}, ErrResultEncode)
	}

	return providerkit.EncodeResult(ClaimsInspect{Claims: maps.Clone(c.Credential.Claims)}, ErrResultEncode)
}
