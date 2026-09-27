//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"
	"golang.org/x/sync/errgroup"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hush"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keystore"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// tokenCred is a previous version of the current token type whose token was numeric
type tokenCred struct {
	Token int `json:"token"`
}

// straySlotCred is a slot a previous version declared and the current definition does not
type straySlotCred struct {
	Value string `json:"value"`
}

// serviceAccountCred is a previous version of the current service-account type that had no email field
type serviceAccountCred struct {
	ProjectID string `json:"projectId" jsonschema:"required"`
}

// unreplacedCred is a retired slot that nothing in the current definition replaces
type unreplacedCred struct {
	Key string `json:"key"`
}

var (
	numericTokenRef          = integrationtypes.NewCredentialRef[tokenCred]()
	straySlotRef             = integrationtypes.NewCredentialRef[straySlotCred]()
	partialServiceAccountRef = integrationtypes.NewCredentialRef[serviceAccountCred]()
	unreplacedRef            = integrationtypes.NewCredentialRef[unreplacedCred]()
)

// previousDefinition builds an earlier version of the shared test definition with one connection over the given slots
func previousDefinition(primary integrationtypes.CredentialSlot, extra ...integrationtypes.CredentialSlot) registry.Builder {
	return func() (integrationtypes.Definition, error) {
		def := integrationtypes.Definition{
			DefinitionSpec: integrationtypes.DefinitionSpec{
				ID:          testint.DefinitionID.ID(),
				DisplayName: "Test Integration",
				Active:      true,
			},
		}

		connection := integrationtypes.ConnectionRegistration{CredentialRef: primary.ID()}

		for _, slot := range append([]integrationtypes.CredentialSlot{primary}, extra...) {
			def.CredentialRegistrations = append(def.CredentialRegistrations, integrationtypes.CredentialRegistration{Ref: slot})
			connection.CredentialRefs = append(connection.CredentialRefs, slot.ID())
		}

		def.Connections = []integrationtypes.ConnectionRegistration{connection}

		return def, nil
	}
}

// installUnder installs the test definition through a runtime running an earlier version of it, with every recurring loop disabled so only the caller triggers an upgrade, and returns the installation and that version
func installUnder(t *testing.T, ctx context.Context, builder registry.Builder, primary integrationtypes.CredentialSlotID, credentials map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet) (*ent.Integration, string) {
	t.Helper()

	previous, err := gala.NewGala(context.Background(), gala.Config{DispatchMode: gala.DispatchModeInMemory, WorkerCount: 1})
	require.NoError(t, err)
	t.Cleanup(func() { _ = previous.Close() })

	store, err := keystore.NewStore(suite.Client.DB)
	require.NoError(t, err)

	rt, err := intruntime.New(intruntime.Config{DB: suite.Client.DB, Gala: previous, Keystore: store, DefinitionBuilders: []registry.Builder{builder}})
	require.NoError(t, err)

	def, ok := rt.Registry().Definition(testint.DefinitionID.ID())
	require.True(t, ok)

	ownerID, err := auth.GetOrganizationIDFromContext(ctx)
	require.NoError(t, err)

	installation, _, err := rt.EnsureInstallation(ctx, ownerID, "", def)
	require.NoError(t, err)

	first := credentials[primary]
	require.NoError(t, rt.Reconcile(ctx, installation, testint.ModeInput("none"), primary, &first, nil))

	for slot, credential := range credentials {
		if slot == primary {
			continue
		}

		require.NoError(t, rt.Reconcile(ctx, reloadIntegration(t, ctx, installation.ID), nil, slot, &credential, nil))
	}

	return reloadIntegration(t, ctx, installation.ID), rt.Registry().Version(testint.DefinitionID.ID())
}

