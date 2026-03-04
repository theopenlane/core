package slack

import (
	"context"
	"encoding/json"
	"testing"

	slackgo "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

func TestBuildSlackClient_MissingToken(t *testing.T) {
	payload := types.CredentialPayload{}
	_, err := buildSlackClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, auth.ErrOAuthTokenMissing)
}

func TestBuildSlackClient_EmptyAccessToken(t *testing.T) {
	payload := types.CredentialPayload{
		Token: &oauth2.Token{AccessToken: ""},
	}
	_, err := buildSlackClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, auth.ErrAccessTokenEmpty)
}

func TestBuildSlackClient_Success(t *testing.T) {
	payload := types.CredentialPayload{
		Token: &oauth2.Token{AccessToken: "xoxb-test-token"},
	}
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
		Credential: types.CredentialPayload{
			Token: &oauth2.Token{AccessToken: "xoxb-test-token"},
		},
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
