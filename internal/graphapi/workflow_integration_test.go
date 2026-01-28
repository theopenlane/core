package graphapi_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowobjectref"
	"github.com/theopenlane/core/internal/graphapi"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/utils/ulids"
)

func TestWorkflowIntegrationApproval(t *testing.T) {
	// Create dedicated test users for workflow testing
	initiator := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	approver := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)

	// Add approver to initiator's organization
	suite.addUserToOrganization(initiator.UserCtx, t, &approver, enums.RoleAdmin, initiator.OrganizationID)

	// Use initiator's context for creating workflow definitions (has user + org)
	// Then use setContext for bypassing privacy on engine operations
	ctx := setContext(initiator.UserCtx, suite.client.db)

	// Create workflow engine with real eventer for event-driven processing
	wfSetup, err := graphapi.SetupWorkflowEngine(suite.client.db)
	assert.NilError(t, err)
	workflowEngine := wfSetup.Engine

	// Create workflow definition with approval action targeting the approver user
	targets := []workflows.TargetConfig{
		{
			Type: enums.WorkflowTargetTypeUser,
			ID:   approver.ID,
		},
	}

	params := struct {
		Targets  []workflows.TargetConfig `json:"targets"`
		Required bool                     `json:"required"`
		Label    string                   `json:"label"`
	}{
		Targets:  targets,
		Required: true,
		Label:    "Control Status Change Approval",
	}

	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	workflowDef, err := suite.client.db.WorkflowDefinition.Create().
		SetName("Control Approval Workflow").
		SetSchemaType("Control").
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetActive(true).
		SetOwnerID(initiator.OrganizationID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{"status"}},
			},
			Conditions: []models.WorkflowCondition{
				{Expression: "true"}, // Simplified condition for integration test
			},
			Actions: []models.WorkflowAction{
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "control_approval",
					Params: paramsBytes,
				},
			},
		}).
		Save(ctx)
	assert.NilError(t, err)

	t.Run("complete approval workflow", func(t *testing.T) {
		// Create a control in NOT_IMPLEMENTED status
		control, err := suite.client.db.Control.Create().
			SetRefCode("CTL-" + ulids.New().String()).
			SetTitle("Test Control for Approval").
			SetStatus(enums.ControlStatusNotImplemented).
			SetOwnerID(initiator.OrganizationID).
			Save(ctx)
		assert.NilError(t, err)

		// Trigger the workflow directly
		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
		}

		instance, err := workflowEngine.TriggerWorkflow(ctx, workflowDef, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"status"},
		})
		assert.NilError(t, err)
		assert.Check(t, instance != nil)

		// Verify workflow instance was created
		assert.Equal(t, enums.WorkflowInstanceStateRunning, instance.State)
		assert.Equal(t, workflowDef.ID, instance.WorkflowDefinitionID)

		// Wait for assignments to be created via event-driven processing
		assignments, err := graphapi.WaitForAssignments(ctx, suite.client.db, instance.ID, 1)
		assert.NilError(t, err)
		assert.Check(t, len(assignments) >= 1, "expected at least one assignment")

		assignment := assignments[0]
		assert.Equal(t, enums.WorkflowAssignmentStatusPending, assignment.Status)
		assert.Check(t, assignment.AssignmentKey != "")

		// Verify WorkflowObjectRef was created
		objectRef, err := suite.client.db.WorkflowObjectRef.Query().
			Where(workflowobjectref.WorkflowInstanceIDEQ(instance.ID)).
			Only(ctx)
		assert.NilError(t, err)
		assert.Equal(t, control.ID, objectRef.ControlID)

		// Approver approves the assignment
		err = workflowEngine.CompleteAssignment(ctx, assignment.ID, enums.WorkflowAssignmentStatusApproved, &models.WorkflowAssignmentApproval{
			ApprovedAt: time.Now().Format(time.RFC3339),
			Label:      "Looks good",
		}, nil)
		assert.NilError(t, err)

		// Wait for workflow instance state to complete via event-driven processing
		instance, err = graphapi.WaitForInstanceState(ctx, suite.client.db, instance.ID, enums.WorkflowInstanceStateCompleted)
		assert.NilError(t, err)
		assert.Equal(t, enums.WorkflowInstanceStateCompleted, instance.State)

		// Verify assignment was updated
		assignment, err = suite.client.db.WorkflowAssignment.Get(ctx, assignment.ID)
		assert.NilError(t, err)
		assert.Equal(t, enums.WorkflowAssignmentStatusApproved, assignment.Status)
		assert.Assert(t, assignment.ApprovalMetadata.ApprovedAt != "")
	})

	t.Run("rejection workflow", func(t *testing.T) {
		// Create another control
		control, err := suite.client.db.Control.Create().
			SetRefCode("CTL-" + ulids.New().String()).
			SetTitle("Test Control for Rejection").
			SetStatus(enums.ControlStatusNotImplemented).
			SetOwnerID(initiator.OrganizationID).
			Save(ctx)
		assert.NilError(t, err)

		// Trigger the workflow
		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
		}

		instance, err := workflowEngine.TriggerWorkflow(ctx, workflowDef, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"status"},
		})
		assert.NilError(t, err)
		assert.Check(t, instance != nil)

		// Wait for assignments to be created via event-driven processing
		assignments, err := graphapi.WaitForAssignments(ctx, suite.client.db, instance.ID, 1)
		assert.NilError(t, err)
		assert.Check(t, len(assignments) >= 1)

		assignment := assignments[0]

		// Approver rejects the assignment
		err = workflowEngine.CompleteAssignment(ctx, assignment.ID, enums.WorkflowAssignmentStatusRejected, nil, &models.WorkflowAssignmentRejection{
			RejectionReason: "Control needs more details",
		})
		assert.NilError(t, err)

		// Wait for workflow instance state to fail via event-driven processing
		instance, err = graphapi.WaitForInstanceState(ctx, suite.client.db, instance.ID, enums.WorkflowInstanceStateFailed)
		assert.NilError(t, err)
		assert.Equal(t, enums.WorkflowInstanceStateFailed, instance.State)

		// Verify assignment was updated
		assignment, err = suite.client.db.WorkflowAssignment.Get(ctx, assignment.ID)
		assert.NilError(t, err)
		assert.Equal(t, enums.WorkflowAssignmentStatusRejected, assignment.Status)
		assert.Equal(t, "Control needs more details", assignment.RejectionMetadata.RejectionReason)
	})
}

