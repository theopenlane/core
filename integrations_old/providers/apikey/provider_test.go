package apikey

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

func TestBuilder_RejectsAuthTypeMismatch(t *testing.T) {
	s := spec.ProviderSpec{
		Name:     "test_apikey",
		AuthType: types.AuthKindOAuth2,
	}

	builder := Builder(types.ProviderType(s.Name))
	_, err := builder.Build(context.Background(), s)
	require.ErrorIs(t, err, ErrAuthTypeMismatch)
}

func TestBuilder_AcceptsAPIKeyAuthType(t *testing.T) {
	schema, err := jsonx.ToRawMessage(map[string]any{"type": "object"})
	require.NoError(t, err)

	s := spec.ProviderSpec{
		Name:              "test_apikey",
		AuthType:          types.AuthKindAPIKey,
		CredentialsSchema: schema,
	}

	builder := Builder(types.ProviderType(s.Name))
	provider, err := builder.Build(context.Background(), s)
	require.NoError(t, err)
	require.NotNil(t, provider)
}

func TestProvider_BeginAuthNotSupported(t *testing.T) {
	s := spec.ProviderSpec{
		Name:     "test_apikey",
		AuthType: types.AuthKindAPIKey,
	}

	builder := Builder(types.ProviderType(s.Name))
	provider, err := builder.Build(context.Background(), s)
	require.NoError(t, err)

	apiProvider, ok := provider.(*Provider)
	require.True(t, ok)

	_, err = apiProvider.BeginAuth(context.Background(), types.AuthContext{})
	require.ErrorIs(t, err, ErrBeginAuthNotSupported)
}

func TestProvider_Mint_Success(t *testing.T) {
	s := spec.ProviderSpec{
		Name:     "test_apikey",
		AuthType: types.AuthKindAPIKey,
	}

	builder := Builder(types.ProviderType(s.Name))
	provider, err := builder.Build(context.Background(), s)
	require.NoError(t, err)

	apiProvider, ok := provider.(*Provider)
	require.True(t, ok)

	credential, err := apiProvider.Mint(context.Background(), types.CredentialMintRequest{
		Provider: types.ProviderType(s.Name),
		Credential: types.CredentialSet{
			ProviderData: json.RawMessage(`{"apiToken":"secret-token","orgUrl":"https://example.com"}`),
		},
	})
	require.NoError(t, err)

	// Token must be accessible via APITokenFromCredential
	token, err := mustAPITokenFromCredential(credential)
	require.NoError(t, err)
	require.Equal(t, "secret-token", token)
}

func TestProvider_Mint_MissingToken(t *testing.T) {
	s := spec.ProviderSpec{
		Name:     "test_apikey",
		AuthType: types.AuthKindAPIKey,
	}

	builder := Builder(types.ProviderType(s.Name))
	provider, err := builder.Build(context.Background(), s)
	require.NoError(t, err)

	apiProvider, ok := provider.(*Provider)
	require.True(t, ok)

	_, err = apiProvider.Mint(context.Background(), types.CredentialMintRequest{
		Provider: types.ProviderType(s.Name),
		Credential: types.CredentialSet{
			ProviderData: json.RawMessage(`{"orgUrl":"https://example.com"}`),
		},
	})
	require.ErrorIs(t, err, ErrTokenFieldRequired)
}

func TestProvider_Mint_EmptyMetadata(t *testing.T) {
	s := spec.ProviderSpec{
		Name:     "test_apikey",
		AuthType: types.AuthKindAPIKey,
	}

	builder := Builder(types.ProviderType(s.Name))
	provider, err := builder.Build(context.Background(), s)
	require.NoError(t, err)

	apiProvider, ok := provider.(*Provider)
	require.True(t, ok)

	_, err = apiProvider.Mint(context.Background(), types.CredentialMintRequest{
		Provider:   types.ProviderType(s.Name),
		Credential: types.CredentialSet{},
	})
	require.ErrorIs(t, err, ErrProviderMetadataRequired)
}

func TestProvider_Mint_CustomTokenField(t *testing.T) {
	s := spec.ProviderSpec{
		Name:     "test_apikey",
		AuthType: types.AuthKindAPIKey,
	}

	builder := Builder(types.ProviderType(s.Name), WithTokenField("accessKey"))
	provider, err := builder.Build(context.Background(), s)
	require.NoError(t, err)

	apiProvider, ok := provider.(*Provider)
	require.True(t, ok)

	credential, err := apiProvider.Mint(context.Background(), types.CredentialMintRequest{
		Provider: types.ProviderType(s.Name),
		Credential: types.CredentialSet{
			ProviderData: json.RawMessage(`{"accessKey":"custom-token","region":"us-east-1"}`),
		},
	})
	require.NoError(t, err)

	// Token must be stored under "apiToken" key for APITokenFromCredential
	token, err := mustAPITokenFromCredential(credential)
	require.NoError(t, err)
	require.Equal(t, "custom-token", token)
}

func TestProvider_Capabilities(t *testing.T) {
	s := spec.ProviderSpec{
		Name:     "test_apikey",
		AuthType: types.AuthKindAPIKey,
	}

	builder := Builder(types.ProviderType(s.Name))
	provider, err := builder.Build(context.Background(), s)
	require.NoError(t, err)

	caps := provider.Capabilities()
	require.False(t, caps.SupportsRefreshTokens)
}

// mustAPITokenFromCredential is a test helper that extracts the apiToken from ProviderData
func mustAPITokenFromCredential(credential types.CredentialSet) (string, error) {
	var data map[string]json.RawMessage
	if err := json.Unmarshal(credential.ProviderData, &data); err != nil {
		return "", err
	}

	var token string
	if err := json.Unmarshal(data["apiToken"], &token); err != nil {
		return "", err
	}

	return token, nil
}
