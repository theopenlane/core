package auth

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func TestBuildOAuthCredentialSet(t *testing.T) {
	credential, err := BuildOAuthCredentialSet(
		&oauth2.Token{AccessToken: "access-token"},
		&oidc.IDTokenClaims{TokenClaims: oidc.TokenClaims{Subject: "subject-123"}},
	)
	require.NoError(t, err)
	require.Equal(t, "access-token", credential.OAuthAccessToken)
	require.Equal(t, "subject-123", credential.Claims["sub"])
}

func TestBuildAPITokenCredentialSet(t *testing.T) {
	credential := BuildAPITokenCredentialSet("token-1", json.RawMessage(`{"apiToken":"token-1","alias":"prod"}`))
	require.Equal(t, "token-1", credential.APIToken)
	require.JSONEq(t, `{"apiToken":"token-1","alias":"prod"}`, string(credential.ProviderData))
}
