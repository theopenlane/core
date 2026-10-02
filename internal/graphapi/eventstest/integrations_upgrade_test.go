//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"
	"golang.org/x/sync/errgroup"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hush"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationwebhook"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keystore"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// tokenCred is a previous version of the current token type whose token was numeric
type tokenCred struct {
	Token int `json:"token"`
}

// straySlotCred is a slot a previous version declared and the current definition does not
type straySlotCred struct {
	Value string `json:"value"`
}

// serviceAccountCred is a previous service-account type with no email field
type serviceAccountCred struct {
	ProjectID string `json:"projectId" jsonschema:"required"`
}

// unreplacedCred is a retired slot that nothing in the current definition replaces
type unreplacedCred struct {
	Key string `json:"key"`
}

// zoneInput is the user input layout an earlier definition version stored
type zoneInput struct {
	Zone string `json:"zone"`
}

// regionInput is the current user input layout, requiring region and token
type regionInput struct {
	Region string `json:"region" jsonschema:"required"`
	Token  string `json:"token" jsonschema:"required"`
}

// retiredSync is the config of the operation an earlier definition version declared
type retiredSync struct{}

// renamedSync is the config of the operation replacing retiredSync
type renamedSync struct{}

// renameEvent is the payload of the events on the renamed webhook contract
type renameEvent struct{}

// backfilledToken is the token the current user input backfills when empty
const backfilledToken = "x"

// suiteQueueName is the durable gala queue the suite harness runs
const suiteQueueName = "graphapi_integration_test"

var (
	numericTokenRef          = integrationtypes.NewCredentialRef[tokenCred](testint.TokenCredential.String())
	straySlotRef             = integrationtypes.NewCredentialRef[straySlotCred]("straySlot")
	partialServiceAccountRef = integrationtypes.NewCredentialRef[serviceAccountCred](testint.ServiceAccountCredential.String())
	unreplacedRef            = integrationtypes.NewCredentialRef[unreplacedCred]("unreplaced")

	zoneInputRef   = integrationtypes.UserInputRefOf[zoneInput]()
	regionInputRef = integrationtypes.UserInputRefOf[regionInput]().Upgraded(upgradeRegionInput)

	retiredSyncOp = integrationtypes.OperationRefOf[retiredSync]()
	renamedSyncOp = integrationtypes.OperationRefOf[renamedSync]()

	retiredEventsWebhook = integrationtypes.NewWebhookRef("events.v1")
	renamedEventsWebhook = integrationtypes.NewWebhookRef("events.v2")
	renameEventA         = integrationtypes.NewWebhookEventRef[renameEvent]("a")
	renameEventB         = integrationtypes.NewWebhookEventRef[renameEvent]("b")
)

// upgradeRegionInput maps a document stored under the zone layout onto the region layout and fills an empty token
func upgradeRegionInput(_ context.Context, _ integrationtypes.InstallationRequest, from string, stored json.RawMessage) (regionInput, error) {
	var (
		current regionInput
		err     error
	)

	switch from {
	case zoneInputRef.Name():
		var old zoneInput

		old, err = jsonx.Decode[zoneInput](stored)
		current = regionInput{Region: old.Zone}
	default:
		current, err = jsonx.Decode[regionInput](stored)
	}

	if err != nil {
		return regionInput{}, err
	}

	if current.Token == "" {
		current.Token = backfilledToken
	}

	return current, nil
}

// retiredSlot pairs a retired credential slot id with its schema
type retiredSlot struct {
	id     integrationtypes.CredentialSlotID
	schema json.RawMessage
}

// slotOf pairs a typed credential ref with the schema it reflects
func slotOf[T any](ref integrationtypes.CredentialRef[T]) retiredSlot {
	return retiredSlot{id: ref.ID(), schema: ref.Schema()}
}

// syncOperation returns a no-op operation registration replacing retired names
func syncOperation[Cfg any](op integrationtypes.OperationRef[Cfg], policy integrationtypes.ExecutionPolicy, replaces ...string) integrationtypes.OperationRegistration {
	return op.Policy(policy).Registration(testint.DefinitionID, integrationtypes.OperationRegistration{
		Replaces: replaces,
		Handle:   func(context.Context, integrationtypes.OperationRequest) (json.RawMessage, error) { return nil, nil },
	})
}

