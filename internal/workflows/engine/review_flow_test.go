//go:build test

package engine_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/v2/internal/workflows"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// TestReviewWorkflowTriggersAndCompletes verifies that REVIEW workflows trigger after a mutation
// applies, create review assignments, and complete once the review is approved.
//
// Workflow Definition (Plain English):
//
//	"After Control.status changes, request a review from a designated reviewer"
//
// Test Flow:
//  1. Creates a REVIEW workflow for Control.status updates
//  2. Updates Control.status (change should apply immediately)
//  3. Verifies a workflow instance exists without a proposal and a review assignment is created
//  4. Completes the review assignment
//  5. Confirms the workflow instance completes
//
// Why This Matters:
//
//	Ensures review-only workflows do not gate mutations but still require human review.
func TestReviewWorkflowTriggersAndCompletes(t *testing.T) {
	s := suite.setupTest(t)
	submitterID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(submitterID, orgID)

	reviewerID, reviewerCtx := s.CreateTestUserInOrg(orgID, enums.RoleMember)

	params := workflows.ReviewActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: reviewerID},
			},
		},
		Required:      boolPtr(true),
		RequiredCount: 1,
		Label:         "Status Review",
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeReview.String(),
		Key:    "status_review",
		Params: paramsBytes,
	}

	doc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{Operation: "UPDATE", Fields: []string{"status"}},
		},
		Conditions: []models.WorkflowCondition{
			{Expression: "true"},
		},
		Actions: []models.WorkflowAction{action},
	}

	operations, fields := workflows.DeriveTriggerPrefilter(doc)

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Review Workflow " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetTriggerOperations(operations).
		SetTriggerFields(fields).
		SetApprovalSubmissionMode(enums.WorkflowApprovalSubmissionModeAutoSubmit).
		SetDefinitionJSON(doc).
		Save(seedCtx)
	assert.NilError(t, err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetStatus(enums.ControlStatusPreparing).
		Save(seedCtx)
	assert.NilError(t, err)

	updated, err := s.client.Control.UpdateOneID(control.ID).
		SetStatus(enums.ControlStatusApproved).
		Save(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.ControlStatusApproved, updated.Status))

	s.WaitForEvents()

	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(instance.WorkflowProposalID, 0))
	assert.Check(t, is.Equal(enums.WorkflowActionTypeReview.String(), instance.DefinitionSnapshot.Actions[0].Type))

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(assignments, 1))
	assert.Check(t, strings.HasPrefix(assignments[0].AssignmentKey, "review_"+action.Key+"_"))
	assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusPending, assignments[0].Status))

	wfEngine := s.Engine()
	err = wfEngine.CompleteAssignment(reviewerCtx, assignments[0].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	assert.NilError(t, err)

	s.WaitForEvents()

	completedInstance, err := s.client.WorkflowInstance.Get(seedCtx, instance.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStateCompleted, completedInstance.State))
}

// TestReviewWorkflowTriggersOnCreate verifies CREATE triggers fire and produce review assignments.
func TestReviewWorkflowTriggersOnCreate(t *testing.T) {
	s := suite.setupTest(t)
	submitterID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(submitterID, orgID)

	reviewerID, _ := s.CreateTestUserInOrg(orgID, enums.RoleMember)

	params := workflows.ReviewActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: reviewerID},
			},
		},
		Required:      boolPtr(true),
		RequiredCount: 1,
		Label:         "Create Review",
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeReview.String(),
		Key:    "create_review",
		Params: paramsBytes,
	}

	doc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{Operation: "CREATE"},
		},
		Conditions: []models.WorkflowCondition{
			{Expression: "true"},
		},
		Actions: []models.WorkflowAction{action},
	}

	operations, fields := workflows.DeriveTriggerPrefilter(doc)

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Create Review Workflow " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetTriggerOperations(operations).
		SetTriggerFields(fields).
		SetApprovalSubmissionMode(enums.WorkflowApprovalSubmissionModeAutoSubmit).
		SetDefinitionJSON(doc).
		Save(seedCtx)
	assert.NilError(t, err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetStatus(enums.ControlStatusPreparing).
		Save(seedCtx)
	assert.NilError(t, err)

	s.WaitForEvents()

	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(instance.WorkflowProposalID, 0))
	assert.Check(t, is.Equal(enums.WorkflowActionTypeReview.String(), instance.DefinitionSnapshot.Actions[0].Type))

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(assignments, 1))
	assert.Check(t, strings.HasPrefix(assignments[0].AssignmentKey, "review_"+action.Key+"_"))
	assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusPending, assignments[0].Status))
}
