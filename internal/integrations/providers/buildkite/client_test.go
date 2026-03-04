package buildkite

import (
	"context"
	"encoding/json"
	"testing"

	buildkitego "github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
)

func TestBuildBuildkiteClient_MissingToken(t *testing.T) {
	payload := types.CredentialPayload{}
	_, err := buildBuildkiteClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, auth.ErrAPITokenMissing)
}

func TestBuildBuildkiteClient_Success(t *testing.T) {
	payload := types.CredentialPayload{
		Data: models.CredentialSet{APIToken: "bkua_test-token"},
	}
	instance, err := buildBuildkiteClient(context.Background(), payload, json.RawMessage(nil))
	require.NoError(t, err)

	client, ok := types.ClientInstanceAs[*buildkitego.Client](instance)
	require.True(t, ok)
	require.NotNil(t, client)
}

func TestBuildBuildkiteClient_OAuthTokenIgnored(t *testing.T) {
	// Buildkite uses API token, not OAuth
	payload := types.CredentialPayload{
		Token: &oauth2.Token{AccessToken: "oauth-token"},
	}
	_, err := buildBuildkiteClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, auth.ErrAPITokenMissing)
}

func TestResolveBuildkiteClient_PooledClient(t *testing.T) {
	pooled, err := buildkitego.NewOpts(buildkitego.WithTokenAuth("pool-token"))
	require.NoError(t, err)

	input := types.OperationInput{
		Client: types.NewClientInstance(pooled),
	}
	result, err := resolveBuildkiteClient(input)
	require.NoError(t, err)
	assert.True(t, pooled == result)
}

func TestResolveBuildkiteClient_BuildsFromCredential(t *testing.T) {
	input := types.OperationInput{
		Credential: types.CredentialPayload{
			Data: models.CredentialSet{APIToken: "bkua_test-token"},
		},
	}
	result, err := resolveBuildkiteClient(input)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestResolveBuildkiteClient_MissingCredential(t *testing.T) {
	input := types.OperationInput{}
	_, err := resolveBuildkiteClient(input)
	require.ErrorIs(t, err, auth.ErrAPITokenMissing)
}
