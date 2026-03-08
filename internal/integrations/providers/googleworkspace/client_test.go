package googleworkspace

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestBuildGoogleWorkspaceClient_MissingToken(t *testing.T) {
	payload := models.CredentialSet{}
	_, err := buildGoogleWorkspaceClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, auth.ErrOAuthTokenMissing)
}

func TestBuildGoogleWorkspaceClient_EmptyAccessToken(t *testing.T) {
	payload := models.CredentialSet{OAuthTokenType: "Bearer"}
	_, err := buildGoogleWorkspaceClient(context.Background(), payload, json.RawMessage(nil))
	require.ErrorIs(t, err, auth.ErrAccessTokenEmpty)
}

func TestBuildGoogleWorkspaceClient_Success(t *testing.T) {
	payload := models.CredentialSet{OAuthAccessToken: "ya29.test-token"}
	instance, err := buildGoogleWorkspaceClient(context.Background(), payload, json.RawMessage(nil))
	require.NoError(t, err)

	svc, ok := types.ClientInstanceAs[*admin.Service](instance)
	require.True(t, ok)
	require.NotNil(t, svc)
}

func TestResolveGoogleWorkspaceClient_PooledClient(t *testing.T) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "ya29.pool-token"})
	pooled, err := admin.NewService(context.Background(), option.WithTokenSource(ts))
	require.NoError(t, err)

	input := types.OperationInput{
		Client: types.NewClientInstance(pooled),
	}
	result, err := resolveGoogleWorkspaceClient(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, pooled == result)
}

func TestResolveGoogleWorkspaceClient_BuildsFromCredential(t *testing.T) {
	input := types.OperationInput{
		Credential: models.CredentialSet{OAuthAccessToken: "ya29.test-token"},
	}
	result, err := resolveGoogleWorkspaceClient(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestResolveGoogleWorkspaceClient_MissingCredential(t *testing.T) {
	input := types.OperationInput{}
	_, err := resolveGoogleWorkspaceClient(context.Background(), input)
	require.ErrorIs(t, err, auth.ErrOAuthTokenMissing)
}
