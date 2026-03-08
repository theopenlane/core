package slack

import (
	"context"
	"encoding/json"
	"testing"

	slackgo "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestBuildSlackClient_MissingToken(t *testing.T) {
	payload := models.CredentialSet{}
	_, err := buildSlackClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, auth.ErrOAuthTokenMissing)
}

func TestBuildSlackClient_EmptyAccessToken(t *testing.T) {
	payload := models.CredentialSet{OAuthTokenType: "Bearer"}
	_, err := buildSlackClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, auth.ErrAccessTokenEmpty)
}

func TestBuildSlackClient_Success(t *testing.T) {
	payload := models.CredentialSet{OAuthAccessToken: "xoxb-test-token"}
	instance, err := buildSlackClient(context.Background(), payload, json.RawMessage(nil))
	require.NoError(t, err)

	client, ok := types.ClientInstanceAs[*slackgo.Client](instance)
	require.True(t, ok)
	require.NotNil(t, client)
}

func TestResolveSlackClient_PooledClient(t *testing.T) {
	pooled := slackgo.New("xoxb-pool-token")
	input := types.OperationInput{
		Client: types.NewClientInstance(pooled),
	}
	result, err := resolveSlackClient(input)
	require.NoError(t, err)
	assert.True(t, pooled == result)
}

func TestResolveSlackClient_BuildsFromCredential(t *testing.T) {
	input := types.OperationInput{
		Credential: models.CredentialSet{OAuthAccessToken: "xoxb-test-token"},
	}
	result, err := resolveSlackClient(input)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestResolveSlackClient_MissingCredential(t *testing.T) {
	input := types.OperationInput{}
	_, err := resolveSlackClient(input)
	require.ErrorIs(t, err, auth.ErrOAuthTokenMissing)
}
