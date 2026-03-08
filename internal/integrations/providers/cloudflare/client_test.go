package cloudflare

import (
	"context"
	"encoding/json"
	"testing"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestBuildCloudflareClient_MissingToken(t *testing.T) {
	payload := models.CredentialSet{}
	_, err := buildCloudflareClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, auth.ErrAPITokenMissing)
}

func TestBuildCloudflareClient_Success(t *testing.T) {
	payload := models.CredentialSet{APIToken: "test-token"}
	instance, err := buildCloudflareClient(context.Background(), payload, json.RawMessage(nil))
	require.NoError(t, err)

	client, ok := types.ClientInstanceAs[*cf.Client](instance)
	require.True(t, ok)
	require.NotNil(t, client)
}

func TestResolveCloudflareClient_PooledClient(t *testing.T) {
	pooled := cf.NewClient(option.WithAPIToken("pool-token"))
	input := types.OperationInput{
		Client: types.NewClientInstance(pooled),
	}
	result, err := resolveCloudflareClient(input)
	require.NoError(t, err)
	assert.True(t, pooled == result)
}

func TestResolveCloudflareClient_BuildsFromCredential(t *testing.T) {
	input := types.OperationInput{
		Credential: models.CredentialSet{APIToken: "test-token"},
	}
	result, err := resolveCloudflareClient(input)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestResolveCloudflareClient_MissingCredential(t *testing.T) {
	input := types.OperationInput{}
	_, err := resolveCloudflareClient(input)
	require.ErrorIs(t, err, auth.ErrAPITokenMissing)
}

func TestBuildCloudflareClient_OAuthTokenIgnored(t *testing.T) {
	// Cloudflare uses API token, not OAuth — OAuth token alone should fail
	payload := models.CredentialSet{OAuthAccessToken: "oauth-token"}
	_, err := buildCloudflareClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, auth.ErrAPITokenMissing)
}