// eventsWebhook returns a webhook registration accepting events, replacing retired names
func eventsWebhook(webhook integrationtypes.WebhookRef, replaces []string, events ...integrationtypes.WebhookEventRef[renameEvent]) integrationtypes.WebhookRegistration {
	return webhook.Registration(integrationtypes.WebhookRegistration{
		Replaces: replaces,
		Event: func(req integrationtypes.WebhookInboundRequest) (integrationtypes.WebhookReceivedEvent, error) {
			return integrationtypes.WebhookReceivedEvent{Name: string(req.Payload), Payload: req.Payload}, nil
		},
		Events: lo.Map(events, func(event integrationtypes.WebhookEventRef[renameEvent], _ int) integrationtypes.WebhookEventRegistration {
			return event.Registration(testint.DefinitionID, integrationtypes.WebhookEventRegistration{
				Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error { return nil },
			})
		}),
	})
}

// previousDefinition returns an earlier version of the shared test definition
func previousDefinition(primary retiredSlot, extra ...retiredSlot) registry.Builder {
	return func() (integrationtypes.Definition, error) {
		def := integrationtypes.Definition{
			DefinitionSpec: integrationtypes.DefinitionSpec{
				ID:          testint.DefinitionID.ID(),
				DisplayName: "Test Integration",
				Active:      true,
			},
			HealthCheck: integrationtypes.CredentialHealthCheck(func(context.Context, integrationtypes.OperationRequest) (json.RawMessage, error) {
				return json.RawMessage(`{"ok":true}`), nil
			}),
		}

		connection := integrationtypes.ConnectionRegistration{CredentialRef: primary.id}

		for _, slot := range append([]retiredSlot{primary}, extra...) {
			def.CredentialRegistrations = append(def.CredentialRegistrations, integrationtypes.CredentialRegistration{Ref: slot.id, Schema: slot.schema})
			connection.CredentialRefs = append(connection.CredentialRefs, slot.id)
		}

		def.Connections = []integrationtypes.ConnectionRegistration{connection}

		return def, nil
	}
}

// definitionOver returns the shared test definition with shape applied over the token slot
func definitionOver(shape func(def *integrationtypes.Definition)) registry.Builder {
	return func() (integrationtypes.Definition, error) {
		def, err := previousDefinition(slotOf(testint.TokenCredential))()
		if err != nil {
			return integrationtypes.Definition{}, err
		}

		shape(&def)

		return def, nil
	}
}

// runtimeFor returns a runtime on an in-memory gala running one definition version
func runtimeFor(t *testing.T, builder registry.Builder) *intruntime.Runtime {
	t.Helper()

	instance, err := gala.NewGala(context.Background(), gala.Config{DispatchMode: gala.DispatchModeInMemory, WorkerCount: 1})
	require.NoError(t, err)
	t.Cleanup(func() { _ = instance.Close() })

	return runtimeOn(t, instance, builder)
}

// queuedRuntimeFor returns a runtime on the suite's durable gala queue with no workers
func queuedRuntimeFor(t *testing.T, builder registry.Builder) *intruntime.Runtime {
	t.Helper()

	instance, err := gala.NewGala(context.Background(), gala.Config{DispatchMode: gala.DispatchModeDurable, ConnectionURI: suite.TF.URI, QueueName: suiteQueueName, WorkerCount: 1})
	require.NoError(t, err)
	t.Cleanup(func() { _ = instance.Close() })

	return runtimeOn(t, instance, builder)
}

// runtimeOn returns a runtime on instance running one definition version
func runtimeOn(t *testing.T, instance *gala.Gala, builder registry.Builder) *intruntime.Runtime {
	t.Helper()

	store, err := keystore.NewStore(suite.Client.DB)
	require.NoError(t, err)

	rt, err := intruntime.New(intruntime.Config{DB: suite.Client.DB, Gala: instance, Keystore: store, DefinitionBuilders: []registry.Builder{builder}})
	require.NoError(t, err)

	return rt
}