func TestInstallationUpgrade(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)
	ctx := th.SetContext(org.UserCtx, suite.Client.DB)

	store, err := keystore.NewStore(suite.Client.DB)
	require.NoError(t, err)

	def, ok := suite.IntegrationsRT.Definition(testint.DefinitionID.ID())
	require.True(t, ok)

	current := suite.IntegrationsRT.Registry().Version(def.ID)
	require.NotEmpty(t, current)

	t.Run("a slot retired by the current definition is upgraded inline on first use", func(t *testing.T) {
		installation, previous := installUnder(t, ctx, previousDefinition(testint.LegacyTokenCredential), testint.LegacyTokenCredential.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			testint.LegacyTokenCredential.ID(): {Data: json.RawMessage(`{"accessToken":"legacy-token"}`)},
		})
		require.NotEmpty(t, previous)
		require.NotEqual(t, current, previous)
		require.Equal(t, previous, installation.DefinitionVersion)

		assessment, err := suite.IntegrationsRT.RunHealthAssessment(ctx, installation)
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"token":"legacy-token"}`, string(rows[testint.TokenCredential.ID()].Data))
		require.Empty(t, slotRowIDs(t, ctx, installation.ID, testint.LegacyTokenCredential.ID()))

		reloaded := reloadIntegration(t, ctx, installation.ID)
		require.Equal(t, current, reloaded.DefinitionVersion)

		state, err := def.ProviderState(reloaded.ProviderState)
		require.NoError(t, err)
		require.Equal(t, testint.TokenCredential.ID(), state.CredentialRef)

		before := slotRowIDs(t, ctx, installation.ID, testint.TokenCredential.ID())

		_, err = suite.IntegrationsRT.RunHealthAssessment(ctx, reloaded)
		require.NoError(t, err)
		require.Equal(t, before, slotRowIDs(t, ctx, installation.ID, testint.TokenCredential.ID()))
	})

	t.Run("a stored type the current schema rejects fails the upgrade and marks the installation errored", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		installation, previous := installUnder(t, subCtx, previousDefinition(numericTokenRef), numericTokenRef.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			numericTokenRef.ID(): {Data: json.RawMessage(`{"token":1}`)},
		})

		_, err := suite.IntegrationsRT.RunHealthAssessment(subCtx, installation)
		require.ErrorIs(t, err, intruntime.ErrInstallationUpgradeFailed)
		require.ErrorIs(t, err, intruntime.ErrCredentialInvalid)

		_, unhealthy := integrationtypes.UnhealthyFrom(err)
		require.True(t, unhealthy)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, previous, reloaded.DefinitionVersion)
		require.Equal(t, enums.IntegrationStatusErrored, reloaded.Status)
		require.Equal(t, 1, integrationNotificationCount(t, subCtx, installation.OwnerID, integrationReconfigurationRequiredObjectType))

		rows, err := store.LoadAllCredentials(subCtx, installation)
		require.NoError(t, err)
		require.JSONEq(t, `{"token":1}`, string(rows[testint.TokenCredential.ID()].Data))
	})

	t.Run("a backfilled slot is completed from the installation during the upgrade", func(t *testing.T) {
		installation, previous := installUnder(t, ctx, previousDefinition(partialServiceAccountRef), partialServiceAccountRef.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			partialServiceAccountRef.ID(): {Data: json.RawMessage(`{"projectId":"p"}`)},
		})
		require.NotEqual(t, current, previous)

		assessment, err := suite.IntegrationsRT.RunHealthAssessment(ctx, installation)
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"projectId":"p","serviceAccountEmail":"`+installation.ID+`@backfilled.example.com"}`, string(rows[testint.ServiceAccountCredential.ID()].Data))
		require.Equal(t, current, reloadIntegration(t, ctx, installation.ID).DefinitionVersion)
	})

	t.Run("reconnecting with a fresh credential recovers an installation whose upgrade failed", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		installation, _ := installUnder(t, subCtx, previousDefinition(numericTokenRef), numericTokenRef.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			numericTokenRef.ID(): {Data: json.RawMessage(`{"token":1}`)},
		})

		_, err := suite.IntegrationsRT.RunHealthAssessment(subCtx, installation)
		require.ErrorIs(t, err, intruntime.ErrInstallationUpgradeFailed)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusErrored, reloaded.Status)
		require.Equal(t, 1, integrationNotificationCount(t, subCtx, installation.OwnerID, integrationReconfigurationRequiredObjectType))

		cred := testint.TokenCredentialSet("fresh")
		require.NoError(t, suite.IntegrationsRT.Reconcile(subCtx, reloaded, nil, testint.TokenCredential.ID(), &cred, nil))

		recovered := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusConnected, recovered.Status)
		require.Equal(t, current, recovered.DefinitionVersion)
		require.Equal(t, 1, integrationNotificationCount(t, subCtx, installation.OwnerID, integrationReconnectedObjectType))

		rows, err := store.LoadAllCredentials(subCtx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"token":"fresh"}`, string(rows[testint.TokenCredential.ID()].Data))
	})

	t.Run("disconnect succeeds when the persisted ref is a retired slot nothing replaces", func(t *testing.T) {
		installation, _ := installUnder(t, ctx, previousDefinition(unreplacedRef), unreplacedRef.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			unreplacedRef.ID(): {Data: json.RawMessage(`{"key":"k"}`)},
		})

		_, err := suite.IntegrationsRT.RunHealthAssessment(ctx, installation)
		reloaded := reloadIntegration(t, ctx, installation.ID)
		require.True(t, err != nil || reloaded.Status == enums.IntegrationStatusErrored)

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.JSONEq(t, `{"key":"k"}`, string(rows[unreplacedRef.ID()].Data))

		_, err = suite.IntegrationsRT.Disconnect(ctx, reloaded)
		require.NoError(t, err)
		require.False(t, integrationVisible(t, ctx, installation.ID))

		remaining, err := suite.Client.DB.Hush.Query().Where(hush.HasIntegrationsWith(integration.IDEQ(installation.ID))).Count(ctx)
		require.NoError(t, err)
		require.Zero(t, remaining)
	})

	t.Run("a crash between the credential move and the ref repoint converges on rerun", func(t *testing.T) {
		installation, _ := installUnder(t, ctx, previousDefinition(testint.LegacyTokenCredential), testint.LegacyTokenCredential.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			testint.LegacyTokenCredential.ID(): {Data: json.RawMessage(`{"accessToken":"legacy-token"}`)},
		})

		stored, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, stored, 1)

		require.NoError(t, store.ReplaceCredentials(ctx, installation, stored, map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			testint.TokenCredential.ID(): testint.TokenCredentialSet("legacy-token"),
		}))
		require.Empty(t, slotRowIDs(t, ctx, installation.ID, testint.LegacyTokenCredential.ID()))

		state, err := def.ProviderState(installation.ProviderState)
		require.NoError(t, err)
		require.Equal(t, testint.LegacyTokenCredential.ID(), state.CredentialRef)

		assessment, err := suite.IntegrationsRT.RunHealthAssessment(ctx, installation)
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)

		reloaded := reloadIntegration(t, ctx, installation.ID)
		require.Equal(t, current, reloaded.DefinitionVersion)

		state, err = def.ProviderState(reloaded.ProviderState)
		require.NoError(t, err)
		require.Equal(t, testint.TokenCredential.ID(), state.CredentialRef)

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"token":"legacy-token"}`, string(rows[testint.TokenCredential.ID()].Data))
	})

	t.Run("concurrent upgrades of one installation both succeed and converge", func(t *testing.T) {
		installation, _ := installUnder(t, ctx, previousDefinition(testint.LegacyTokenCredential), testint.LegacyTokenCredential.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			testint.LegacyTokenCredential.ID(): {Data: json.RawMessage(`{"accessToken":"legacy-token"}`)},
		})

		first := reloadIntegration(t, ctx, installation.ID)
		second := reloadIntegration(t, ctx, installation.ID)

		var group errgroup.Group

		group.Go(func() error {
			_, err := suite.IntegrationsRT.RunHealthAssessment(ctx, first)

			return err
		})
		group.Go(func() error {
			_, err := suite.IntegrationsRT.RunHealthAssessment(ctx, second)

			return err
		})
		require.NoError(t, group.Wait())

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"token":"legacy-token"}`, string(rows[testint.TokenCredential.ID()].Data))
		require.Empty(t, slotRowIDs(t, ctx, installation.ID, testint.LegacyTokenCredential.ID()))

		reloaded := reloadIntegration(t, ctx, installation.ID)
		require.Equal(t, current, reloaded.DefinitionVersion)

		state, err := def.ProviderState(reloaded.ProviderState)
		require.NoError(t, err)
		require.Equal(t, testint.TokenCredential.ID(), state.CredentialRef)
	})

	t.Run("a slot the current definition no longer declares is left in place", func(t *testing.T) {
		installation, _ := installUnder(t, ctx, previousDefinition(testint.LegacyTokenCredential, straySlotRef), testint.LegacyTokenCredential.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			testint.LegacyTokenCredential.ID(): {Data: json.RawMessage(`{"accessToken":"legacy-token"}`)},
			straySlotRef.ID():                  {Data: json.RawMessage(`{"value":"kept"}`)},
		})

		stored, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, stored, 2)

		_, err = suite.IntegrationsRT.RunHealthAssessment(ctx, installation)
		require.NoError(t, err)

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.JSONEq(t, `{"value":"kept"}`, string(rows[straySlotRef.ID()].Data))
		require.JSONEq(t, `{"token":"legacy-token"}`, string(rows[testint.TokenCredential.ID()].Data))
		require.Equal(t, current, reloadIntegration(t, ctx, installation.ID).DefinitionVersion)
	})

	t.Run("a new connection is stamped with the current version", func(t *testing.T) {
		installation, _ := newHarnessInstallation(t, ctx, testint.ModeRecurring)
		require.Equal(t, current, installation.DefinitionVersion)
	})
}
