package eventstest_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/iam/auth"
)

func TestIntegrationBuilder(t *testing.T) {
	orgUser := suite.UserBuilder(context.Background(), t)

	t.Run("Create integration with builder", func(t *testing.T) {
		integration := (&th.IntegrationBuilder{Client: suite.Client}).MustNew(orgUser.UserCtx, t)

		assert.Check(t, integration.ID != "")
		assert.Check(t, is.DeepEqual("GitHub Integration Test", integration.Name))
		assert.Check(t, is.DeepEqual("github", integration.Kind))
		assert.Check(t, is.DeepEqual(orgUser.OrganizationID, integration.OwnerID))

		// Clean up
		ctx := auth.WithInternalOperationContext(orgUser.UserCtx)
		err := suite.Client.DB.Integration.DeleteOneID(integration.ID).Exec(ctx)
		assert.NilError(t, err)
	})

	t.Run("Create custom integration with builder", func(t *testing.T) {
		integration := (&th.IntegrationBuilder{
			Client:      suite.Client,
			Name:        "Slack Integration Test",
			Description: "Custom Slack integration",
			Kind:        "slack",
		}).MustNew(orgUser.UserCtx, t)

		assert.Check(t, integration.ID != "")
		assert.Check(t, is.DeepEqual("Slack Integration Test", integration.Name))
		assert.Check(t, is.DeepEqual("slack", integration.Kind))
		assert.Check(t, is.DeepEqual("Custom Slack integration", integration.Description))

		// Clean up, add the client to the context for fga checks instead of just allowing
		ctx := th.SetInternalContext(orgUser.UserCtx, suite.Client.DB)
		err := suite.Client.DB.Integration.DeleteOneID(integration.ID).Exec(ctx)
		assert.NilError(t, err)
	})
}

func TestSecretBuilder(t *testing.T) {
	orgUser := suite.UserBuilder(context.Background(), t)
	ctx := th.SetInternalContext(orgUser.UserCtx, suite.Client.DB)

	integration := (&th.IntegrationBuilder{Client: suite.Client}).MustNew(ctx, t)

	t.Run("Create secret with builder", func(t *testing.T) {
		secret := (&th.SecretBuilder{Client: suite.Client}).
			WithIntegration(integration.ID).
			WithSecretName("github_access_token").
			WithSecretValue("gho_test_token_123").
			MustNew(ctx, t)

		assert.Check(t, secret.ID != "")
		assert.Check(t, is.DeepEqual("github_access_token", secret.SecretName))
		assert.Check(t, is.DeepEqual("gho_test_token_123", secret.SecretValue))
		assert.Check(t, is.DeepEqual(orgUser.OrganizationID, secret.OwnerID))

		integrationSecrets, err := suite.Client.DB.Integration.QuerySecrets(integration).All(ctx)
		assert.NilError(t, err)
		assert.Check(t, is.Len(integrationSecrets, 1))
		assert.Check(t, is.DeepEqual(secret.ID, integrationSecrets[0].ID))

		err = suite.Client.DB.Hush.DeleteOneID(secret.ID).Exec(ctx)
		assert.NilError(t, err)
	})

	err := suite.Client.DB.Integration.DeleteOneID(integration.ID).Exec(ctx)
	assert.NilError(t, err)
}

func TestIntegrationWithSecretsRelationship(t *testing.T) {
	orgUser := suite.UserBuilder(context.Background(), t)
	ctx := th.SetInternalContext(orgUser.UserCtx, suite.Client.DB)

	integration := (&th.IntegrationBuilder{Client: suite.Client}).MustNew(orgUser.UserCtx, t)

	accessToken := (&th.SecretBuilder{Client: suite.Client}).
		WithIntegration(integration.ID).
		WithSecretName("github_access_token").
		WithSecretValue("gho_access_123").
		MustNew(orgUser.UserCtx, t)

	(&th.SecretBuilder{Client: suite.Client}).
		WithIntegration(integration.ID).
		WithSecretName("github_refresh_token").
		WithSecretValue("ghr_refresh_456").
		MustNew(orgUser.UserCtx, t)

	(&th.SecretBuilder{Client: suite.Client}).
		WithIntegration(integration.ID).
		WithSecretName("github_expires_at").
		WithSecretValue("2024-12-31T23:59:59Z").
		MustNew(orgUser.UserCtx, t)

	t.Run("Integration can query its secrets", func(t *testing.T) {
		secrets, err := suite.Client.DB.Integration.QuerySecrets(integration).All(ctx)
		assert.NilError(t, err)
		assert.Check(t, is.Len(secrets, 3))

		secretNames := make([]string, len(secrets))
		for i, secret := range secrets {
			secretNames[i] = secret.SecretName
		}

		expectedNames := []string{"github_access_token", "github_refresh_token", "github_expires_at"}
		assert.Check(t, is.DeepEqual(secretNames, expectedNames))
	})

	t.Run("Secrets can query their integration", func(t *testing.T) {
		integrationFromSecret, err := suite.Client.DB.Hush.QueryIntegrations(accessToken).Only(ctx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(integration.ID, integrationFromSecret.ID))
	})

	th.CleanupOrganizationDataWithContext(ctx, t)
}