// seedRetiredLoop queues one future reconcile cycle under a retired operation name
func seedRetiredLoop(t *testing.T, ctx context.Context, installation *ent.Integration, operationName string) {
	t.Helper()

	oc := integrationtypes.NewOperationContext(installation.OwnerID, operationName, integrationtypes.IntegrationSource{
		IntegrationID: installation.ID,
		DefinitionID:  installation.DefinitionID,
		RunType:       enums.IntegrationRunTypeReconcile,
	})

	emitCtx, headers := intobvs.EmitContext(ctx, oc)
	headers.SkipUniqueKey = true
	headers.ScheduledAt = lo.ToPtr(time.Now().Add(time.Hour))

	_, err := suite.GalaRuntime.EmitWithHeaders(emitCtx, operations.ReconcileTopic.Name, operations.ReconcileEnvelope{OperationContext: oc}, headers)
	require.NoError(t, err)
}

// installOn returns the installation of the shared test definition through rt
func installOn(t *testing.T, ctx context.Context, rt *intruntime.Runtime, userInput json.RawMessage, operationConfig map[string]json.RawMessage, primary integrationtypes.CredentialSlotID, credential integrationtypes.CredentialSet) *ent.Integration {
	t.Helper()

	def, ok := rt.Registry().Definition(testint.DefinitionID.ID())
	require.True(t, ok)

	ownerID, err := auth.GetOrganizationIDFromContext(ctx)
	require.NoError(t, err)

	installation, _, err := rt.EnsureInstallation(ctx, ownerID, "", def)
	require.NoError(t, err)

	require.NoError(t, rt.Reconcile(ctx, installation, userInput, operationConfig, primary, &credential, nil))

	return reloadIntegration(t, ctx, installation.ID)
}

// installUnder returns the installation and version from an earlier definition runtime
func installUnder(t *testing.T, ctx context.Context, builder registry.Builder, primary integrationtypes.CredentialSlotID, credentials map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet) (*ent.Integration, string) {
	t.Helper()

	rt := runtimeFor(t, builder)

	installation := installOn(t, ctx, rt, nil, nil, primary, credentials[primary])

	for slot, credential := range credentials {
		if slot == primary {
			continue
		}

		require.NoError(t, rt.Reconcile(ctx, reloadIntegration(t, ctx, installation.ID), nil, nil, slot, &credential, nil))
	}

	return reloadIntegration(t, ctx, installation.ID), rt.Registry().Version(testint.DefinitionID.ID())
}

// endpointRows returns the installation's webhook endpoint rows, excluding dedupe rows
func endpointRows(t *testing.T, ctx context.Context, integrationID string) []*ent.IntegrationWebhook {
	t.Helper()

	rows, err := suite.Client.DB.IntegrationWebhook.Query().
		Where(integrationwebhook.IntegrationIDEQ(integrationID), integrationwebhook.ExternalEventIDIsNil()).
		All(ctx)
	require.NoError(t, err)

	return rows
}

