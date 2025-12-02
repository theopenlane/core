package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/privacy"
)

func TestIntegrationBuilder(t *testing.T) {
	t.Parallel()

	// setup user context
	orgUser := suite.userBuilder(context.Background(), t)

	// Test that we can create an integration using the builder
	t.Run("Create integration with builder", func(t *testing.T) {
		integration := (&IntegrationBuilder{client: suite.client}).MustNew(orgUser.UserCtx, t)

		assert.Check(t, integration.ID != "")
		assert.Check(t, is.DeepEqual("GitHub Integration Test", integration.Name))
		assert.Check(t, is.DeepEqual("github", integration.Kind))
		assert.Check(t, is.DeepEqual(orgUser.OrganizationID, integration.OwnerID))

		// Clean up
		ctx := privacy.DecisionContext(orgUser.UserCtx, privacy.Allow)
		err := suite.client.db.Integration.DeleteOneID(integration.ID).Exec(ctx)
		assert.NilError(t, err)
	})

	// Test custom integration creation
	t.Run("Create custom integration with builder", func(t *testing.T) {
		integration := (&IntegrationBuilder{
			client:      suite.client,
			Name:        "Slack Integration Test",
			Description: "Custom Slack integration",
			Kind:        "slack",
		}).MustNew(orgUser.UserCtx, t)

		assert.Check(t, integration.ID != "")
		assert.Check(t, is.DeepEqual("Slack Integration Test", integration.Name))
		assert.Check(t, is.DeepEqual("slack", integration.Kind))
		assert.Check(t, is.DeepEqual("Custom Slack integration", integration.Description))

		// Clean up, add the client to the context for fga checks instead of just allowing
		ctx := setContext(orgUser.UserCtx, suite.client.db)
		err := suite.client.db.Integration.DeleteOneID(integration.ID).Exec(ctx)
		assert.NilError(t, err)
	})
}

func TestSecretBuilder(t *testing.T) {
	t.Parallel()

	// setup user context
	orgUser := suite.userBuilder(context.Background(), t)
	ctx := setContext(orgUser.UserCtx, suite.client.db)

	// Create integration first
	integration := (&IntegrationBuilder{client: suite.client}).MustNew(ctx, t)

	t.Run("Create secret with builder", func(t *testing.T) {
		secret := (&SecretBuilder{client: suite.client}).
			WithIntegration(integration.ID).
			WithSecretName("github_access_token").
			WithSecretValue("gho_test_token_123").
			MustNew(ctx, t)

		assert.Check(t, secret.ID != "")
		assert.Check(t, is.DeepEqual("github_access_token", secret.SecretName))
		assert.Check(t, is.DeepEqual("gho_test_token_123", secret.SecretValue))
		assert.Check(t, is.DeepEqual(orgUser.OrganizationID, secret.OwnerID))

		// Verify it's associated with the integration
		integrationSecrets, err := suite.client.db.Integration.QuerySecrets(integration).All(ctx)
		assert.NilError(t, err)
		assert.Check(t, is.Len(integrationSecrets, 1))
		assert.Check(t, is.DeepEqual(secret.ID, integrationSecrets[0].ID))

		// Clean up
		err = suite.client.db.Hush.DeleteOneID(secret.ID).Exec(ctx)
		assert.NilError(t, err)
	})

	// Clean up
	err := suite.client.db.Integration.DeleteOneID(integration.ID).Exec(ctx)
	assert.NilError(t, err)
}

func TestIntegrationWithSecretsRelationship(t *testing.T) {
	t.Parallel()

	// setup user context
	orgUser := suite.userBuilder(context.Background(), t)
	ctx := setContext(orgUser.UserCtx, suite.client.db)

	// Create integration
	integration := (&IntegrationBuilder{client: suite.client}).MustNew(orgUser.UserCtx, t)

	// Create multiple secrets for OAuth tokens
	accessToken := (&SecretBuilder{client: suite.client}).
		WithIntegration(integration.ID).
		WithSecretName("github_access_token").
		WithSecretValue("gho_access_123").
		MustNew(orgUser.UserCtx, t)

	refreshToken := (&SecretBuilder{client: suite.client}).
		WithIntegration(integration.ID).
		WithSecretName("github_refresh_token").
		WithSecretValue("ghr_refresh_456").
		MustNew(orgUser.UserCtx, t)

	expiresAt := (&SecretBuilder{client: suite.client}).
		WithIntegration(integration.ID).
		WithSecretName("github_expires_at").
		WithSecretValue("2024-12-31T23:59:59Z").
		MustNew(orgUser.UserCtx, t)

	t.Run("Integration can query its secrets", func(t *testing.T) {
		secrets, err := suite.client.db.Integration.QuerySecrets(integration).All(ctx)
		assert.NilError(t, err)
		assert.Check(t, is.Len(secrets, 3))

		// Verify secret names
		secretNames := make([]string, len(secrets))
		for i, secret := range secrets {
			secretNames[i] = secret.SecretName
		}

		expectedNames := []string{"github_access_token", "github_refresh_token", "github_expires_at"}
		assert.Check(t, is.DeepEqual(secretNames, expectedNames))
	})

	t.Run("Secrets can query their integration", func(t *testing.T) {
		// Query integration from secret
		integrationFromSecret, err := suite.client.db.Hush.QueryIntegrations(accessToken).Only(ctx)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(integration.ID, integrationFromSecret.ID))
	})

	// Clean up
	err := suite.client.db.Hush.DeleteOneID(accessToken.ID).Exec(ctx)
	assert.NilError(t, err)
	err = suite.client.db.Hush.DeleteOneID(refreshToken.ID).Exec(ctx)
	assert.NilError(t, err)
	err = suite.client.db.Hush.DeleteOneID(expiresAt.ID).Exec(ctx)
	assert.NilError(t, err)
	err = suite.client.db.Integration.DeleteOneID(integration.ID).Exec(ctx)
	assert.NilError(t, err)
}

