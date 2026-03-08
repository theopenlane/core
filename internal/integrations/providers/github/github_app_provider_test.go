package github

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/state"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TestGitHubAppInstallationIDFromCredential validates installation ID resolution from payload and provider state.
func TestGitHubAppInstallationIDFromCredential(t *testing.T) {
	payload := models.CredentialSet{ProviderData: json.RawMessage(`{}`)}

	_, err := githubAppInstallationIDFromCredential(payload, nil)
	require.ErrorIs(t, err, ErrInstallationIDMissing)

	providerState := state.IntegrationProviderState{}
	_, err = providerState.MergeProviderData(string(TypeGitHubApp), json.RawMessage(`{"installationId":"789"}`))
	require.NoError(t, err)

	installationID, err := githubAppInstallationIDFromCredential(payload, &providerState)
	require.NoError(t, err)
	require.Equal(t, "789", installationID)

	payload.ProviderData = json.RawMessage(`{"installationId":"456"}`)
	installationID, err = githubAppInstallationIDFromCredential(payload, &providerState)
	require.NoError(t, err)
	require.Equal(t, "456", installationID)
}

func TestNormalizePrivateKeyEscapedNewlines(t *testing.T) {
	require.Equal(t, "line1\nline2", normalizePrivateKey("line1\\nline2"))
	require.Equal(t, "line1\nline2", normalizePrivateKey("line1\nline2"))
}

func TestResolveMintInputsUsesProviderRuntimeConfigAndProviderState(t *testing.T) {
	providerState := state.IntegrationProviderState{}
	_, err := providerState.MergeProviderData(string(TypeGitHubApp), json.RawMessage(`{"installationId":"789"}`))
	require.NoError(t, err)

	provider := &appProvider{
		provider:   TypeGitHubApp,
		appID:      "123",
		privateKey: "line1\nline2",
	}

	appID, installationID, privateKey, err := provider.resolveMintInputs(
		models.CredentialSet{ProviderData: json.RawMessage(`{}`)},
		&providerState,
	)
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
