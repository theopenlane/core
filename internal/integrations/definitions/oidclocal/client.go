package oidclocal

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// buildClient wraps the stored credential as the client the local OIDC operations read
func buildClient(_ context.Context, req types.ConnectionRequest[oidcLocalCred]) (Client, error) {
	return Client{Credential: req.Credential}, nil
}