func TestMutationDeleteIntegration(t *testing.T) {
	// Create integration with secrets
	integration1 := (&IntegrationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	integration2 := (&IntegrationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	integration3 := (&IntegrationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name          string
		integrationID string
		client        *testclient.TestClient
		ctx           context.Context
		errorMsg      string
	}{

		{
			name:          "delete integration, happy path using api token",
			client:        suite.client.apiWithToken,
			ctx:           testUser1.UserCtx,
			integrationID: integration1.ID,
		},
		{
			name:          "delete integration, happy path using personal access token",
			client:        suite.client.apiWithPAT,
			ctx:           testUser1.UserCtx,
			integrationID: integration2.ID,
		},
		{
			name:          "delete integration, no access",
			client:        suite.client.api,
			ctx:           viewOnlyUser.UserCtx,
			integrationID: integration3.ID,
			errorMsg:      notAuthorizedErrorMsg,
		},
		{
			name:          "delete integration, no access another org",
			client:        suite.client.api,
			ctx:           testUser2.UserCtx,
			integrationID: integration3.ID,
			errorMsg:      notFoundErrorMsg,
		},
		{
			name:          "delete integration, happy path",
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			integrationID: integration3.ID,
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

			// make sure the deletedID matches the ID we wanted to delete
			assert.Check(t, is.Equal(tc.integrationID, resp.DeleteIntegration.DeletedID))
		})
	}
}

func TestQueryIntegration(t *testing.T) {
	// create an integration to be queried using testUser1
	integration := (&IntegrationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the Integration
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
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: integration.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: integration.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "integration not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "integration not found, using not authorized user",
			queryID:  integration.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
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

			// add additional assertions for the object
			assert.Check(t, is.Equal(integration.Name, resp.Integration.Name))
			assert.Check(t, is.Equal(integration.Description, *resp.Integration.Description))
			assert.Check(t, is.Equal(integration.Kind, *resp.Integration.Kind))
			assert.Check(t, is.Equal(integration.OwnerID, *resp.Integration.OwnerID))
		})
	}

	(&Cleanup[*generated.IntegrationDeleteOne]{client: suite.client.db.Integration, ID: integration.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryIntegrationWithSecrets(t *testing.T) {
	// create an integration to be queried using testUser1
	integration := (&IntegrationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	// Create multiple secrets for OAuth tokens
	accessToken := (&SecretBuilder{client: suite.client}).
		WithIntegration(integration.ID).
		WithSecretName("github_access_token").
		WithSecretValue("gho_access_123").
		MustNew(testUser1.UserCtx, t)

	refreshToken := (&SecretBuilder{client: suite.client}).
		WithIntegration(integration.ID).
		WithSecretName("github_refresh_token").
		WithSecretValue("ghr_refresh_456").
		MustNew(testUser1.UserCtx, t)

	expiresAt := (&SecretBuilder{client: suite.client}).
		WithIntegration(integration.ID).
		WithSecretName("github_expires_at").
		WithSecretValue("2024-12-31T23:59:59Z").
		MustNew(testUser1.UserCtx, t)

	// add test cases for querying the Integration
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
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:     "read only user, cannot query secrets, only the integration",
			queryID:  integration.ID,
			client:   suite.client.api,
			ctx:      viewOnlyUser.UserCtx,
			errorMsg: notAuthorizedErrorMsg,
		},
		{
			name:    "happy path using personal access token",
			queryID: integration.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "integration not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "integration not found, using not authorized user",
			queryID:  integration.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
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

			// add additional assertions for the object
			assert.Check(t, is.Equal(integration.Name, resp.Integration.Name))
			assert.Check(t, is.Equal(integration.Description, *resp.Integration.Description))
			assert.Check(t, is.Equal(integration.Kind, *resp.Integration.Kind))
			assert.Check(t, is.Equal(integration.OwnerID, *resp.Integration.OwnerID))
			assert.Check(t, is.Len(resp.Integration.Secrets.Edges, 3))
		})
	}

	(&Cleanup[*generated.IntegrationDeleteOne]{client: suite.client.db.Integration, ID: integration.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.HushDeleteOne]{client: suite.client.db.Hush, IDs: []string{accessToken.ID, refreshToken.ID, expiresAt.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestListIntegrations(t *testing.T) {
	// create an integration to be queried using testUser1
	integration1 := (&IntegrationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	integration2 := (&IntegrationBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the Integration
	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no integrations should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
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

	(&Cleanup[*generated.IntegrationDeleteOne]{client: suite.client.db.Integration, IDs: []string{integration1.ID, integration2.ID}}).MustDelete(testUser1.UserCtx, t)
}
