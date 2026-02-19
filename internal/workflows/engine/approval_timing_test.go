//go:build test

package engine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/ent/generated/workflowproposal"
	"github.com/theopenlane/core/internal/workflows"
)

func (s *WorkflowEngineTestSuite) createApprovalWorkflowDefinitionWithTiming(ctx context.Context, orgID string, action models.WorkflowAction, timing enums.WorkflowApprovalTiming) *generated.WorkflowDefinition {
	triggerFields := []string{"status"}
	if len(action.Params) > 0 {
		var params workflows.ApprovalActionParams
		err := json.Unmarshal(action.Params, &params)
		s.Require().NoError(err, "failed to parse approval action params")
		fields := eventqueue.NormalizeStrings(params.Fields)
		if len(fields) > 0 {
			triggerFields = fields
		}
	}
	sort.Strings(triggerFields)

	doc := models.WorkflowDefinitionDocument{
		ApprovalSubmissionMode: enums.WorkflowApprovalSubmissionModeAutoSubmit,
		ApprovalTiming:         timing,
		Triggers: []models.WorkflowTrigger{
			{Operation: "UPDATE", Fields: triggerFields},
		},
		Conditions: []models.WorkflowCondition{
			{Expression: "true"},
		},
		Actions: []models.WorkflowAction{action},
	}

	operations, fields := workflows.DeriveTriggerPrefilter(doc)
	fields = triggerFields

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Approval Workflow " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetTriggerOperations(operations).
		SetTriggerFields(fields).
		SetApprovalSubmissionMode(enums.WorkflowApprovalSubmissionModeAutoSubmit).
		SetDefinitionJSON(doc).
		Save(ctx)
	s.Require().NoError(err)

	return def
}

// TestApprovalTimingPreCommitRoutesToProposal verifies that explicit PRE_COMMIT approvals
// still route changes into proposals and create approval assignments.
//
// Workflow Definition (Plain English):
//
//	"Require approval BEFORE Control.status changes are applied (PRE_COMMIT)"
//
// Test Flow:
//  1. Creates a PRE_COMMIT approval workflow for Control.status
//  2. Updates Control.status (change should be intercepted)
//  3. Verifies the Control status remains unchanged
//  4. Confirms a proposal and approval assignment were created
//
// Why This Matters:
//
//	Ensures PRE_COMMIT timing preserves the proposal-first approval gating behavior.
func (s *WorkflowEngineTestSuite) TestApprovalTimingPreCommitRoutesToProposal() {
	approverID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(approverID, orgID)

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approverID},
			},
		},
		Required: boolPtr(true),
		Label:    "Status Approval",
		Fields:   []string{"status"},
	}
	paramsBytes, err := json.Marshal(params)
	s.Require().NoError(err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "status_pre_commit",
		Params: paramsBytes,
	}

	def := s.createApprovalWorkflowDefinitionWithTiming(seedCtx, orgID, action, enums.WorkflowApprovalTimingPreCommit)

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
	s.Equal(enums.ControlStatusPreparing, updated.Status)

	s.WaitForEvents()

	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	s.Require().NoError(err)
	s.NotEmpty(instance.WorkflowProposalID)
	s.Equal(enums.WorkflowActionTypeApproval.String(), instance.DefinitionSnapshot.Actions[0].Type)

	_, err = s.client.WorkflowProposal.Get(seedCtx, instance.WorkflowProposalID)
	s.Require().NoError(err)

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	s.Require().NoError(err)
	s.Len(assignments, 1)
	s.True(strings.HasPrefix(assignments[0].AssignmentKey, "approval_"+action.Key+"_"))
	s.Equal("APPROVER", assignments[0].Role)
}

// TestApprovalTimingPostCommitCreatesReviewAssignments verifies that POST_COMMIT approvals
// do not block changes and are converted to review actions at snapshot time.
//
// Workflow Definition (Plain English):
//
//	"Allow Control.status to change immediately, then request a review (POST_COMMIT)"
//
// Test Flow:
//  1. Creates a POST_COMMIT approval workflow for Control.status
//  2. Updates Control.status (change should apply immediately)
//  3. Verifies no proposal was created
//  4. Confirms a workflow instance exists with REVIEW assignments
//
// Why This Matters:
//
//	Ensures POST_COMMIT timing behaves as a review and does not gate the mutation.
func (s *WorkflowEngineTestSuite) TestApprovalTimingPostCommitCreatesReviewAssignments() {
	approverID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(approverID, orgID)

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approverID},
			},
		},
		Required: boolPtr(true),
		Label:    "Status Review",
		Fields:   []string{"status"},
	}
	paramsBytes, err := json.Marshal(params)
	s.Require().NoError(err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "status_post_commit",
		Params: paramsBytes,
	}

	def := s.createApprovalWorkflowDefinitionWithTiming(seedCtx, orgID, action, enums.WorkflowApprovalTimingPostCommit)
	freshDef, err := s.client.WorkflowDefinition.Get(seedCtx, def.ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowApprovalTimingPostCommit, freshDef.DefinitionJSON.ApprovalTiming)
	s.False(workflows.DefinitionUsesPreCommitApprovals(freshDef.DefinitionJSON))

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

	proposalCount, err := s.client.WorkflowProposal.Query().
		Where(workflowproposal.OwnerIDEQ(orgID)).
		Count(seedCtx)
	s.Require().NoError(err)
	s.Zero(proposalCount)

	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	s.Require().NoError(err)
	s.Empty(instance.WorkflowProposalID)
	s.Equal(enums.WorkflowInstanceStatePaused, instance.State)
	s.Equal(enums.WorkflowActionTypeReview.String(), instance.DefinitionSnapshot.Actions[0].Type)

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	s.Require().NoError(err)
	s.Len(assignments, 1)
	s.True(strings.HasPrefix(assignments[0].AssignmentKey, "review_"+action.Key+"_"))
	s.Equal("REVIEWER", assignments[0].Role)
}

