//go:build test

package hooks_test

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/events"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/gala"
)

// TestWorkflowMutationTriggerParityLegacyVsGala verifies equivalent workflow trigger outcomes
// between legacy Soiree listeners and Gala listeners for workflow object mutations.
func (suite *HookTestSuite) TestWorkflowMutationTriggerParityLegacyVsGala() {
	t := suite.T()

	user := suite.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].OrganizationID
	userCtx := entgen.NewContext(auth.NewTestContextForSystemAdmin(user.ID, orgID), suite.client)
	internalCtx := workflows.AllowContext(userCtx)

	wfEngine, err := engine.NewWorkflowEngine(suite.client, nil)
	suite.NoError(err)

	previousEngine := suite.client.WorkflowEngine
	suite.client.WorkflowEngine = wfEngine
	t.Cleanup(func() {
		suite.client.WorkflowEngine = previousEngine
	})

	definition := suite.client.WorkflowDefinition.Create().
		SetName("Workflow Trigger Parity " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindLifecycle).
		SetSchemaType(entgen.TypeControl).
		SetActive(true).
		SetDraft(false).
		SetTriggerOperations([]string{"UPDATE"}).
		SetTriggerFields([]string{control.FieldStatus}).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			Triggers: []models.WorkflowTrigger{
				{
					Operation: "UPDATE",
					Fields:    []string{control.FieldStatus},
				},
			},
		}).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	legacyControl := suite.client.Control.Create().
		SetRefCode("CTL-LEGACY-" + ulids.New().String()).
		SetOwnerID(orgID).
		SaveX(internalCtx)
	galaControl := suite.client.Control.Create().
		SetRefCode("CTL-GALA-" + ulids.New().String()).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	legacyBus := soiree.New(soiree.Client(suite.client))
	t.Cleanup(func() {
		require.NoError(t, legacyBus.Close())
	})

	legacyEventer := hooks.NewEventer(
		hooks.WithEventerEmitter(legacyBus),
		hooks.WithWorkflowListenersEnabled(false),
	)
	hooks.RegisterWorkflowListeners(legacyEventer)
	suite.NoError(hooks.RegisterListeners(legacyEventer))

	legacyPayload := &events.MutationPayload{
		MutationType:  entgen.TypeControl,
		Operation:     ent.OpUpdateOne.String(),
		EntityID:      legacyControl.ID,
		ChangedFields: []string{control.FieldStatus},
		ProposedChanges: map[string]any{
			control.FieldStatus: enums.ControlStatusApproved,
		},
		Client: suite.client,
	}

	emitMutationEvent(
		t,
		legacyBus,
		entgen.TypeControl,
		legacyPayload,
		userCtx,
		mutationProperties(legacyControl.ID, entgen.TypeControl),
	)

	require.Eventually(t, func() bool {
		count, queryErr := suite.client.WorkflowInstance.Query().
			Where(
				workflowinstance.WorkflowDefinitionIDEQ(definition.ID),
				workflowinstance.ControlIDEQ(legacyControl.ID),
			).
			Count(internalCtx)
		if queryErr != nil {
			return false
		}

		return count == 1
	}, 5*time.Second, 100*time.Millisecond)

	galaRuntime := newWorkflowGalaRuntime(t, suite.client)
	galaPayload := &events.MutationPayload{
		MutationType:  entgen.TypeControl,
		Operation:     ent.OpUpdateOne.String(),
		EntityID:      galaControl.ID,
		ChangedFields: []string{control.FieldStatus},
		ProposedChanges: map[string]any{
			control.FieldStatus: enums.ControlStatusApproved,
		},
		Client: suite.client,
	}

	dispatchGalaMutation(
		t,
		galaRuntime,
		userCtx,
		entgen.TypeControl,
		galaPayload,
	)

	require.Eventually(t, func() bool {
		count, queryErr := suite.client.WorkflowInstance.Query().
			Where(
				workflowinstance.WorkflowDefinitionIDEQ(definition.ID),
				workflowinstance.ControlIDEQ(galaControl.ID),
			).
			Count(internalCtx)
		if queryErr != nil {
			return false
		}

		return count == 1
	}, 5*time.Second, 100*time.Millisecond)
}

