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
type retiredSync struct {
	integrationtypes.OperationSettings
}

// renamedSync is the config of the operation replacing retiredSync
type renamedSync struct {
	integrationtypes.OperationSettings
}

// renameEvent is the payload of the events on the renamed webhook contract
type renameEvent struct{}

// backfilledToken is the token the current user input backfills when empty
const backfilledToken = "x"

// suiteQueueName is the durable gala queue the suite harness runs
const suiteQueueName = "graphapi_integration_test"

var (
	numericTokenRef          = integrationtypes.NewConnection[tokenCred](testint.Token.Connection().Credential.Name)
	straySlotRef             = integrationtypes.NewConnection[straySlotCred]("straySlot")
	partialServiceAccountRef = integrationtypes.NewConnection[serviceAccountCred](testint.ServiceAccount.Connection().Credential.Name)
	unreplacedRef            = integrationtypes.NewConnection[unreplacedCred]("unreplaced")

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

// retiredClient is the client every retired connection provides
type retiredClient struct{}

// testMetadata is the installation metadata layout every retired definition shares with the current test definition
type testMetadata struct{}

// retiredSlot pairs a retired connection name with the connector declaring it
type retiredSlot struct {
	name      string
	connector integrationtypes.Connector
}

// slotOf pairs a typed connection ref with its connector providing a no-op client and verification
func slotOf[T any](ref integrationtypes.ConnectionRef[T]) retiredSlot {
	connector := ref.
		Provides(func(context.Context, integrationtypes.ConnectionRequest[T]) (*retiredClient, error) {
			return &retiredClient{}, nil
		}).
		Verified(func(context.Context, integrationtypes.ConnectionRequest[T], *retiredClient) (testMetadata, error) {
			return testMetadata{}, nil
		})

	return retiredSlot{name: ref.Connection().Credential.Name, connector: connector}
}

// syncOperation returns a no-op operation registration for the ref, carrying whatever the ref replaces
func syncOperation[Config any](op integrationtypes.OperationRef[Config], policy integrationtypes.ExecutionPolicy) integrationtypes.OperationRegistration {
	return op.Policy(policy).HandlesRequest(func(context.Context, integrationtypes.OperationRequest, Config) (json.RawMessage, error) {
		return nil, nil
	}).Registration()
}

// eventsWebhook returns a webhook registration accepting events, carrying whatever the ref replaces
func eventsWebhook(webhook integrationtypes.WebhookRef, events ...integrationtypes.WebhookEventRef[renameEvent]) integrationtypes.WebhookRegistration {
	return webhook.Registration(integrationtypes.WebhookRegistration{
		Event: func(req integrationtypes.WebhookInboundRequest) (integrationtypes.WebhookReceivedEvent, error) {
			return integrationtypes.WebhookReceivedEvent{Name: string(req.Payload), Payload: req.Payload}, nil
		},
		Events: lo.Map(events, func(event integrationtypes.WebhookEventRef[renameEvent], _ int) integrationtypes.WebhookEventRegistration {
			return event.Registration(integrationtypes.WebhookEventRegistration{
				Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error { return nil },
			})
		}),
	})
}

// previousDefinition returns an earlier version of the shared test definition
func previousDefinition(primary retiredSlot, extra ...retiredSlot) registry.Builder {
	return func() (integrationtypes.Definition, error) {
		return integrationtypes.Definition{
			DefinitionSpec: integrationtypes.DefinitionSpec{
				ID:          testint.DefinitionID.ID(),
				DisplayName: "Test Integration",
				Active:      true,
			},
			Installation: integrationtypes.InstallationOf[testMetadata]().Registration(),
			Connections: lo.Map(append([]retiredSlot{primary}, extra...), func(slot retiredSlot, _ int) integrationtypes.Connector {
				return slot.connector
			}),
		}, nil
	}
}

// definitionOver returns the shared test definition with shape applied over the token slot
func definitionOver(shape func(def *integrationtypes.Definition)) registry.Builder {
	return func() (integrationtypes.Definition, error) {
		def, err := previousDefinition(slotOf(testint.Token))()
		if err != nil {
			return integrationtypes.Definition{}, err
		}

		shape(&def)

		return def, nil
	}
}

// runtimeFor returns a runtime on an in-memory gala running one definition version minted when called
func runtimeFor(t *testing.T, builder registry.Builder) *intruntime.Runtime {
	t.Helper()

	return runtimeOn(t, inMemoryGala(t), versionedRegistry(t, builder))
}

// queuedRuntimeFor returns a runtime on the suite's durable gala queue with no workers, running one definition version minted when called
func queuedRuntimeFor(t *testing.T, builder registry.Builder) *intruntime.Runtime {
	t.Helper()

	instance, err := gala.NewGala(context.Background(), gala.Config{DispatchMode: gala.DispatchModeDurable, ConnectionURI: suite.TF.URI, QueueName: suiteQueueName, WorkerCount: 1})
	require.NoError(t, err)
	t.Cleanup(func() { _ = instance.Close() })

	return runtimeOn(t, instance, versionedRegistry(t, builder))
}

// inMemoryGala returns an in-memory gala closed when the test ends
func inMemoryGala(t *testing.T) *gala.Gala {
	t.Helper()

	instance, err := gala.NewGala(context.Background(), gala.Config{DispatchMode: gala.DispatchModeInMemory, WorkerCount: 1})
	require.NoError(t, err)
	t.Cleanup(func() { _ = instance.Close() })

	return instance
}

// versionedRegistry returns a registry running builder under a version minted when called
func versionedRegistry(t *testing.T, builder registry.Builder) *registry.Registry {
	t.Helper()

	reg, err := testint.VersionedRegistry(builder)
	require.NoError(t, err)

	return reg
}

// runtimeOn returns a runtime on instance running the definitions of reg
func runtimeOn(t *testing.T, instance *gala.Gala, reg *registry.Registry) *intruntime.Runtime {
	t.Helper()

	store, err := keystore.NewStore(suite.Client.DB)
	require.NoError(t, err)

	rt, err := intruntime.New(intruntime.Config{DB: suite.Client.DB, Gala: instance, Keystore: store, Registry: reg})
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

// installOn returns the installation of the single definition rt runs
func installOn(t *testing.T, ctx context.Context, rt *intruntime.Runtime, userInput json.RawMessage, operationConfig map[string]json.RawMessage, primary string, credential integrationtypes.CredentialSet) *ent.Integration {
	t.Helper()

	definitions := rt.Registry().Definitions()
	require.Len(t, definitions, 1)

	def := definitions[0]

	ownerID, err := auth.GetOrganizationIDFromContext(ctx)
	require.NoError(t, err)

	installation, _, err := rt.EnsureInstallation(ctx, ownerID, "", def, userInput, operationConfig)
	require.NoError(t, err)

	require.NoError(t, rt.ReconcileCredential(ctx, installation, primary, credential))

	return reloadIntegration(t, ctx, installation.ID)
}

// installUnder returns the installation and version from an earlier, unversioned definition runtime, older than every versioned runtime
func installUnder(t *testing.T, ctx context.Context, builder registry.Builder, primary string, credentials map[string]integrationtypes.CredentialSet) (*ent.Integration, string) {
	t.Helper()

	reg := registry.New()
	require.NoError(t, reg.RegisterAll(builder))

	rt := runtimeOn(t, inMemoryGala(t), reg)

	installation := installOn(t, ctx, rt, nil, nil, primary, credentials[primary])

	store, err := keystore.NewStore(suite.Client.DB)
	require.NoError(t, err)

	for slot, credential := range credentials {
		if slot == primary {
			continue
		}

		require.NoError(t, store.SaveCredential(ctx, reloadIntegration(t, ctx, installation.ID), slot, credential))
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
	ctx := th.SetInternalContext(org.UserCtx, suite.Client.DB)

	store, err := keystore.NewStore(suite.Client.DB)
	require.NoError(t, err)

	def, ok := suite.IntegrationsRT.Registry().Definition(testint.DefinitionID.ID())
	require.True(t, ok)

	current := suite.IntegrationsRT.Registry().Version(def.ID)
	require.NotEmpty(t, current)

	t.Run("a slot retired by the current definition is upgraded inline on first use", func(t *testing.T) {
		installation, previous := installUnder(t, ctx, previousDefinition(slotOf(testint.LegacyToken)), testint.LegacyToken.Connection().Credential.Name, map[string]integrationtypes.CredentialSet{
			testint.LegacyToken.Connection().Credential.Name: {Data: json.RawMessage(`{"accessToken":"legacy-token"}`)},
		})
		require.Less(t, previous, current)
		require.Equal(t, previous, installation.DefinitionVersion)

		assessment, err := suite.IntegrationsRT.RunHealthAssessment(ctx, installation)
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"token":"legacy-token"}`, string(rows[testint.Token.Connection().Credential.Name].Data))
		require.Empty(t, slotRowIDs(t, ctx, installation.ID, testint.LegacyToken.Connection().Credential.Name))

		reloaded := reloadIntegration(t, ctx, installation.ID)
		require.Equal(t, current, reloaded.DefinitionVersion)

		state, err := def.ProviderState(reloaded.ProviderState)
		require.NoError(t, err)
		require.Equal(t, testint.Token.Connection().Credential.Name, state.CredentialRef)

		before := slotRowIDs(t, ctx, installation.ID, testint.Token.Connection().Credential.Name)

		_, err = suite.IntegrationsRT.RunHealthAssessment(ctx, reloaded)
		require.NoError(t, err)
		require.Equal(t, before, slotRowIDs(t, ctx, installation.ID, testint.Token.Connection().Credential.Name))
	})

	t.Run("a stored type the current schema rejects fails the upgrade and marks the installation errored", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetInternalContext(subOrg.UserCtx, suite.Client.DB)

		installation, previous := installUnder(t, subCtx, previousDefinition(slotOf(numericTokenRef)), numericTokenRef.Connection().Credential.Name, map[string]integrationtypes.CredentialSet{
			numericTokenRef.Connection().Credential.Name: {Data: json.RawMessage(`{"token":1}`)},
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
		require.JSONEq(t, `{"token":1}`, string(rows[testint.Token.Connection().Credential.Name].Data))
	})

	t.Run("a backfilled slot is completed from the installation during the upgrade", func(t *testing.T) {
		installation, previous := installUnder(t, ctx, previousDefinition(slotOf(partialServiceAccountRef)), partialServiceAccountRef.Connection().Credential.Name, map[string]integrationtypes.CredentialSet{
			partialServiceAccountRef.Connection().Credential.Name: {Data: json.RawMessage(`{"projectId":"p"}`)},
		})
		require.NotEqual(t, current, previous)

		assessment, err := suite.IntegrationsRT.RunHealthAssessment(ctx, installation)
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"projectId":"p","serviceAccountEmail":"`+installation.ID+`@backfilled.example.com"}`, string(rows[testint.ServiceAccount.Connection().Credential.Name].Data))
		require.Equal(t, current, reloadIntegration(t, ctx, installation.ID).DefinitionVersion)
	})

	t.Run("reconnecting with a fresh credential recovers an installation whose upgrade failed", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetInternalContext(subOrg.UserCtx, suite.Client.DB)

		installation, _ := installUnder(t, subCtx, previousDefinition(slotOf(numericTokenRef)), numericTokenRef.Connection().Credential.Name, map[string]integrationtypes.CredentialSet{
			numericTokenRef.Connection().Credential.Name: {Data: json.RawMessage(`{"token":1}`)},
		})

		_, err := suite.IntegrationsRT.RunHealthAssessment(subCtx, installation)
		require.ErrorIs(t, err, intruntime.ErrInstallationUpgradeFailed)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusErrored, reloaded.Status)
		require.Equal(t, 1, integrationNotificationCount(t, subCtx, installation.OwnerID, integrationReconfigurationRequiredObjectType))

		cred := testint.TokenCredentialSet("fresh")
		require.NoError(t, suite.IntegrationsRT.ReconcileCredential(subCtx, reloaded, testint.Token.Connection().Credential.Name, cred))

		recovered := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusConnected, recovered.Status)
		require.Equal(t, current, recovered.DefinitionVersion)
		require.Equal(t, 1, integrationNotificationCount(t, subCtx, installation.OwnerID, integrationReconnectedObjectType))

		rows, err := store.LoadAllCredentials(subCtx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"token":"fresh"}`, string(rows[testint.Token.Connection().Credential.Name].Data))
	})

	t.Run("disconnect succeeds when the persisted ref is a retired slot nothing replaces", func(t *testing.T) {
		installation, _ := installUnder(t, ctx, previousDefinition(slotOf(unreplacedRef)), unreplacedRef.Connection().Credential.Name, map[string]integrationtypes.CredentialSet{
			unreplacedRef.Connection().Credential.Name: {Data: json.RawMessage(`{"key":"k"}`)},
		})

		_, err := suite.IntegrationsRT.RunHealthAssessment(ctx, installation)
		require.ErrorIs(t, err, intruntime.ErrInstallationUpgradeFailed)

		reloaded := reloadIntegration(t, ctx, installation.ID)
		require.Equal(t, enums.IntegrationStatusErrored, reloaded.Status)

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.JSONEq(t, `{"key":"k"}`, string(rows[unreplacedRef.Connection().Credential.Name].Data))

		_, err = suite.IntegrationsRT.Disconnect(ctx, reloaded)
		require.NoError(t, err)
		require.False(t, integrationVisible(t, ctx, installation.ID))

		remaining, err := suite.Client.DB.Hush.Query().Where(hush.HasIntegrationsWith(integration.IDEQ(installation.ID))).Count(ctx)
		require.NoError(t, err)
		require.Zero(t, remaining)
	})

	t.Run("a crash between the credential move and the ref repoint converges on rerun", func(t *testing.T) {
		installation, _ := installUnder(t, ctx, previousDefinition(slotOf(testint.LegacyToken)), testint.LegacyToken.Connection().Credential.Name, map[string]integrationtypes.CredentialSet{
			testint.LegacyToken.Connection().Credential.Name: {Data: json.RawMessage(`{"accessToken":"legacy-token"}`)},
		})

		stored, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, stored, 1)

		require.NoError(t, store.ReplaceCredentials(ctx, installation, stored, map[string]integrationtypes.CredentialSet{
			testint.Token.Connection().Credential.Name: testint.TokenCredentialSet("legacy-token"),
		}))
		require.Empty(t, slotRowIDs(t, ctx, installation.ID, testint.LegacyToken.Connection().Credential.Name))

		state, err := def.ProviderState(installation.ProviderState)
		require.NoError(t, err)
		require.Equal(t, testint.LegacyToken.Connection().Credential.Name, state.CredentialRef)

		assessment, err := suite.IntegrationsRT.RunHealthAssessment(ctx, installation)
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)

		reloaded := reloadIntegration(t, ctx, installation.ID)
		require.Equal(t, current, reloaded.DefinitionVersion)

		state, err = def.ProviderState(reloaded.ProviderState)
		require.NoError(t, err)
		require.Equal(t, testint.Token.Connection().Credential.Name, state.CredentialRef)

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"token":"legacy-token"}`, string(rows[testint.Token.Connection().Credential.Name].Data))
	})

	t.Run("concurrent upgrades of one installation both succeed and converge", func(t *testing.T) {
		installation, _ := installUnder(t, ctx, previousDefinition(slotOf(testint.LegacyToken)), testint.LegacyToken.Connection().Credential.Name, map[string]integrationtypes.CredentialSet{
			testint.LegacyToken.Connection().Credential.Name: {Data: json.RawMessage(`{"accessToken":"legacy-token"}`)},
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
		require.JSONEq(t, `{"token":"legacy-token"}`, string(rows[testint.Token.Connection().Credential.Name].Data))
		require.Empty(t, slotRowIDs(t, ctx, installation.ID, testint.LegacyToken.Connection().Credential.Name))

		reloaded := reloadIntegration(t, ctx, installation.ID)
		require.Equal(t, current, reloaded.DefinitionVersion)

		state, err := def.ProviderState(reloaded.ProviderState)
		require.NoError(t, err)
		require.Equal(t, testint.Token.Connection().Credential.Name, state.CredentialRef)
	})

	t.Run("a slot the current definition no longer declares is left in place", func(t *testing.T) {
		installation, _ := installUnder(t, ctx, previousDefinition(slotOf(testint.LegacyToken), slotOf(straySlotRef)), testint.LegacyToken.Connection().Credential.Name, map[string]integrationtypes.CredentialSet{
			testint.LegacyToken.Connection().Credential.Name: {Data: json.RawMessage(`{"accessToken":"legacy-token"}`)},
			straySlotRef.Connection().Credential.Name:        {Data: json.RawMessage(`{"value":"kept"}`)},
		})

		stored, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, stored, 2)

		_, err = suite.IntegrationsRT.RunHealthAssessment(ctx, installation)
		require.NoError(t, err)

		rows, err := store.LoadAllCredentials(ctx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.JSONEq(t, `{"value":"kept"}`, string(rows[straySlotRef.Connection().Credential.Name].Data))
		require.JSONEq(t, `{"token":"legacy-token"}`, string(rows[testint.Token.Connection().Credential.Name].Data))
		require.Equal(t, current, reloadIntegration(t, ctx, installation.ID).DefinitionVersion)
	})

	t.Run("user input is renamed and backfilled on upgrade", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetInternalContext(subOrg.UserCtx, suite.Client.DB)

		previous := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.UserInput = zoneInputRef.Registration()
		}))
		installation := installOn(t, subCtx, previous, json.RawMessage(`{"zone":"eu"}`), nil, testint.Token.Connection().Credential.Name, testint.TokenCredentialSet("token"))
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
		subCtx := th.SetInternalContext(subOrg.UserCtx, suite.Client.DB)

		previous := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.Operations = []integrationtypes.OperationRegistration{syncOperation(retiredSyncOp, integrationtypes.ExecutionPolicy{Inline: true})}
		}))
		installation := installOn(t, subCtx, previous, nil, nil, testint.Token.Connection().Credential.Name, testint.TokenCredentialSet("token"))

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
			def.Operations = []integrationtypes.OperationRegistration{syncOperation(renamedSyncOp.Replacing(retiredSyncOp), integrationtypes.ExecutionPolicy{Inline: true})}
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

	t.Run("an operation rename leaves the queue untouched until the next loop reset cancels the retired loop and seeds one under the current name", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetInternalContext(subOrg.UserCtx, suite.Client.DB)

		previous := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.Operations = []integrationtypes.OperationRegistration{syncOperation(retiredSyncOp, integrationtypes.ExecutionPolicy{Inline: true})}
		}))
		installation := installOn(t, subCtx, previous, nil, nil, testint.Token.Connection().Credential.Name, testint.TokenCredentialSet("token"))

		retiredFragment := reconcileLoopFragment(t, installation.ID, retiredSyncOp.Name())
		currentFragment := reconcileLoopFragment(t, installation.ID, testint.RecurringOp.Name())

		waitForEvents()

		_, err := suite.GalaRuntime.PurgeActiveJobsWithMetadata(subCtx, currentFragment)
		require.NoError(t, err)
		require.Zero(t, activeReconcileJobs(t, currentFragment))

		seedRetiredLoop(t, subCtx, installation, retiredSyncOp.Name())
		require.Equal(t, 1, activeReconcileJobs(t, retiredFragment))

		renamed := queuedRuntimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.Operations = []integrationtypes.OperationRegistration{syncOperation(testint.RecurringOp.Replacing(retiredSyncOp), integrationtypes.ExecutionPolicy{Reconcile: true})}
		}))

		assessment, err := renamed.RunHealthAssessment(subCtx, reloadIntegration(t, subCtx, installation.ID))
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)
		require.Equal(t, renamed.Registry().Version(testint.DefinitionID.ID()), reloadIntegration(t, subCtx, installation.ID).DefinitionVersion)

		waitForEvents()

		require.Equal(t, 1, activeReconcileJobs(t, retiredFragment))
		require.Zero(t, activeReconcileJobs(t, currentFragment))

		require.NoError(t, renamed.ResetReconcileLoops(subCtx, reloadIntegration(t, subCtx, installation.ID)))
		require.Zero(t, activeReconcileJobs(t, retiredFragment))
		require.Equal(t, 1, activeReconcileJobs(t, currentFragment))
	})

	t.Run("a webhook rename keeps the provider-facing endpoint", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetInternalContext(subOrg.UserCtx, suite.Client.DB)

		previous := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.Webhooks = []integrationtypes.WebhookRegistration{eventsWebhook(retiredEventsWebhook, renameEventA)}
		}))
		installation := installOn(t, subCtx, previous, nil, nil, testint.Token.Connection().Credential.Name, testint.TokenCredentialSet("token"))

		rows := endpointRows(t, subCtx, installation.ID)
		require.Len(t, rows, 1)
		require.Equal(t, retiredEventsWebhook.Name(), rows[0].Name)

		endpointID := lo.FromPtr(rows[0].EndpointID)
		secret := rows[0].SecretToken
		require.NotEmpty(t, endpointID)
		require.NotEmpty(t, secret)

		renamed := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.Webhooks = []integrationtypes.WebhookRegistration{eventsWebhook(renamedEventsWebhook.Replacing(retiredEventsWebhook), renameEventA, renameEventB)}
		}))

		early, err := renamed.EnsureWebhook(subCtx, installation, renamedEventsWebhook.Name(), "")
		require.NoError(t, err)
		require.Equal(t, endpointID, lo.FromPtr(early.EndpointID))
		require.Len(t, endpointRows(t, subCtx, installation.ID), 1)

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
		require.NoError(t, renamed.ReconcileCredential(subCtx, reloadIntegration(t, subCtx, installation.ID), testint.Token.Connection().Credential.Name, fresh))

		rows = endpointRows(t, subCtx, installation.ID)
		require.Len(t, rows, 1)
		require.Equal(t, endpointID, lo.FromPtr(rows[0].EndpointID))

		untouched := installOn(t, subCtx, previous, nil, nil, testint.Token.Connection().Credential.Name, testint.TokenCredentialSet("token"))
		untouchedRows := endpointRows(t, subCtx, untouched.ID)
		require.Len(t, untouchedRows, 1)
		untouchedEndpoint := lo.FromPtr(untouchedRows[0].EndpointID)

		_, err = renamed.EnsureWebhook(subCtx, untouched, renamedEventsWebhook.Name(), "")
		require.NoError(t, err)

		require.NoError(t, renamed.ReconcileCredential(subCtx, untouched, testint.Token.Connection().Credential.Name, fresh))

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
