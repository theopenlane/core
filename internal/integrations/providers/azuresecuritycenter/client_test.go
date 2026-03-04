package azuresecuritycenter

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
)

func TestBuildAzureSecurityClient_MissingToken(t *testing.T) {
	payload := types.CredentialPayload{}
	_, err := buildAzureSecurityClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, auth.ErrOAuthTokenMissing)
}

func TestBuildAzureSecurityClient_MissingSubscriptionID(t *testing.T) {
	payload := types.CredentialPayload{
		Token: &oauth2.Token{AccessToken: "test-token"},
	}
	_, err := buildAzureSecurityClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, ErrSubscriptionIDMissing)
}

func TestBuildAzureSecurityClient_Success(t *testing.T) {
	payload := types.CredentialPayload{
		Token: &oauth2.Token{AccessToken: "test-token"},
		Data: models.CredentialSet{
			ProviderData: map[string]any{
				"subscriptionId": "sub-123",
			},
		},
	}
	instance, err := buildAzureSecurityClient(context.Background(), payload, json.RawMessage(nil))
	require.NoError(t, err)

	client, ok := types.ClientInstanceAs[*azurePricingsClient](instance)
	require.True(t, ok)
	require.NotNil(t, client)
	assert.Equal(t, "subscriptions/sub-123", client.scope)
}

func TestResolveAzureSecurityClient_PooledClient(t *testing.T) {
	pooled := &azurePricingsClient{scope: "subscriptions/sub-pool"}
	input := types.OperationInput{
		Client: types.NewClientInstance(pooled),
	}
	result, err := resolveAzureSecurityClient(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, pooled == result)
}

func TestResolveAzureSecurityClient_BuildsFromCredential(t *testing.T) {
	input := types.OperationInput{
		Credential: types.CredentialPayload{
			Token: &oauth2.Token{AccessToken: "test-token"},
			Data: models.CredentialSet{
				ProviderData: map[string]any{
					"subscriptionId": "sub-123",
				},
			},
		},
	}
	result, err := resolveAzureSecurityClient(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "subscriptions/sub-123", result.scope)
}

func TestResolveAzureSecurityClient_MissingToken(t *testing.T) {
	input := types.OperationInput{}
	_, err := resolveAzureSecurityClient(context.Background(), input)
	require.ErrorIs(t, err, auth.ErrOAuthTokenMissing)
}

func TestResolveAzureSecurityClient_MissingSubscriptionID(t *testing.T) {
	input := types.OperationInput{
		Credential: types.CredentialPayload{
			Token: &oauth2.Token{AccessToken: "test-token"},
		},
	}
	_, err := resolveAzureSecurityClient(context.Background(), input)
	require.ErrorIs(t, err, ErrSubscriptionIDMissing)
}
