package apikey

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
)

func TestProviderMint(t *testing.T) {
	spec := config.ProviderSpec{
		Name:              "test_apikey",
		AuthType:          types.AuthKindAPIKey,
		CredentialsSchema: map[string]any{"type": "object"},
	}

	builder := Builder(types.ProviderType(spec.Name))
	provider, err := builder.Build(context.Background(), spec)
	require.NoError(t, err)

	apiProvider, ok := provider.(*Provider)
	require.True(t, ok)

	_, err = apiProvider.BeginAuth(context.Background(), types.AuthContext{})
	require.Error(t, err)

	payload, err := apiProvider.Mint(context.Background(), types.CredentialSubject{
		Provider: types.ProviderType(spec.Name),
		Credential: mustBuildPayload(t, spec.Name, map[string]any{
			"apiToken": "secret-token",
			"alias":    "test",
		}),
	})
	require.NoError(t, err)

	require.Equal(t, types.CredentialKindAPIKey, payload.Kind)
	require.Equal(t, "secret-token", payload.Data.APIToken)
	require.Equal(t, map[string]any{"apiToken": "secret-token", "alias": "test"}, payload.Data.ProviderData)
}

func mustBuildPayload(t *testing.T, provider string, providerData map[string]any) types.CredentialPayload {
	t.Helper()

	payload, err := types.NewCredentialBuilder(types.ProviderType(provider)).
		With(
			types.WithCredentialKind(types.CredentialKindMetadata),
			types.WithCredentialSet(models.CredentialSet{
				ProviderData: providerData,
			}),
		).Build()
	require.NoError(t, err)
	return payload
}
