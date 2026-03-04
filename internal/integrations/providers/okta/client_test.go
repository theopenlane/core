package okta

import (
	"context"
	"encoding/json"
	"testing"

	okta "github.com/okta/okta-sdk-golang/v5/okta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
)

func TestBuildOktaClient_MissingToken(t *testing.T) {
	payload := types.CredentialPayload{}
	_, err := buildOktaClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, auth.ErrAPITokenMissing)
}

func TestBuildOktaClient_MissingOrgURL(t *testing.T) {
	payload := types.CredentialPayload{
		Data: models.CredentialSet{APIToken: "test-token"},
	}
	_, err := buildOktaClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, ErrCredentialsMissing)
}

func TestBuildOktaClient_Success(t *testing.T) {
	payload := types.CredentialPayload{
		Data: models.CredentialSet{
			APIToken: "test-token",
			ProviderData: map[string]any{
				"orgUrl": "https://example.okta.com",
			},
		},
	}
	instance, err := buildOktaClient(context.Background(), payload, json.RawMessage(nil))
	require.NoError(t, err)

	client, ok := types.ClientInstanceAs[*okta.APIClient](instance)
	require.True(t, ok)
	require.NotNil(t, client)
}

func TestResolveOktaClient_PooledClient(t *testing.T) {
	cfg, err := okta.NewConfiguration(
		okta.WithOrgUrl("https://example.okta.com"),
		okta.WithToken("pool-token"),
	)
	require.NoError(t, err)

	pooled := okta.NewAPIClient(cfg)
	input := types.OperationInput{
		Client: types.NewClientInstance(pooled),
	}
	result, err := resolveOktaClient(input)
	require.NoError(t, err)
	assert.True(t, pooled == result)
}

func TestResolveOktaClient_BuildsFromCredential(t *testing.T) {
	input := types.OperationInput{
		Credential: types.CredentialPayload{
			Data: models.CredentialSet{
				APIToken: "test-token",
				ProviderData: map[string]any{
					"orgUrl": "https://example.okta.com",
				},
			},
		},
	}
	result, err := resolveOktaClient(input)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestResolveOktaClient_MissingToken(t *testing.T) {
	input := types.OperationInput{}
	_, err := resolveOktaClient(input)
	require.ErrorIs(t, err, auth.ErrAPITokenMissing)
}

func TestResolveOktaClient_MissingOrgURL(t *testing.T) {
	input := types.OperationInput{
		Credential: types.CredentialPayload{
			Data: models.CredentialSet{APIToken: "test-token"},
		},
	}
	_, err := resolveOktaClient(input)
	require.ErrorIs(t, err, ErrCredentialsMissing)
}
