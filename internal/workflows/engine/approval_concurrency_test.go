//go:build test

package engine_test

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// TestApprovalFlowConcurrentApprovalsResumesOnce verifies that when multiple approvers submit
// their approvals simultaneously (race condition), the workflow instance is resumed exactly once,
// not multiple times.
//
// Workflow Definition (Plain English):
//
//	"Require 2 approvals before changing Control.status"
//
// Test Flow:
//  1. Triggers an approval workflow requiring 2 approvers
//  2. Both approvers submit their approvals CONCURRENTLY (using goroutines)
//  3. Waits for both goroutines to complete
//  4. Verifies the proposal was applied exactly once
//  5. Verifies the Control.status was updated
//  6. Verifies the workflow completed
//  7. Counts INSTANCE_COMPLETED events - must be exactly 1 (not 2)
//
// Why This Matters:
//
//	In a distributed system, concurrent approvals could race to resume the workflow. The
//	engine must ensure idempotent completion handling to prevent duplicate side effects
//	(e.g., double-applying the proposal, sending duplicate notifications).
func TestApprovalFlowConcurrentApprovalsResumesOnce(t *testing.T) {
	s := suite.setupTest(t)
	approver1ID, orgID, userCtx := s.SetupTestUser()
	approver2ID, approver2Ctx := s.CreateTestUserInOrg(orgID, enums.RoleMember)

	wfEngine := s.Engine()

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approver1ID},
				{Type: enums.WorkflowTargetTypeUser, ID: approver2ID},
			},
		},
		Required:      boolPtr(true),
		RequiredCount: 2,
		Label:         "Concurrent Approval",
		Fields:        []string{"status"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "status_approval",
		Params: paramsBytes,
	}

	def := s.CreateApprovalWorkflowDefinition(userCtx, orgID, action)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-CONCURRENT-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetStatus(enums.ControlStatusNotImplemented).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}

	proposedChanges := map[string]any{"status": enums.ControlStatusApproved}

	instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
		EventType:       "UPDATE",
		ChangedFields:   []string{"status"},
		ProposedChanges: proposedChanges,
	})
	assert.NilError(t, err)
	assert.Assert(t, instance != nil)

	// wait until handlers have drained before asserting
	s.WaitForEvents()

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(assignments, 2))

	assignmentsByUser := map[string]*generated.WorkflowAssignment{}
	for _, assignment := range assignments {
		switch {
		case strings.HasSuffix(assignment.AssignmentKey, approver1ID):
			assignmentsByUser[approver1ID] = assignment
		case strings.HasSuffix(assignment.AssignmentKey, approver2ID):
			assignmentsByUser[approver2ID] = assignment
		}
	}
	assert.Assert(t, assignmentsByUser[approver1ID] != nil)
	assert.Assert(t, assignmentsByUser[approver2ID] != nil)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(2)
	go func() {
		defer wg.Done()
		errCh <- wfEngine.CompleteAssignment(userCtx, assignmentsByUser[approver1ID].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	}()
	go func() {
		defer wg.Done()
		errCh <- wfEngine.CompleteAssignment(approver2Ctx, assignmentsByUser[approver2ID].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	}()
	wg.Wait()
	close(errCh)

	for err := range errCh {
		assert.NilError(t, err)
	}

	// Wait for async event processing to complete before checking state
	s.WaitForEvents()

	proposal, err := s.client.WorkflowProposal.Get(userCtx, instance.WorkflowProposalID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowProposalStateApplied, proposal.State))

	updatedControl, err := s.client.Control.Get(userCtx, control.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.ControlStatusApproved, updatedControl.Status))

	updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStateCompleted, updatedInstance.State))

	events, err := s.client.WorkflowEvent.Query().
		Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	assert.NilError(t, err)

	completedCount := 0
	for _, event := range events {
		switch event.EventType {
		case enums.WorkflowEventTypeInstanceCompleted:
			completedCount++
		}
	}

	assert.Check(t, is.Equal(1, completedCount))
}

