//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated/hush"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keystore"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
)

// slotRowIDs returns the ids of every row stored under one slot for an installation
func slotRowIDs(t *testing.T, ctx context.Context, integrationID string, slot string) []string {
	t.Helper()

	ids, err := suite.Client.DB.Hush.Query().
		Where(hush.HasIntegrationsWith(integration.IDEQ(integrationID)), hush.SecretNameEQ(slot)).
		IDs(ctx)
	assert.NilError(t, err)

	return ids
}

func TestKeystoreCredentialSlots(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)
	ctx := th.SetInternalContext(org.UserCtx, suite.Client.DB)

	store, err := keystore.NewStore(suite.Client.DB)
	assert.NilError(t, err)

	installation, _ := newHarnessInstallation(t, ctx, testint.ModeRecurring)

	token := testint.Token.Connection().Credential.Name
	serviceAccount := testint.ServiceAccount.Connection().Credential.Name
	oauth := testint.OAuth.Connection().Credential.Name

	t.Run("LoadAllCredentials returns every slot with the newest row winning", func(t *testing.T) {
		assert.NilError(t, store.SaveCredential(ctx, installation, serviceAccount, testint.ServiceAccountCredentialSet("proj", "svc@example.com")))

		assert.NilError(t, suite.Client.DB.Hush.Create().
			SetOwnerID(org.OrganizationID).
			SetName(token).
			SetSecretName(token).
			SetCredentialSet(testint.TokenCredentialSet("newest-token")).
			AddIntegrationIDs(installation.ID).
			Exec(ctx))

		rows, err := store.LoadAllCredentials(ctx, installation)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(rows, 2))
		assertJSONEqual(t, `{"token":"newest-token"}`, string(rows[token].Data))
		assertJSONEqual(t, `{"projectId":"proj","serviceAccountEmail":"svc@example.com"}`, string(rows[serviceAccount].Data))
		assert.Assert(t, is.Len(slotRowIDs(t, ctx, installation.ID, token), 2))
	})

	t.Run("ReplaceCredentials updates changed slots in place and leaves unchanged slots alone", func(t *testing.T) {
		before, err := store.LoadAllCredentials(ctx, installation)
		assert.NilError(t, err)

		serviceAccountIDs := slotRowIDs(t, ctx, installation.ID, serviceAccount)

		next := map[string]integrationtypes.CredentialSet{
			token:          testint.TokenCredentialSet("replaced-token"),
			serviceAccount: before[serviceAccount],
		}
		assert.NilError(t, store.ReplaceCredentials(ctx, installation, before, next))

		rows, err := store.LoadAllCredentials(ctx, installation)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(rows, 2))
		assertJSONEqual(t, `{"token":"replaced-token"}`, string(rows[token].Data))
		assert.DeepEqual(t, serviceAccountIDs, slotRowIDs(t, ctx, installation.ID, serviceAccount))
		assert.Assert(t, is.Len(slotRowIDs(t, ctx, installation.ID, token), 2))
	})

	t.Run("ReplaceCredentials creates new slots and deletes every row of removed slots", func(t *testing.T) {
		before, err := store.LoadAllCredentials(ctx, installation)
		assert.NilError(t, err)

		next := map[string]integrationtypes.CredentialSet{
			serviceAccount: testint.ServiceAccountCredentialSet("proj", "svc@example.com"),
			oauth:          {Data: json.RawMessage(`{"access_token":"a"}`)},
		}
		assert.NilError(t, store.ReplaceCredentials(ctx, installation, before, next))

		rows, err := store.LoadAllCredentials(ctx, installation)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(rows, 2))
		assert.Assert(t, is.Len(slotRowIDs(t, ctx, installation.ID, token), 0))
		assert.Assert(t, is.Len(slotRowIDs(t, ctx, installation.ID, oauth), 1))
	})

	t.Run("ReplaceCredentials evicts pooled clients", func(t *testing.T) {
		builds := 0

		build := func(context.Context, integrationtypes.ConnectionInput) (any, error) {
			builds++

			return "client", nil
		}

		_, err := store.BuildClient(ctx, installation, token, "string", build, integrationtypes.CredentialSet{}, false)
		assert.NilError(t, err)
		_, err = store.BuildClient(ctx, installation, token, "string", build, integrationtypes.CredentialSet{}, false)
		assert.NilError(t, err)
		assert.Equal(t, 1, builds)

		before, err := store.LoadAllCredentials(ctx, installation)
		assert.NilError(t, err)

		assert.NilError(t, store.ReplaceCredentials(ctx, installation, before, map[string]integrationtypes.CredentialSet{
			serviceAccount: testint.ServiceAccountCredentialSet("proj", "svc@example.com"),
		}))

		_, err = store.BuildClient(ctx, installation, token, "string", build, integrationtypes.CredentialSet{}, false)
		assert.NilError(t, err)
		assert.Equal(t, 2, builds)
	})

	t.Run("ReplaceCredentials leaves rows outside the previous snapshot untouched", func(t *testing.T) {
		before, err := store.LoadAllCredentials(ctx, installation)
		assert.NilError(t, err)

		assert.NilError(t, store.SaveCredential(ctx, installation, token, testint.TokenCredentialSet("concurrent-token")))

		assert.NilError(t, store.ReplaceCredentials(ctx, installation, before, map[string]integrationtypes.CredentialSet{
			serviceAccount: before[serviceAccount],
		}))

		rows, err := store.LoadAllCredentials(ctx, installation)
		assert.NilError(t, err)
		assertJSONEqual(t, `{"token":"concurrent-token"}`, string(rows[token].Data))
	})

	t.Run("ReplaceCredentials does not overwrite an existing row for a slot outside previous", func(t *testing.T) {
		before, err := store.LoadAllCredentials(ctx, installation)
		assert.NilError(t, err)

		delete(before, token)

		assert.NilError(t, store.SaveCredential(ctx, installation, token, testint.TokenCredentialSet("direct-token")))

		tokenIDs := slotRowIDs(t, ctx, installation.ID, token)

		assert.NilError(t, store.ReplaceCredentials(ctx, installation, before, map[string]integrationtypes.CredentialSet{
			serviceAccount: before[serviceAccount],
			token:          testint.TokenCredentialSet("overwritten-token"),
		}))

		rows, err := store.LoadAllCredentials(ctx, installation)
		assert.NilError(t, err)
		assertJSONEqual(t, `{"token":"direct-token"}`, string(rows[token].Data))
		assert.DeepEqual(t, tokenIDs, slotRowIDs(t, ctx, installation.ID, token))
	})
}
