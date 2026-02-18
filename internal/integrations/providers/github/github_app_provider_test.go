package github

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
)

// TestGitHubAppCredentialsFromPayload validates credential parsing and normalization.
func TestGitHubAppCredentialsFromPayload(t *testing.T) {
	payload := types.CredentialPayload{Provider: types.ProviderUnknown}
	_, _, _, err := githubAppCredentialsFromPayload(payload)
	require.ErrorIs(t, err, ErrProviderNotInitialized)

	payload = types.CredentialPayload{
		Provider: TypeGitHubApp,
		Data:     models.CredentialSet{ProviderData: map[string]any{}},
	}
	_, _, _, err = githubAppCredentialsFromPayload(payload)
	require.ErrorIs(t, err, ErrAppIDMissing)

	payload.Data.ProviderData["appId"] = "123"
	_, _, _, err = githubAppCredentialsFromPayload(payload)
	require.ErrorIs(t, err, ErrInstallationIDMissing)

	payload.Data.ProviderData["installationId"] = "456"
	_, _, _, err = githubAppCredentialsFromPayload(payload)
	require.ErrorIs(t, err, ErrPrivateKeyMissing)

	payload.Data.ProviderData["privateKey"] = "line1\\nline2"
	appID, installationID, privateKey, err := githubAppCredentialsFromPayload(payload)
	require.NoError(t, err)
	require.Equal(t, "123", appID)
	require.Equal(t, "456", installationID)
	require.Equal(t, "line1\nline2", privateKey)
}
