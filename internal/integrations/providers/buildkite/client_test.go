package buildkite

import (
	"context"
	"encoding/json"
	"testing"

	buildkitego "github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestBuildBuildkiteClient_MissingToken(t *testing.T) {
	payload := models.CredentialSet{}
	_, err := buildBuildkiteClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, auth.ErrAPITokenMissing)
}

func TestBuildBuildkiteClient_Success(t *testing.T) {
	payload := models.CredentialSet{APIToken: "bkua_test-token"}
	instance, err := buildBuildkiteClient(context.Background(), payload, json.RawMessage(nil))
	require.NoError(t, err)

	client, ok := types.ClientInstanceAs[*buildkitego.Client](instance)
	require.True(t, ok)
	require.NotNil(t, client)
}

func TestBuildBuildkiteClient_OAuthTokenIgnored(t *testing.T) {
	// Buildkite uses API token, not OAuth
	payload := models.CredentialSet{OAuthAccessToken: "oauth-token"}
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
		Credential: models.CredentialSet{APIToken: "bkua_test-token"},
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
