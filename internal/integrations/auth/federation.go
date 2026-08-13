package auth

import (
	"context"

	"golang.org/x/oauth2"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/oidc"
)

// FederationSpec describes one identity federation exchange for an installation-scoped client
type FederationSpec struct {
	// Audience is the relying party identity the assertion is exchanged against
	Audience string
	// Scopes are the OAuth scopes requested during the exchange
	Scopes []string
	// Endpoint is the RFC 8693 token exchange endpoint
	Endpoint string
	// AssertionOptions customize the minted assertions
	AssertionOptions []tokens.ConfigOpt
}

// FederatedTokenSource builds a caching token source authenticating the installation's organization via identity federation
func FederatedTokenSource(ctx context.Context, req types.ClientBuildRequest, spec FederationSpec) (oauth2.TokenSource, error) {
	return oidc.NewTokenSource(ctx, oidc.FederationSource{
		Manager:          req.TokenManager,
		OrganizationID:   req.Integration.OwnerID,
		Audience:         spec.Audience,
		Scopes:           spec.Scopes,
		Endpoint:         spec.Endpoint,
		AssertionOptions: spec.AssertionOptions,
	})
}