// TestApprovalFlowLateApprovalDoesNotReapply verifies that "late" approvals (approvals submitted
// after the quorum was already met and the workflow completed) do not re-apply changes or
// generate duplicate completion events.
//
// Workflow Definition (Plain English):
//
//	"Require 1 approval (out of 2 possible approvers) before changing Control.status"
//	RequiredCount = 1, but 2 approvers are assigned
//
// Test Flow:
//  1. Triggers an approval workflow with 2 approvers but only 1 required
//  2. First approver approves - workflow completes (quorum met)
//  3. Records the number of INSTANCE_COMPLETED events
//  4. Second approver approves (late - after completion)
//  5. Verifies NO new INSTANCE_COMPLETED events were generated
//  6. Verifies the Control.status is still the correctly applied value
//
// Why This Matters:
//
//	Users may approve workflows that have already completed. The engine must handle these
//	"late" approvals gracefully without re-applying changes or corrupting workflow state.
func TestApprovalFlowLateApprovalDoesNotReapply(t *testing.T) {
	s := suite.setupTest(t)
	approver1ID, orgID, userCtx := s.SetupTestUser()
	approver2ID, approver2Ctx := s.CreateTestUserInOrg(orgID, enums.RoleMember)

	wfEngine := s.Engine()

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approver1ID},
				{Type: enums.WorkflowTargetTypeUser, ID: approver2ID},
			},
		},
		Required:      boolPtr(true),
		RequiredCount: 1,
		Label:         "Single Approval",
		Fields:        []string{"status"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "status_approval",
		Params: paramsBytes,
	}

	def := s.CreateApprovalWorkflowDefinition(userCtx, orgID, action)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-LATE-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetStatus(enums.ControlStatusNotImplemented).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}

	proposedChanges := map[string]any{"status": enums.ControlStatusApproved}

	instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
		EventType:       "UPDATE",
		ChangedFields:   []string{"status"},
		ProposedChanges: proposedChanges,
	})
	assert.NilError(t, err)
	assert.Assert(t, instance != nil)

	s.WaitForEvents()

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(assignments, 2))

	assignmentsByUser := map[string]*generated.WorkflowAssignment{}
	for _, assignment := range assignments {
		switch {
		case strings.HasSuffix(assignment.AssignmentKey, approver1ID):
			assignmentsByUser[approver1ID] = assignment
		case strings.HasSuffix(assignment.AssignmentKey, approver2ID):
			assignmentsByUser[approver2ID] = assignment
		}
	}
	assert.Assert(t, assignmentsByUser[approver1ID] != nil)
	assert.Assert(t, assignmentsByUser[approver2ID] != nil)

	err = wfEngine.CompleteAssignment(userCtx, assignmentsByUser[approver1ID].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	assert.NilError(t, err)

	s.WaitForEvents()

	eventsBefore, err := s.client.WorkflowEvent.Query().
		Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	assert.NilError(t, err)

	completedBefore := 0
	for _, event := range eventsBefore {
		switch event.EventType {
		case enums.WorkflowEventTypeInstanceCompleted:
			completedBefore++
		}
	}

	err = wfEngine.CompleteAssignment(approver2Ctx, assignmentsByUser[approver2ID].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	assert.NilError(t, err)

	s.WaitForEvents()

	eventsAfter, err := s.client.WorkflowEvent.Query().
		Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	assert.NilError(t, err)

	completedAfter := 0
	for _, event := range eventsAfter {
		switch event.EventType {
		case enums.WorkflowEventTypeInstanceCompleted:
			completedAfter++
		}
	}

	assert.Check(t, is.DeepEqual(completedBefore, completedAfter))

	updatedControl, err := s.client.Control.Get(userCtx, control.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.ControlStatusApproved, updatedControl.Status))

	updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStateCompleted, updatedInstance.State))
}

