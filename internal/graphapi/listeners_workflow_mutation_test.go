//go:build test

package graphapi_test

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
)

func TestWorkflowAssignmentMutationListener(t *testing.T) {
	initiator := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	approver := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	suite.addUserToOrganization(initiator.UserCtx, t, &approver, enums.RoleAdmin, initiator.OrganizationID)

	ctx := setContext(initiator.UserCtx, suite.client.db)

	workflowEngine, workflowRuntime := suite.acquireWorkflowRuntime(t)

	params, err := json.Marshal(struct {
		Targets  []workflows.TargetConfig `json:"targets"`
		Required bool                     `json:"required"`
		Label    string                   `json:"label"`
	}{
		Targets:  []workflows.TargetConfig{{Type: enums.WorkflowTargetTypeUser, ID: approver.ID}},
		Required: true,
		Label:    "Assignment Listener Approval",
	})
	assert.NilError(t, err)

	workflowDef, err := suite.client.db.WorkflowDefinition.Create().
		SetName("Assignment Listener Workflow").
		SetSchemaType("Control").
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetActive(true).
		SetOwnerID(initiator.OrganizationID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			Triggers:   []models.WorkflowTrigger{{Operation: "UPDATE", Fields: []string{"status"}}},
			Conditions: []models.WorkflowCondition{{Expression: "true"}},
			Actions: []models.WorkflowAction{{
				Type:   enums.WorkflowActionTypeApproval.String(),
				Key:    "assignment_listener_approval",
				Params: params,
			}},
		}).
		Save(ctx)
	assert.NilError(t, err)

	control, err := suite.client.db.Control.Create().
		SetRefCode("CTL-" + ulids.New().String()).
		SetTitle("Assignment Listener Control").
		SetStatus(enums.ControlStatusNotImplemented).
		SetOwnerID(initiator.OrganizationID).
		Save(ctx)
	assert.NilError(t, err)

	instance, err := workflowEngine.TriggerWorkflow(ctx, workflowDef, &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}, engine.TriggerInput{EventType: "UPDATE", ChangedFields: []string{"status"}})
	assert.NilError(t, err)

	waitForGala(t, workflowRuntime)

	assignments, err := graphapi.WaitForAssignments(ctx, suite.client.db, instance.ID, 1)
	assert.NilError(t, err)
	assignment := assignments[0]
	assert.Check(t, is.Equal(enums.WorkflowAssignmentStatusPending, assignment.Status))

	_, err = graphapi.WaitForInstanceState(ctx, suite.client.db, instance.ID, enums.WorkflowInstanceStatePaused)
	assert.NilError(t, err)

	assertStillPending := func(t *testing.T) {
		t.Helper()

		waitForGala(t, workflowRuntime)

		current, err := suite.client.db.WorkflowInstance.Get(ctx, instance.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.WorkflowInstanceStatePaused, current.State))

		reloaded, err := suite.client.db.WorkflowAssignment.Get(ctx, assignment.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.WorkflowAssignmentStatusPending, reloaded.Status))
	}

	t.Run("update without status change is skipped", func(t *testing.T) {
		assert.NilError(t, suite.client.db.WorkflowAssignment.UpdateOneID(assignment.ID).
			SetNotes("still deciding").
			Exec(ctx))

		assertStillPending(t)
	})

	t.Run("status set to pending is skipped", func(t *testing.T) {
		assert.NilError(t, suite.client.db.WorkflowAssignment.UpdateOneID(assignment.ID).
			SetStatus(enums.WorkflowAssignmentStatusPending).
			Exec(ctx))

		assertStillPending(t)
	})

	t.Run("non-pending status completes the assignment and the instance", func(t *testing.T) {
		assert.NilError(t, suite.client.db.WorkflowAssignment.UpdateOneID(assignment.ID).
			SetStatus(enums.WorkflowAssignmentStatusApproved).
			Exec(ctx))

		waitForGala(t, workflowRuntime)

		completed, err := graphapi.WaitForInstanceState(ctx, suite.client.db, instance.ID, enums.WorkflowInstanceStateCompleted)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.WorkflowInstanceStateCompleted, completed.State))

		reloaded, err := suite.client.db.WorkflowAssignment.Get(ctx, assignment.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.WorkflowAssignmentStatusApproved, reloaded.Status))
	})
}