func TestWorkflowIntegrationMultipleApprovers(t *testing.T) {
	// Create dedicated test users
	initiator := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	approver1 := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	approver2 := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)

	// Add approvers to initiator's organization
	suite.addUserToOrganization(initiator.UserCtx, t, &approver1, enums.RoleAdmin, initiator.OrganizationID)
	suite.addUserToOrganization(initiator.UserCtx, t, &approver2, enums.RoleAdmin, initiator.OrganizationID)

	// Use initiator's context for creating workflow definitions (has user + org)
	// Then use setContext for bypassing privacy on engine operations
	ctx := setContext(initiator.UserCtx, suite.client.db)

	// Create workflow engine with real eventer for event-driven processing
	wfSetup, err := graphapi.SetupWorkflowEngine(suite.client.db)
	assert.NilError(t, err)
	workflowEngine := wfSetup.Engine

	// Create workflow with two sequential approval actions
	targets1 := []workflows.TargetConfig{
		{
			Type: enums.WorkflowTargetTypeUser,
			ID:   approver1.ID,
		},
	}

	params1 := struct {
		Targets  []workflows.TargetConfig `json:"targets"`
		Required bool                     `json:"required"`
		Label    string                   `json:"label"`
	}{
		Targets:  targets1,
		Required: true,
		Label:    "First Approval",
	}

	params1Bytes, err := json.Marshal(params1)
	assert.NilError(t, err)

	targets2 := []workflows.TargetConfig{
		{
			Type: enums.WorkflowTargetTypeUser,
			ID:   approver2.ID,
		},
	}

	params2 := struct {
		Targets  []workflows.TargetConfig `json:"targets"`
		Required bool                     `json:"required"`
		Label    string                   `json:"label"`
	}{
		Targets:  targets2,
		Required: true,
		Label:    "Second Approval",
	}

	params2Bytes, err := json.Marshal(params2)
	assert.NilError(t, err)

	workflowDef, err := suite.client.db.WorkflowDefinition.Create().
		SetName("Multi-Approval Workflow").
		SetSchemaType("Control").
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetActive(true).
		SetOwnerID(initiator.OrganizationID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{"status"}},
			},
			Conditions: []models.WorkflowCondition{
				{Expression: "true"}, // Simplified condition for integration test
			},
			Actions: []models.WorkflowAction{
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "first_approval",
					Params: params1Bytes,
				},
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "second_approval",
					Params: params2Bytes,
				},
			},
		}).
		Save(ctx)
	assert.NilError(t, err)

	t.Run("sequential approvals", func(t *testing.T) {
		// Create a control
		control, err := suite.client.db.Control.Create().
			SetRefCode("CTL-" + ulids.New().String()).
			SetTitle("Multi-Approval Control").
			SetStatus(enums.ControlStatusNotImplemented).
			SetOwnerID(initiator.OrganizationID).
			Save(ctx)
		assert.NilError(t, err)

		// Trigger the workflow
		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
		}

		instance, err := workflowEngine.TriggerWorkflow(ctx, workflowDef, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"status"},
		})
		assert.NilError(t, err)
		assert.Check(t, instance != nil)

		// Wait for first assignment to be created (only first approval action runs initially)
		assignments, err := graphapi.WaitForAssignments(ctx, suite.client.db, instance.ID, 1)
		assert.NilError(t, err)
		assert.Check(t, len(assignments) >= 1)

		firstAssignment := assignments[0]

		// First approver approves
		err = workflowEngine.CompleteAssignment(ctx, firstAssignment.ID, enums.WorkflowAssignmentStatusApproved, &models.WorkflowAssignmentApproval{
			ApprovedAt:       time.Now().Format(time.RFC3339),
			ApprovedByUserID: approver1.ID,
		}, nil)
		assert.NilError(t, err)

		// Wait for second assignment to be created (workflow resumes after first approval)
		assignments, err = graphapi.WaitForAssignmentsWithTimeout(ctx, suite.client.db, instance.ID, 2, 10*time.Second)
		assert.NilError(t, err)
		assert.Check(t, len(assignments) >= 2, "should have at least 2 assignments after first approval")

		// Find the second assignment (not the first one we already approved)
		var secondAssignment *generated.WorkflowAssignment
		for _, a := range assignments {
			if a.ID != firstAssignment.ID && a.Status == enums.WorkflowAssignmentStatusPending {
				secondAssignment = a
				break
			}
		}
		assert.Check(t, secondAssignment != nil, "should find a second pending assignment")

		// Second approver approves
		err = workflowEngine.CompleteAssignment(ctx, secondAssignment.ID, enums.WorkflowAssignmentStatusApproved, &models.WorkflowAssignmentApproval{
			ApprovedAt:       time.Now().Format(time.RFC3339),
			ApprovedByUserID: approver2.ID,
		}, nil)
		assert.NilError(t, err)

		// Wait for workflow to complete
		instance, err = graphapi.WaitForInstanceState(ctx, suite.client.db, instance.ID, enums.WorkflowInstanceStateCompleted)
		assert.NilError(t, err)
		assert.Equal(t, enums.WorkflowInstanceStateCompleted, instance.State)
	})

	t.Run("first approver rejects", func(t *testing.T) {
		// Create another control
		control, err := suite.client.db.Control.Create().
			SetRefCode("CTL-" + ulids.New().String()).
			SetTitle("Rejection Test Control").
			SetStatus(enums.ControlStatusNotImplemented).
			SetOwnerID(initiator.OrganizationID).
			Save(ctx)
		assert.NilError(t, err)

		// Trigger the workflow
		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
		}

		instance, err := workflowEngine.TriggerWorkflow(ctx, workflowDef, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"status"},
		})
		assert.NilError(t, err)
		assert.Check(t, instance != nil)

		// Wait for first assignment
		assignments, err := graphapi.WaitForAssignments(ctx, suite.client.db, instance.ID, 1)
		assert.NilError(t, err)
		assert.Check(t, len(assignments) >= 1)

		firstAssignment := assignments[0]

		// First approver rejects
		err = workflowEngine.CompleteAssignment(ctx, firstAssignment.ID, enums.WorkflowAssignmentStatusRejected, nil, &models.WorkflowAssignmentRejection{
			RejectionReason: "Needs more work",
		})
		assert.NilError(t, err)

		// Wait for workflow to fail
		instance, err = graphapi.WaitForInstanceState(ctx, suite.client.db, instance.ID, enums.WorkflowInstanceStateFailed)
		assert.NilError(t, err)
		assert.Equal(t, enums.WorkflowInstanceStateFailed, instance.State)
	})
}
