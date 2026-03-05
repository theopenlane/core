package github

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/state"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TestGitHubAppCredentialsFromPayload validates credential parsing and normalization
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

func TestResolveMintInputsUsesProviderRuntimeConfigAndProviderState(t *testing.T) {
	providerState := state.IntegrationProviderState{}
	_, err := providerState.MergeProviderData(string(TypeGitHubApp), map[string]any{
		"installationId": "789",
	})
	require.NoError(t, err)

	provider := &appProvider{
		provider:   TypeGitHubApp,
		appID:      "123",
		privateKey: "line1\nline2",
	}

	appID, installationID, privateKey, err := provider.resolveMintInputs(types.CredentialPayload{
		Provider:      TypeGitHubApp,
		ProviderState: &providerState,
		Data:          models.CredentialSet{ProviderData: map[string]any{}},
	})
	require.NoError(t, err)
	require.Equal(t, "123", appID)
	require.Equal(t, "789", installationID)
	require.Equal(t, "line1\nline2", privateKey)
}

// TestAppBuilderClientDescriptors verifies GitHub App providers publish pooled REST and GraphQL clients
func TestAppBuilderClientDescriptors(t *testing.T) {
	spec := config.ProviderSpec{
		Name:     string(TypeGitHubApp),
		AuthType: types.AuthKindGitHubApp,
		GitHubApp: &config.GitHubAppSpec{
			BaseURL: "https://api.github.com",
		},
	}

	provider, err := AppBuilder().Build(context.Background(), spec)
	require.NoError(t, err)
	require.NotNil(t, provider)
	require.True(t, provider.Capabilities().SupportsClientPooling)

	clientProvider, ok := provider.(types.ClientProvider)
	require.True(t, ok)

	descriptors := clientProvider.ClientDescriptors()
	require.Len(t, descriptors, 2)
	require.Equal(t, ClientGitHubAPI, descriptors[0].Name)
	require.Equal(t, ClientGitHubGraphQL, descriptors[1].Name)
}

// TestAppBuilderOperationsIncludeVulnerabilityCollect verifies GitHub App providers expose vulnerability collection
func TestAppBuilderOperationsIncludeVulnerabilityCollect(t *testing.T) {
	spec := config.ProviderSpec{
		Name:     string(TypeGitHubApp),
		AuthType: types.AuthKindGitHubApp,
	}

	provider, err := AppBuilder().Build(context.Background(), spec)
	require.NoError(t, err)
	require.NotNil(t, provider)

	operationProvider, ok := provider.(types.OperationProvider)
	require.True(t, ok)

	operations := operationProvider.Operations()
	require.True(t, lo.ContainsBy(operations, func(op types.OperationDescriptor) bool {
		return op.Name == githubOperationVulnCollect
	}), "expected github app operations to include %q", githubOperationVulnCollect)
}
