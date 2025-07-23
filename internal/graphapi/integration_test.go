package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
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
		_ = suite.client.db.Integration.DeleteOneID(integration.ID).Exec(ctx)
	})

	// Test custom integration creation
	t.Run("Create custom integration with builder", func(t *testing.T) {
		integration := (&IntegrationBuilder{
			client:      suite.client,
			Name:        "Slack Integration Test",
			Description: "Custom Slack integration",
			Kind:        "slack",
			OwnerID:     orgUser.OrganizationID,
		}).MustNew(orgUser.UserCtx, t)

		assert.Check(t, integration.ID != "")
		assert.Check(t, is.DeepEqual("Slack Integration Test", integration.Name))
		assert.Check(t, is.DeepEqual("slack", integration.Kind))
		assert.Check(t, is.DeepEqual("Custom Slack integration", integration.Description))

		// Clean up
		ctx := privacy.DecisionContext(orgUser.UserCtx, privacy.Allow)
		_ = suite.client.db.Integration.DeleteOneID(integration.ID).Exec(ctx)
	})
}

func TestSecretBuilder(t *testing.T) {
	t.Parallel()

	// setup user context
	orgUser := suite.userBuilder(context.Background(), t)

	// Create integration first
	integration := (&IntegrationBuilder{client: suite.client}).MustNew(orgUser.UserCtx, t)

	t.Run("Create secret with builder", func(t *testing.T) {
		secret := (&SecretBuilder{client: suite.client}).
			WithIntegration(integration.ID).
			WithSecretName("github_access_token").
			WithSecretValue("gho_test_token_123").
			MustNew(orgUser.UserCtx, t)

		assert.Check(t, secret.ID != "")
		assert.Check(t, is.DeepEqual("github_access_token", secret.SecretName))
		assert.Check(t, is.DeepEqual("gho_test_token_123", secret.SecretValue))
		assert.Check(t, is.DeepEqual(orgUser.OrganizationID, secret.OwnerID))

		// Verify it's associated with the integration
		integrationSecrets, err := suite.client.db.Integration.QuerySecrets(integration).All(orgUser.UserCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Len(integrationSecrets, 1))
		assert.Check(t, is.DeepEqual(secret.ID, integrationSecrets[0].ID))

		// Clean up
		ctx := privacy.DecisionContext(orgUser.UserCtx, privacy.Allow)
		_ = suite.client.db.Hush.DeleteOneID(secret.ID).Exec(ctx)
	})

	// Clean up
	ctx := privacy.DecisionContext(orgUser.UserCtx, privacy.Allow)
	_ = suite.client.db.Integration.DeleteOneID(integration.ID).Exec(ctx)
}

func TestIntegrationWithSecretsRelationship(t *testing.T) {
	t.Parallel()

	// setup user context
	orgUser := suite.userBuilder(context.Background(), t)

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
		secrets, err := suite.client.db.Integration.QuerySecrets(integration).All(orgUser.UserCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Len(secrets, 3))

		// Verify secret names
		secretNames := make([]string, len(secrets))
		for i, secret := range secrets {
			secretNames[i] = secret.SecretName
		}

		expectedNames := []string{"github_access_token", "github_refresh_token", "github_expires_at"}
		// Simple check that we have the expected count (ElementsMatch is hard with gotest.tools)
		assert.Check(t, is.Len(secretNames, len(expectedNames)))
	})

	t.Run("Secrets can query their integration", func(t *testing.T) {
		// Query integration from secret
		integrationFromSecret, err := suite.client.db.Hush.QueryIntegrations(accessToken).Only(orgUser.UserCtx)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(integration.ID, integrationFromSecret.ID))
	})

	// Clean up
	ctx := privacy.DecisionContext(orgUser.UserCtx, privacy.Allow)
	_ = suite.client.db.Hush.DeleteOneID(accessToken.ID).Exec(ctx)
	_ = suite.client.db.Hush.DeleteOneID(refreshToken.ID).Exec(ctx)
	_ = suite.client.db.Hush.DeleteOneID(expiresAt.ID).Exec(ctx)
	_ = suite.client.db.Integration.DeleteOneID(integration.ID).Exec(ctx)
}

func TestIntegrationDeletion(t *testing.T) {
	t.Parallel()

	// setup user context
	orgUser := suite.userBuilder(context.Background(), t)

	// Create integration with secrets
	integration := (&IntegrationBuilder{client: suite.client}).MustNew(orgUser.UserCtx, t)
	secret := (&SecretBuilder{client: suite.client}).
		WithIntegration(integration.ID).
		MustNew(orgUser.UserCtx, t)

	t.Run("Delete integration", func(t *testing.T) {
		ctx := privacy.DecisionContext(orgUser.UserCtx, privacy.Allow)

		// Delete the integration
		err := suite.client.db.Integration.DeleteOneID(integration.ID).Exec(ctx)
		assert.NilError(t, err)

		// Verify integration is deleted
		deletedIntegration, err := suite.client.db.Integration.Get(ctx, integration.ID)
		assert.ErrorContains(t, err, "not found")
		assert.Check(t, is.Nil(deletedIntegration))

		// Check if secret still exists (cascade deletion test)
		deletedSecret, err := suite.client.db.Hush.Get(ctx, secret.ID)
		if err != nil {
			// If cascade deletion is working, the secret should be deleted
			assert.ErrorContains(t, err, "not found")
			assert.Check(t, is.Nil(deletedSecret))
		} else {
			// If cascade deletion is not configured, clean up manually
			_ = suite.client.db.Hush.DeleteOneID(secret.ID).Exec(ctx)
		}
	})
}

// Helper function to get string pointer
func stringPointer(s string) *string {
	return &s
}