func TestInstallationUpgrade(t *testing.T) {
	org := suite.UserBuilder(context.Background(), t)
	ctx := th.SetContext(org.UserCtx, suite.Client.DB)

	store, err := keystore.NewStore(suite.Client.DB)
	require.NoError(t, err)

	def, ok := suite.IntegrationsRT.Registry().Definition(testint.DefinitionID.ID())
	require.True(t, ok)

	current := suite.IntegrationsRT.Registry().Version(def.ID)
	require.NotEmpty(t, current)

	t.Run("a slot retired by the current definition is upgraded inline on first use", func(t *testing.T) {
		installation, previous := installUnder(t, ctx, previousDefinition(slotOf(testint.LegacyTokenCredential)), testint.LegacyTokenCredential.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
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

		installation, previous := installUnder(t, subCtx, previousDefinition(slotOf(numericTokenRef)), numericTokenRef.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
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
		installation, previous := installUnder(t, ctx, previousDefinition(slotOf(partialServiceAccountRef)), partialServiceAccountRef.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
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

		installation, _ := installUnder(t, subCtx, previousDefinition(slotOf(numericTokenRef)), numericTokenRef.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			numericTokenRef.ID(): {Data: json.RawMessage(`{"token":1}`)},
		})

		_, err := suite.IntegrationsRT.RunHealthAssessment(subCtx, installation)
		require.ErrorIs(t, err, intruntime.ErrInstallationUpgradeFailed)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusErrored, reloaded.Status)
		require.Equal(t, 1, integrationNotificationCount(t, subCtx, installation.OwnerID, integrationReconfigurationRequiredObjectType))

		cred := testint.TokenCredentialSet("fresh")
		require.NoError(t, suite.IntegrationsRT.Reconcile(subCtx, reloaded, nil, nil, testint.TokenCredential.ID(), &cred, nil))

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
		installation, _ := installUnder(t, ctx, previousDefinition(slotOf(unreplacedRef)), unreplacedRef.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
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
		installation, _ := installUnder(t, ctx, previousDefinition(slotOf(testint.LegacyTokenCredential)), testint.LegacyTokenCredential.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
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
		installation, _ := installUnder(t, ctx, previousDefinition(slotOf(testint.LegacyTokenCredential)), testint.LegacyTokenCredential.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
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
		installation, _ := installUnder(t, ctx, previousDefinition(slotOf(testint.LegacyTokenCredential), slotOf(straySlotRef)), testint.LegacyTokenCredential.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
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

	t.Run("user input is renamed and backfilled on upgrade", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		previous := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.UserInput = zoneInputRef.Registration()
		}))
		installation := installOn(t, subCtx, previous, json.RawMessage(`{"zone":"eu"}`), nil, testint.TokenCredential.ID(), testint.TokenCredentialSet("token"))
		require.Equal(t, zoneInputRef.Name(), installation.UserInput.Layout)
		require.JSONEq(t, `{"zone":"eu"}`, string(installation.UserInput.Data))

		renamed := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.UserInput = regionInputRef.Registration()
		}))
		renamedVersion := renamed.Registry().Version(testint.DefinitionID.ID())
		require.NotEqual(t, renamedVersion, installation.DefinitionVersion)

		assessment, err := renamed.RunHealthAssessment(subCtx, installation)
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)
		require.Equal(t, regionInputRef.Name(), installation.UserInput.Layout)
		require.JSONEq(t, `{"region":"eu","token":"x"}`, string(installation.UserInput.Data))
		require.Equal(t, renamedVersion, installation.DefinitionVersion)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, regionInputRef.Name(), reloaded.UserInput.Layout)
		require.JSONEq(t, `{"region":"eu","token":"x"}`, string(reloaded.UserInput.Data))
		require.Equal(t, renamedVersion, reloaded.DefinitionVersion)
		require.Equal(t, enums.IntegrationStatusConnected, reloaded.Status)
		require.Zero(t, integrationNotificationCount(t, subCtx, installation.OwnerID, integrationReconfigurationRequiredObjectType))
	})

	t.Run("an operation rename moves health keys and run history", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		previous := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.Operations = []integrationtypes.OperationRegistration{syncOperation(retiredSyncOp, integrationtypes.ExecutionPolicy{Inline: true})}
		}))
		installation := installOn(t, subCtx, previous, nil, nil, testint.TokenCredential.ID(), testint.TokenCredentialSet("token"))

		retired, err := previous.Registry().Operation(testint.DefinitionID.ID(), retiredSyncOp.Name())
		require.NoError(t, err)

		finished, err := operations.CreatePendingRun(subCtx, suite.Client.DB, installation, retired, enums.IntegrationRunTypeManual, nil)
		require.NoError(t, err)
		require.NoError(t, operations.CompleteRun(subCtx, suite.Client.DB, finished.ID, finished.StartedAt, operations.RunResult{}))

		_, err = operations.CreatePendingRun(subCtx, suite.Client.DB, installation, retired, enums.IntegrationRunTypeManual, nil)
		require.NoError(t, err)

		finishedAt, err := operations.LastSuccessfulRunAt(subCtx, suite.Client.DB, installation.ID, retiredSyncOp.Name())
		require.NoError(t, err)
		require.NotNil(t, finishedAt)

		require.NoError(t, suite.Client.DB.Integration.UpdateOneID(installation.ID).
			SetHealth(models.IntegrationHealth{UnhealthyOperations: map[string]string{retiredSyncOp.Name(): "boom", "gone": "x"}}).
			Exec(subCtx))

		renamed := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.Operations = []integrationtypes.OperationRegistration{syncOperation(renamedSyncOp, integrationtypes.ExecutionPolicy{Inline: true}, retiredSyncOp.Name())}
		}))

		assessment, err := renamed.RunHealthAssessment(subCtx, reloadIntegration(t, subCtx, installation.ID))
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)
		require.Equal(t, []intruntime.OperationHealthResult{{Name: renamedSyncOp.Name(), Reason: "boom"}}, assessment.Operations)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, map[string]string{renamedSyncOp.Name(): "boom"}, reloaded.Health.UnhealthyOperations)
		require.Equal(t, renamed.Registry().Version(testint.DefinitionID.ID()), reloaded.DefinitionVersion)

		names, err := suite.Client.DB.IntegrationRun.Query().
			Where(integrationrun.IntegrationIDEQ(installation.ID), integrationrun.OperationNameIn(retiredSyncOp.Name(), renamedSyncOp.Name())).
			Select(integrationrun.FieldOperationName).
			Strings(subCtx)
		require.NoError(t, err)
		require.Equal(t, []string{renamedSyncOp.Name(), renamedSyncOp.Name()}, names)

		movedAt, err := operations.LastSuccessfulRunAt(subCtx, suite.Client.DB, installation.ID, renamedSyncOp.Name())
		require.NoError(t, err)
		require.NotNil(t, movedAt)
		require.True(t, movedAt.Equal(*finishedAt))

		retiredAt, err := operations.LastSuccessfulRunAt(subCtx, suite.Client.DB, installation.ID, retiredSyncOp.Name())
		require.NoError(t, err)
		require.Nil(t, retiredAt)
	})

	t.Run("an operation rename cancels the loop queued under the retired name and reseeds one under the current name", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		previous := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.Operations = []integrationtypes.OperationRegistration{syncOperation(retiredSyncOp, integrationtypes.ExecutionPolicy{Inline: true})}
		}))
		installation := installOn(t, subCtx, previous, nil, nil, testint.TokenCredential.ID(), testint.TokenCredentialSet("token"))

		retiredFragment := reconcileLoopFragment(t, installation.ID, retiredSyncOp.Name())
		currentFragment := reconcileLoopFragment(t, installation.ID, testint.RecurringOp.Name())

		waitForEvents()

		_, err := suite.GalaRuntime.PurgeActiveJobsWithMetadata(subCtx, currentFragment)
		require.NoError(t, err)
		require.Zero(t, activeReconcileJobs(t, currentFragment))

		seedRetiredLoop(t, subCtx, installation, retiredSyncOp.Name())
		require.Equal(t, 1, activeReconcileJobs(t, retiredFragment))

		renamed := queuedRuntimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.Operations = []integrationtypes.OperationRegistration{syncOperation(testint.RecurringOp, integrationtypes.ExecutionPolicy{Reconcile: true}, retiredSyncOp.Name())}
		}))

		assessment, err := renamed.RunHealthAssessment(subCtx, reloadIntegration(t, subCtx, installation.ID))
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)
		require.Equal(t, renamed.Registry().Version(testint.DefinitionID.ID()), reloadIntegration(t, subCtx, installation.ID).DefinitionVersion)

		waitForEvents()

		require.Zero(t, activeReconcileJobs(t, retiredFragment))
		require.Equal(t, 1, activeReconcileJobs(t, currentFragment))
	})

	t.Run("a webhook rename keeps the provider-facing endpoint", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		previous := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.Webhooks = []integrationtypes.WebhookRegistration{eventsWebhook(retiredEventsWebhook, nil, renameEventA)}
		}))
		installation := installOn(t, subCtx, previous, nil, nil, testint.TokenCredential.ID(), testint.TokenCredentialSet("token"))

		rows := endpointRows(t, subCtx, installation.ID)
		require.Len(t, rows, 1)
		require.Equal(t, retiredEventsWebhook.Name(), rows[0].Name)

		endpointID := lo.FromPtr(rows[0].EndpointID)
		secret := rows[0].SecretToken
		require.NotEmpty(t, endpointID)
		require.NotEmpty(t, secret)

		renamed := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.Webhooks = []integrationtypes.WebhookRegistration{eventsWebhook(renamedEventsWebhook, []string{retiredEventsWebhook.Name()}, renameEventA, renameEventB)}
		}))

		early, err := renamed.EnsureWebhook(subCtx, installation, renamedEventsWebhook.Name(), "")
		require.NoError(t, err)
		require.NotEqual(t, endpointID, lo.FromPtr(early.EndpointID))
		require.Len(t, endpointRows(t, subCtx, installation.ID), 2)

		duplicate, err := previous.PrepareWebhookDelivery(subCtx, rows[0], "d1")
		require.NoError(t, err)
		require.False(t, duplicate)

		_, err = renamed.RunHealthAssessment(subCtx, installation)
		require.NoError(t, err)

		rows = endpointRows(t, subCtx, installation.ID)
		require.Len(t, rows, 1)
		require.Equal(t, renamedEventsWebhook.Name(), rows[0].Name)
		require.Equal(t, endpointID, lo.FromPtr(rows[0].EndpointID))
		require.Equal(t, secret, rows[0].SecretToken)
		require.Equal(t, []string{renameEventA.Name(), renameEventB.Name()}, rows[0].AllowedEvents)

		dedupes, err := suite.Client.DB.IntegrationWebhook.Query().
			Where(integrationwebhook.IntegrationIDEQ(installation.ID), integrationwebhook.ExternalEventIDEQ("d1")).
			All(subCtx)
		require.NoError(t, err)
		require.Len(t, dedupes, 1)
		require.Equal(t, renamedEventsWebhook.Name(), dedupes[0].Name)

		fresh := testint.TokenCredentialSet("fresh")
		require.NoError(t, renamed.Reconcile(subCtx, reloadIntegration(t, subCtx, installation.ID), nil, nil, testint.TokenCredential.ID(), &fresh, nil))

		rows = endpointRows(t, subCtx, installation.ID)
		require.Len(t, rows, 1)
		require.Equal(t, endpointID, lo.FromPtr(rows[0].EndpointID))

		untouched := installOn(t, subCtx, previous, nil, nil, testint.TokenCredential.ID(), testint.TokenCredentialSet("token"))
		untouchedRows := endpointRows(t, subCtx, untouched.ID)
		require.Len(t, untouchedRows, 1)
		untouchedEndpoint := lo.FromPtr(untouchedRows[0].EndpointID)

		_, err = renamed.EnsureWebhook(subCtx, untouched, renamedEventsWebhook.Name(), "")
		require.NoError(t, err)

		require.NoError(t, renamed.Reconcile(subCtx, untouched, nil, nil, testint.TokenCredential.ID(), &fresh, nil))

		untouchedRows = endpointRows(t, subCtx, untouched.ID)
		require.Len(t, untouchedRows, 1)
		require.Equal(t, renamedEventsWebhook.Name(), untouchedRows[0].Name)
		require.Equal(t, untouchedEndpoint, lo.FromPtr(untouchedRows[0].EndpointID))
		require.Equal(t, renamed.Registry().Version(testint.DefinitionID.ID()), reloadIntegration(t, subCtx, untouched.ID).DefinitionVersion)

		_, err = renamed.Registry().Webhook(testint.DefinitionID.ID(), retiredEventsWebhook.Name())
		require.ErrorIs(t, err, registry.ErrWebhookNotFound)

		renamedDef, ok := renamed.Registry().Definition(testint.DefinitionID.ID())
		require.True(t, ok)

		registration, replaced, found := renamedDef.ResolveWebhook(retiredEventsWebhook.Name())
		require.True(t, found)
		require.True(t, replaced)
		require.Equal(t, renamedEventsWebhook.Name(), registration.Name)
	})

	t.Run("a new connection is stamped with the current version", func(t *testing.T) {
		installation, _ := newHarnessInstallation(t, ctx, testint.ModeRecurring)
		require.Equal(t, current, installation.DefinitionVersion)
	})
}
