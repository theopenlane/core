//go:build test

package engine_test

import (
	"encoding/json"
	"strings"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/workflows"
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
func (s *WorkflowEngineTestSuite) TestReviewWorkflowTriggersAndCompletes() {
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
	s.Require().NoError(err)

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
	s.Require().NoError(err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetStatus(enums.ControlStatusPreparing).
		Save(seedCtx)
	s.Require().NoError(err)

	updated, err := s.client.Control.UpdateOneID(control.ID).
		SetStatus(enums.ControlStatusApproved).
		Save(seedCtx)
	s.Require().NoError(err)
	s.Equal(enums.ControlStatusApproved, updated.Status)

	s.WaitForEvents()

	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	s.Require().NoError(err)
	s.Empty(instance.WorkflowProposalID)
	s.Equal(enums.WorkflowActionTypeReview.String(), instance.DefinitionSnapshot.Actions[0].Type)

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	s.Require().NoError(err)
	s.Len(assignments, 1)
	s.True(strings.HasPrefix(assignments[0].AssignmentKey, "review_"+action.Key+"_"))
	s.Equal(enums.WorkflowAssignmentStatusPending, assignments[0].Status)

	wfEngine := s.Engine()
	err = wfEngine.CompleteAssignment(reviewerCtx, assignments[0].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	s.Require().NoError(err)

	s.WaitForEvents()

	completedInstance, err := s.client.WorkflowInstance.Get(seedCtx, instance.ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowInstanceStateCompleted, completedInstance.State)
}

// TestReviewWorkflowTriggersOnCreate verifies CREATE triggers fire and produce review assignments.
func (s *WorkflowEngineTestSuite) TestReviewWorkflowTriggersOnCreate() {
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
	s.Require().NoError(err)

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
	s.Require().NoError(err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetStatus(enums.ControlStatusPreparing).
		Save(seedCtx)
	s.Require().NoError(err)

	s.WaitForEvents()

	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	s.Require().NoError(err)
	s.Empty(instance.WorkflowProposalID)
	s.Equal(enums.WorkflowActionTypeReview.String(), instance.DefinitionSnapshot.Actions[0].Type)

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	s.Require().NoError(err)
	s.Len(assignments, 1)
	s.True(strings.HasPrefix(assignments[0].AssignmentKey, "review_"+action.Key+"_"))
	s.Equal(enums.WorkflowAssignmentStatusPending, assignments[0].Status)
}
