package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/common/integrations/types"
)

func TestBuildOAuthCredentialPayload(t *testing.T) {
	provider := types.ProviderType("oidcgeneric")

	payload, err := BuildOAuthCredentialPayload(
		provider,
		&oauth2.Token{AccessToken: "access-token"},
		&oidc.IDTokenClaims{TokenClaims: oidc.TokenClaims{Subject: "subject-123"}},
	)
	require.NoError(t, err)
	require.Equal(t, types.CredentialKindOAuthToken, payload.Kind)
	require.NotNil(t, payload.Token)
	require.Equal(t, "access-token", payload.Token.AccessToken)
	require.NotNil(t, payload.Claims)
	require.Equal(t, "subject-123", payload.Claims.Subject)
}

func TestBuildAPITokenCredentialPayload(t *testing.T) {
	provider := types.ProviderType("cloudflare")
	meta := map[string]any{
		"apiToken": "token-1",
		"alias":    "prod",
	}

	payload, err := BuildAPITokenCredentialPayload(provider, "token-1", meta)
	require.NoError(t, err)
	require.Equal(t, types.CredentialKindAPIKey, payload.Kind)
	require.Equal(t, "token-1", payload.Data.APIToken)
	require.Equal(t, meta, payload.Data.ProviderData)
}
