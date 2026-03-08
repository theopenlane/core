package apikey

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TestProviderMint verifies API key credentials are normalized and persisted
func TestProviderMint(t *testing.T) {
	schema, err := jsonx.ToRawMessage(map[string]any{"type": "object"})
	require.NoError(t, err)

	spec := config.ProviderSpec{
		Name:              "test_apikey",
		AuthType:          types.AuthKindAPIKey,
		CredentialsSchema: schema,
	}

	builder := Builder(types.ProviderType(spec.Name))
	provider, err := builder.Build(context.Background(), spec)
	require.NoError(t, err)

	apiProvider, ok := provider.(*Provider)
	require.True(t, ok)

	_, err = apiProvider.BeginAuth(context.Background(), types.AuthContext{})
	require.Error(t, err)

	payload, err := apiProvider.Mint(context.Background(), types.CredentialMintRequest{
		Provider:   types.ProviderType(spec.Name),
		Credential: mustBuildCredential(json.RawMessage(`{"apiToken":"secret-token","alias":"test"}`)),
	})
	require.NoError(t, err)

	require.Equal(t, "secret-token", payload.APIToken)
	require.JSONEq(t, `{"alias":"test"}`, string(payload.ProviderData))
}

func TestProviderMint_NormalizesExistingTokenWithoutMetadata(t *testing.T) {
	schema, err := jsonx.ToRawMessage(map[string]any{"type": "object"})
	require.NoError(t, err)

	spec := config.ProviderSpec{
		Name:              "test_apikey",
		AuthType:          types.AuthKindAPIKey,
		CredentialsSchema: schema,
	}

	builder := Builder(types.ProviderType(spec.Name))
	provider, err := builder.Build(context.Background(), spec)
	require.NoError(t, err)

	apiProvider, ok := provider.(*Provider)
	require.True(t, ok)

	payload, err := apiProvider.Mint(context.Background(), types.CredentialMintRequest{
		Provider: types.ProviderType(spec.Name),
		Credential: models.CredentialSet{
			APIToken: "secret-token",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "secret-token", payload.APIToken)
	require.Empty(t, payload.ProviderData)
}

// mustBuildCredential builds a credential set for provider metadata tests.
func mustBuildCredential(providerData json.RawMessage) models.CredentialSet {
	return models.CredentialSet{ProviderData: providerData}
}