// TestConcurrentApprovalActionsGateUntilAllSatisfied ensures that when a workflow has multiple
// INDEPENDENT approval actions (different action keys), ALL must be satisfied before the
// workflow completes. Completing one action does not resume execution until all are done.
//
// Workflow Definition (Plain English):
//
//	Actions:
//	  1. "Approval A" - requires User1 to approve "status" field changes
//	  2. "Approval B" - requires User2 to approve "reference_id" field changes
//	Both actions have when="true" (always execute)
//
// Test Flow:
//  1. Triggers a workflow with two independent approval actions
//  2. Verifies 2 assignments are created (one per action)
//  3. User2 approves "Approval B" first
//  4. Verifies the workflow is STILL PAUSED (Approval A pending)
//  5. Verifies the proposal is still in DRAFT state
//  6. User1 approves "Approval A"
//  7. Verifies the workflow COMPLETED (all approvals satisfied)
//  8. Verifies the proposal was APPLIED
//
// Why This Matters:
//
//	Complex workflows may require multiple independent sign-offs. The engine must gate
//	completion until ALL required approval actions are resolved, not just the first one.
func TestConcurrentApprovalActionsGateUntilAllSatisfied(t *testing.T) {
	s := suite.setupTest(t)
	approver1ID, orgID, userCtx := s.SetupTestUser()
	approver2ID, approver2Ctx := s.CreateTestUserInOrg(orgID, enums.RoleMember)

	wfEngine := s.Engine()

	paramsA := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approver1ID},
			},
		},
		Required:      boolPtr(true),
		RequiredCount: 1,
		Label:         "Approval A",
		Fields:        []string{"status"},
	}
	paramsB := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approver2ID},
			},
		},
		Required:      boolPtr(true),
		RequiredCount: 1,
		Label:         "Approval B",
		Fields:        []string{"reference_id"},
	}
	paramsABytes, err := json.Marshal(paramsA)
	assert.NilError(t, err)
	paramsBBytes, err := json.Marshal(paramsB)
	assert.NilError(t, err)

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Concurrent Approval Actions " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetOwnerID(orgID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{"status", "reference_id"}},
			},
			Actions: []models.WorkflowAction{
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "approval_a",
					Params: paramsABytes,
					When:   "true",
				},
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "approval_b",
					Params: paramsBBytes,
					When:   "true",
				},
			},
		}).
		Save(userCtx)
	assert.NilError(t, err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-CONCURRENT-ACTIONS-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetStatus(enums.ControlStatusNotImplemented).
		SetReferenceID("REF-OLD").
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}

	instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
		EventType:       "UPDATE",
		ChangedFields:   []string{"status", "reference_id"},
		ProposedChanges: map[string]any{"status": enums.ControlStatusApproved, "reference_id": "REF-NEW"},
	})
	assert.NilError(t, err)
	assert.Assert(t, instance != nil)

	s.WaitForEvents()

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(assignments, 2))

	var assignmentA, assignmentB *generated.WorkflowAssignment
	for _, assignment := range assignments {
		switch assignment.ApprovalMetadata.ActionKey {
		case "approval_a":
			assignmentA = assignment
		case "approval_b":
			assignmentB = assignment
		}
	}
	assert.Assert(t, assignmentA != nil)
	assert.Assert(t, assignmentB != nil)

	// Complete approval B first; workflow should remain paused until approval A is resolved.
	err = wfEngine.CompleteAssignment(approver2Ctx, assignmentB.ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	assert.NilError(t, err)

	s.WaitForEvents()

	updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStatePaused, updatedInstance.State))

	assignmentA, err = s.client.WorkflowAssignment.Get(userCtx, assignmentA.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusPending, assignmentA.Status))

	if instance.WorkflowProposalID != "" {
		proposal, err := s.client.WorkflowProposal.Get(userCtx, instance.WorkflowProposalID)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(enums.WorkflowProposalStateDraft, proposal.State))
	}

	err = wfEngine.CompleteAssignment(userCtx, assignmentA.ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	assert.NilError(t, err)

	s.WaitForEvents()

	updatedInstance, err = s.client.WorkflowInstance.Get(userCtx, instance.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStateCompleted, updatedInstance.State))

	if instance.WorkflowProposalID != "" {
		proposal, err := s.client.WorkflowProposal.Get(userCtx, instance.WorkflowProposalID)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(enums.WorkflowProposalStateApplied, proposal.State))
	}
}