// TestApprovalTimingPostCommitReviewCompletes verifies that POST_COMMIT reviews can be completed
// without blocking the underlying mutation.
func (s *WorkflowEngineTestSuite) TestApprovalTimingPostCommitReviewCompletes() {
	approverID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(approverID, orgID)

	wfEngine := s.Engine()
	s.client.WorkflowEngine = wfEngine

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approverID},
			},
		},
		Required: boolPtr(true),
		Label:    "Status Review",
		Fields:   []string{"status"},
	}
	paramsBytes, err := json.Marshal(params)
	s.Require().NoError(err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "status_post_commit_complete",
		Params: paramsBytes,
	}

	def := s.createApprovalWorkflowDefinitionWithTiming(seedCtx, orgID, action, enums.WorkflowApprovalTimingPostCommit)

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
	s.Equal(enums.WorkflowInstanceStatePaused, instance.State)

	assignmentKey := fmt.Sprintf("review_%s_%s", action.Key, approverID)
	assignment, err := s.client.WorkflowAssignment.Query().
		Where(
			workflowassignment.WorkflowInstanceIDEQ(instance.ID),
			workflowassignment.AssignmentKeyEQ(assignmentKey),
		).
		Only(seedCtx)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowAssignmentStatusPending, assignment.Status)

	err = wfEngine.CompleteAssignment(userCtx, assignment.ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	s.Require().NoError(err)

	s.WaitForEvents()

	completedInstance, err := s.client.WorkflowInstance.Get(seedCtx, instance.ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowInstanceStateCompleted, completedInstance.State)

	completedAssignment, err := s.client.WorkflowAssignment.Get(seedCtx, assignment.ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowAssignmentStatusApproved, completedAssignment.Status)
}

// TestPreCommitAndPostCommitWorkflowsTriggerInSequence verifies that PRE_COMMIT approvals
// gate changes while POST_COMMIT reviews fire only after proposals are applied.
//
// Workflow Definitions (Plain English):
//
//	Definition 1: "Require approval before Control.status changes (PRE_COMMIT)"
//	Definition 2: "Request review after Control.status changes (POST_COMMIT)"
//
// Test Flow:
//  1. Creates both PRE_COMMIT and POST_COMMIT approval workflows for Control.status
//  2. Updates Control.status (should be intercepted by PRE_COMMIT)
//  3. Verifies no POST_COMMIT instance exists yet
//  4. Approves the PRE_COMMIT assignment (proposal applied)
//  5. Verifies POST_COMMIT review instance and review assignment are created
//
// Why This Matters:
//
//	Ensures PRE_COMMIT and POST_COMMIT workflows can coexist without double-triggering.
func (s *WorkflowEngineTestSuite) TestPreCommitAndPostCommitWorkflowsTriggerInSequence() {
	approverID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(approverID, orgID)

	wfEngine := s.Engine()

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approverID},
			},
		},
		Required: boolPtr(true),
		Label:    "Status Approval",
		Fields:   []string{"status"},
	}
	paramsBytes, err := json.Marshal(params)
	s.Require().NoError(err)

	preAction := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "status_pre_commit",
		Params: paramsBytes,
	}
	postAction := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "status_post_commit",
		Params: paramsBytes,
	}

	preDef := s.createApprovalWorkflowDefinitionWithTiming(seedCtx, orgID, preAction, enums.WorkflowApprovalTimingPreCommit)
	postDef := s.createApprovalWorkflowDefinitionWithTiming(seedCtx, orgID, postAction, enums.WorkflowApprovalTimingPostCommit)

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
	s.Equal(enums.ControlStatusPreparing, updated.Status)

	s.WaitForEvents()

	preInstance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(preDef.ID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	s.Require().NoError(err)
	s.NotEmpty(preInstance.WorkflowProposalID)

	postCount, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(postDef.ID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Count(seedCtx)
	s.Require().NoError(err)
	s.Zero(postCount)

	preAssignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(preInstance.ID)).
		All(userCtx)
	s.Require().NoError(err)
	s.Len(preAssignments, 1)
	s.True(strings.HasPrefix(preAssignments[0].AssignmentKey, "approval_"+preAction.Key+"_"))

	err = wfEngine.CompleteAssignment(userCtx, preAssignments[0].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	s.Require().NoError(err)

	s.WaitForEvents()

	updatedControl, err := s.client.Control.Get(userCtx, control.ID)
	s.Require().NoError(err)
	s.Equal(enums.ControlStatusApproved, updatedControl.Status)

	preCount, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(preDef.ID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Count(seedCtx)
	s.Require().NoError(err)
	s.Equal(1, preCount)

	postInstance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(postDef.ID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowActionTypeReview.String(), postInstance.DefinitionSnapshot.Actions[0].Type)

	postAssignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(postInstance.ID)).
		All(userCtx)
	s.Require().NoError(err)
	s.Len(postAssignments, 1)
	s.True(strings.HasPrefix(postAssignments[0].AssignmentKey, "review_"+postAction.Key+"_"))
}
