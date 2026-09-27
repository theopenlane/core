//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/v2/internal/ent/generated/hush"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keystore"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
)

// slotRowIDs returns the ids of every row stored under one slot for an installation
func slotRowIDs(t *testing.T, ctx context.Context, integrationID string, slot integrationtypes.CredentialSlotID) []string {
	t.Helper()

	ids, err := suite.Client.DB.Hush.Query().
		Where(hush.HasIntegrationsWith(integration.IDEQ(integrationID)), hush.SecretNameEQ(slot.String())).
		IDs(ctx)
	require.NoError(t, err)

	return ids
}

func TestKeystoreCredentialSlots(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)
	ctx := th.SetContext(org.UserCtx, suite.Client.DB)

	store, err := keystore.NewStore(suite.Client.DB)
	require.NoError(t, err)

	installation, _ := newHarnessInstallation(t, ctx, testint.ModeRecurring)

	token := testint.TokenCredential.ID()
	serviceAccount := testint.ServiceAccountCredential.ID()

	t.Run("LoadAllCredentials returns every slot with the newest row winning", func(t *testing.T) {
		require.NoError(t, store.SaveCredential(ctx, installation, serviceAccount, testint.ServiceAccountCredentialSet("proj", "svc@example.com")))

		require.NoError(t, suite.Client.DB.Hush.Create().
			SetOwnerID(org.OrganizationID).
			SetName(token.String()).
			SetSecretName(token.String()).
			SetCredentialSet(testint.TokenCredentialSet("newest-token")).
			AddIntegrationIDs(installation.ID).
			Exec(ctx))

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.JSONEq(t, `{"token":"newest-token"}`, string(rows[token].Data))
		require.JSONEq(t, `{"projectId":"proj","serviceAccountEmail":"svc@example.com"}`, string(rows[serviceAccount].Data))
		require.Len(t, slotRowIDs(t, ctx, installation.ID, token), 2)
	})

	t.Run("ReplaceCredentials updates changed slots in place and leaves unchanged slots alone", func(t *testing.T) {
		before, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)

		serviceAccountIDs := slotRowIDs(t, ctx, installation.ID, serviceAccount)

		next := map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			token:          testint.TokenCredentialSet("replaced-token"),
			serviceAccount: before[serviceAccount],
		}
		require.NoError(t, store.ReplaceCredentials(ctx, installation, before, next))

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.JSONEq(t, `{"token":"replaced-token"}`, string(rows[token].Data))
		require.Equal(t, serviceAccountIDs, slotRowIDs(t, ctx, installation.ID, serviceAccount))
		require.Len(t, slotRowIDs(t, ctx, installation.ID, token), 2)
	})

	t.Run("ReplaceCredentials creates new slots and deletes every row of removed slots", func(t *testing.T) {
		before, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)

		next := map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			serviceAccount:               testint.ServiceAccountCredentialSet("proj", "svc@example.com"),
			testint.OAuthCredential.ID(): {Data: json.RawMessage(`{"access_token":"a"}`)},
		}
		require.NoError(t, store.ReplaceCredentials(ctx, installation, before, next))

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.Empty(t, slotRowIDs(t, ctx, installation.ID, token))
		require.Len(t, slotRowIDs(t, ctx, installation.ID, testint.OAuthCredential.ID()), 1)
	})

	t.Run("ReplaceCredentials evicts pooled clients", func(t *testing.T) {
		builds := 0

		registration := integrationtypes.ClientRegistration{
			Ref: integrationtypes.NewClientRef[string]().ID(),
			Build: func(context.Context, integrationtypes.ClientBuildRequest) (any, error) {
				builds++

				return "client", nil
			},
		}

		_, err := store.BuildClient(ctx, installation, registration, nil, nil, false)
		require.NoError(t, err)
		_, err = store.BuildClient(ctx, installation, registration, nil, nil, false)
		require.NoError(t, err)
		require.Equal(t, 1, builds)

		before, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)

		require.NoError(t, store.ReplaceCredentials(ctx, installation, before, map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			serviceAccount: testint.ServiceAccountCredentialSet("proj", "svc@example.com"),
		}))

		_, err = store.BuildClient(ctx, installation, registration, nil, nil, false)
		require.NoError(t, err)
		require.Equal(t, 2, builds)
	})

	t.Run("ReplaceCredentials leaves rows outside the previous snapshot untouched", func(t *testing.T) {
		before, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)

		require.NoError(t, store.SaveCredential(ctx, installation, token, testint.TokenCredentialSet("concurrent-token")))

		require.NoError(t, store.ReplaceCredentials(ctx, installation, before, map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			serviceAccount: before[serviceAccount],
		}))

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.JSONEq(t, `{"token":"concurrent-token"}`, string(rows[token].Data))
	})

	t.Run("ReplaceCredentials does not overwrite an existing row for a slot outside previous", func(t *testing.T) {
		before, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)

		delete(before, token)

		require.NoError(t, store.SaveCredential(ctx, installation, token, testint.TokenCredentialSet("direct-token")))

		tokenIDs := slotRowIDs(t, ctx, installation.ID, token)

		require.NoError(t, store.ReplaceCredentials(ctx, installation, before, map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			serviceAccount: before[serviceAccount],
			token:          testint.TokenCredentialSet("overwritten-token"),
		}))

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.JSONEq(t, `{"token":"direct-token"}`, string(rows[token].Data))
		require.Equal(t, tokenIDs, slotRowIDs(t, ctx, installation.ID, token))
	})
}