func TestMutationDeleteIntegration(t *testing.T) {
	integration1 := (&th.IntegrationBuilder{Client: suite.Client, Kind: "github"}).MustNew(th.SharedTestUser1.UserCtx, t)
	integration2 := (&th.IntegrationBuilder{Client: suite.Client, Kind: "jira"}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name          string
		integrationID string
		client        *testclient.TestClient
		ctx           context.Context
		errorMsg      string
	}{

		{
			name:          "delete integration, happy path using api token",
			client:        suite.Client.APIWithToken,
			ctx:           context.Background(),
			integrationID: integration1.ID,
			errorMsg:      th.NotAuthorizedErrorMsg,
		},
		{
			name:          "delete integration, happy path using personal access token",
			client:        suite.Client.APIWithPAT,
			ctx:           context.Background(),
			integrationID: integration1.ID,
		},
		{
			name:          "delete integration, no access",
			client:        suite.Client.API,
			ctx:           th.SharedViewOnlyUser.UserCtx,
			integrationID: integration2.ID,
			errorMsg:      th.NotAuthorizedErrorMsg,
		},
		{
			name:          "delete integration, no access another org",
			client:        suite.Client.API,
			ctx:           th.SharedTestUser2.UserCtx,
			integrationID: integration2.ID,
			errorMsg:      th.NotFoundErrorMsg,
		},
		{
			name:          "delete integration, happy path",
			client:        suite.Client.API,
			ctx:           th.SharedTestUser1.UserCtx,
			integrationID: integration2.ID,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteIntegration(tc.ctx, tc.integrationID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.DeleteIntegration.DeletedID != "")

			assert.Check(t, is.Equal(tc.integrationID, resp.DeleteIntegration.DeletedID))
		})
	}
}

func TestQueryIntegration(t *testing.T) {
	integration := (&th.IntegrationBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: integration.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: integration.ID,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: integration.ID,
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "integration not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "integration not found, using not authorized user",
			queryID:  integration.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetIntegrationByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Integration.ID))

			assert.Check(t, is.Equal(integration.Name, resp.Integration.Name))
			assert.Check(t, is.Equal(integration.Description, *resp.Integration.Description))
			assert.Check(t, is.Equal(integration.Kind, *resp.Integration.Kind))
			assert.Check(t, is.Equal(integration.OwnerID, *resp.Integration.OwnerID))
		})
	}

	(&th.Cleanup[*generated.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: integration.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryIntegrationWithSecrets(t *testing.T) {
	integration := (&th.IntegrationBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	accessToken := (&th.SecretBuilder{Client: suite.Client}).
		WithIntegration(integration.ID).
		WithSecretName("github_access_token").
		WithSecretValue("gho_access_123").
		MustNew(th.SharedTestUser1.UserCtx, t)

	refreshToken := (&th.SecretBuilder{Client: suite.Client}).
		WithIntegration(integration.ID).
		WithSecretName("github_refresh_token").
		WithSecretValue("ghr_refresh_456").
		MustNew(th.SharedTestUser1.UserCtx, t)

	expiresAt := (&th.SecretBuilder{Client: suite.Client}).
		WithIntegration(integration.ID).
		WithSecretName("github_expires_at").
		WithSecretValue("2024-12-31T23:59:59Z").
		MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: integration.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:     "read only user, cannot query secrets, only the integration",
			queryID:  integration.ID,
			client:   suite.Client.API,
			ctx:      th.SharedViewOnlyUser.UserCtx,
			errorMsg: th.NotAuthorizedErrorMsg,
		},
		{
			name:    "happy path using personal access token",
			queryID: integration.ID,
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "integration not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "integration not found, using not authorized user",
			queryID:  integration.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetIntegrationByIDWithSecrets(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Integration.ID))

			assert.Check(t, is.Equal(integration.Name, resp.Integration.Name))
			assert.Check(t, is.Equal(integration.Description, *resp.Integration.Description))
			assert.Check(t, is.Equal(integration.Kind, *resp.Integration.Kind))
			assert.Check(t, is.Equal(integration.OwnerID, *resp.Integration.OwnerID))
			assert.Check(t, is.Len(resp.Integration.Secrets.Edges, 3))
		})
	}

	(&th.Cleanup[*generated.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: integration.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.HushDeleteOne]{Client: suite.Client.DB.Hush, IDs: []string{accessToken.ID, refreshToken.ID, expiresAt.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestListIntegrations(t *testing.T) {
	integration1 := (&th.IntegrationBuilder{Client: suite.Client, Kind: "github"}).MustNew(th.SharedTestUser1.UserCtx, t)
	integration2 := (&th.IntegrationBuilder{Client: suite.Client, Kind: "slack"}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using api token",
			client:          suite.Client.APIWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.Client.APIWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no integrations should be returned",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllIntegrations(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Integrations.Edges, tc.expectedResults))
		})
	}

	(&th.Cleanup[*generated.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, IDs: []string{integration1.ID, integration2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}