// TestWorkflowAssignmentCompletionParityLegacyVsGala verifies equivalent assignment
// completion outcomes between legacy Soiree listeners and Gala listeners.
func (suite *HookTestSuite) TestWorkflowAssignmentCompletionParityLegacyVsGala() {
	t := suite.T()

	user := suite.seedSystemAdmin()
	orgID := user.Edges.OrgMemberships[0].OrganizationID
	userCtx := entgen.NewContext(auth.NewTestContextForSystemAdmin(user.ID, orgID), suite.client)
	internalCtx := workflows.AllowContext(userCtx)

	wfEngine, err := engine.NewWorkflowEngine(suite.client, nil)
	suite.NoError(err)

	previousEngine := suite.client.WorkflowEngine
	suite.client.WorkflowEngine = wfEngine
	t.Cleanup(func() {
		suite.client.WorkflowEngine = previousEngine
	})

	definition := suite.client.WorkflowDefinition.Create().
		SetName("Workflow Assignment Parity " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindLifecycle).
		SetSchemaType(entgen.TypeControl).
		SetActive(true).
		SetDraft(false).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{}).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	legacyControl := suite.client.Control.Create().
		SetRefCode("CTL-ASSIGN-LEGACY-" + ulids.New().String()).
		SetOwnerID(orgID).
		SaveX(internalCtx)
	galaControl := suite.client.Control.Create().
		SetRefCode("CTL-ASSIGN-GALA-" + ulids.New().String()).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	legacyInstance := suite.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID(definition.ID).
		SetState(enums.WorkflowInstanceStatePaused).
		SetControlID(legacyControl.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)
	galaInstance := suite.client.WorkflowInstance.Create().
		SetWorkflowDefinitionID(definition.ID).
		SetState(enums.WorkflowInstanceStatePaused).
		SetControlID(galaControl.ID).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	legacyAssignment := suite.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(legacyInstance.ID).
		SetAssignmentKey("legacy-assignment-" + ulids.New().String()).
		SetStatus(enums.WorkflowAssignmentStatusPending).
		SetOwnerID(orgID).
		SaveX(internalCtx)
	galaAssignment := suite.client.WorkflowAssignment.Create().
		SetWorkflowInstanceID(galaInstance.ID).
		SetAssignmentKey("gala-assignment-" + ulids.New().String()).
		SetStatus(enums.WorkflowAssignmentStatusPending).
		SetOwnerID(orgID).
		SaveX(internalCtx)

	legacyBus := soiree.New(soiree.Client(suite.client))
	t.Cleanup(func() {
		require.NoError(t, legacyBus.Close())
	})

	legacyEventer := hooks.NewEventer(
		hooks.WithEventerEmitter(legacyBus),
		hooks.WithWorkflowListenersEnabled(false),
	)
	hooks.RegisterWorkflowListeners(legacyEventer)
	suite.NoError(hooks.RegisterListeners(legacyEventer))

	legacyPayload := &events.MutationPayload{
		MutationType:  entgen.TypeWorkflowAssignment,
		Operation:     ent.OpUpdateOne.String(),
		EntityID:      legacyAssignment.ID,
		ChangedFields: []string{workflowassignment.FieldStatus},
		ProposedChanges: map[string]any{
			workflowassignment.FieldStatus: enums.WorkflowAssignmentStatusApproved,
		},
		Client: suite.client,
	}

	emitMutationEvent(
		t,
		legacyBus,
		entgen.TypeWorkflowAssignment,
		legacyPayload,
		userCtx,
		mutationProperties(legacyAssignment.ID, entgen.TypeWorkflowAssignment),
	)

	reloadedLegacyAssignment, err := suite.client.WorkflowAssignment.Get(internalCtx, legacyAssignment.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowAssignmentStatusApproved, reloadedLegacyAssignment.Status)

	galaRuntime := newWorkflowGalaRuntime(t, suite.client)
	galaPayload := &events.MutationPayload{
		MutationType:  entgen.TypeWorkflowAssignment,
		Operation:     ent.OpUpdateOne.String(),
		EntityID:      galaAssignment.ID,
		ChangedFields: []string{workflowassignment.FieldStatus},
		ProposedChanges: map[string]any{
			workflowassignment.FieldStatus: enums.WorkflowAssignmentStatusApproved,
		},
		Client: suite.client,
	}

	dispatchGalaMutation(
		t,
		galaRuntime,
		userCtx,
		entgen.TypeWorkflowAssignment,
		galaPayload,
	)

	reloadedGalaAssignment, err := suite.client.WorkflowAssignment.Get(internalCtx, galaAssignment.ID)
	suite.NoError(err)
	suite.Equal(enums.WorkflowAssignmentStatusApproved, reloadedGalaAssignment.Status)
}

// emitMutationEvent emits a mutation event and blocks until all listener outcomes are drained.
func emitMutationEvent(
	t *testing.T,
	bus *soiree.EventBus,
	topic string,
	payload *events.MutationPayload,
	ctx context.Context,
	props soiree.Properties,
) {
	t.Helper()

	event := soiree.NewBaseEvent(topic, payload)
	event.SetContext(ctx)
	event.SetProperties(props)

	for emitErr := range bus.Emit(topic, event) {
		require.NoError(t, emitErr)
	}
}

// newWorkflowGalaRuntime builds a runtime with workflow listeners registered for parity tests.
func newWorkflowGalaRuntime(t *testing.T, client *entgen.Client) *gala.Runtime {
	t.Helper()

	runtime, err := gala.NewRuntime(gala.RuntimeOptions{})
	require.NoError(t, err)

	gala.ProvideValue(runtime.Injector(), client)
	wfEngine, ok := client.WorkflowEngine.(*engine.WorkflowEngine)
	require.True(t, ok)
	require.NotNil(t, wfEngine)
	gala.ProvideValue(runtime.Injector(), wfEngine)

	_, err = hooks.RegisterGalaWorkflowListeners(runtime.Registry())
	require.NoError(t, err)

	return runtime
}

// dispatchGalaMutation builds and dispatches a mutation envelope through the Gala runtime.
func dispatchGalaMutation(
	t *testing.T,
	runtime *gala.Runtime,
	ctx context.Context,
	topic string,
	payload *events.MutationPayload,
) {
	t.Helper()

	metadata := eventqueue.NewMutationGalaMetadata("", payload)
	envelope, err := eventqueue.NewMutationGalaEnvelope(
		ctx,
		runtime,
		gala.Topic[eventqueue.MutationGalaPayload]{
			Name: gala.TopicName(topic),
		},
		payload,
		metadata,
	)
	require.NoError(t, err)
	require.NoError(t, runtime.DispatchEnvelope(ctx, envelope))
}

// mutationProperties returns the shared mutation metadata used by legacy and Gala dispatch.
func mutationProperties(entityID, mutationType string) soiree.Properties {
	props := soiree.NewProperties()
	props.Set("ID", entityID)
	props.Set("mutation_type", mutationType)

	return props
}
