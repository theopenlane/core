//go:build test

package engine_test

import (
	"testing"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowdefinition"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowobjectref"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowproposal"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

type actionCompletionSummary struct {
	// Skipped reports whether the action was skipped
	Skipped bool `json:"skipped"`
}

// hasSkippedAction reports whether any action completion event was skipped
func hasSkippedAction(events []*generated.WorkflowEvent) bool {
	for _, event := range events {
		if event.EventType != enums.WorkflowEventTypeActionCompleted {
			continue
		}
		var summary actionCompletionSummary
		if err := json.Unmarshal(event.Payload.Details, &summary); err != nil {
			continue
		}
		if summary.Skipped {
			return true
		}
	}
	return false
}

// TestApprovalFlowQuorumAppliesProposal verifies quorum-based approval workflows correctly apply
// proposed changes only after all required approvals are obtained.
//
// Workflow Definition (Plain English):
//
//	"Require exactly 2 approvers before allowing Control.reference_id to be changed"
//
// Test Flow:
//  1. Creates a Control with an initial reference_id
//  2. Attempts to update the reference_id, which triggers the approval workflow
//  3. Verifies the update was intercepted and routed to a proposal (original value unchanged)
//  4. Verifies 2 approval assignments were created (one per approver)
//  5. First approver approves - verifies proposal NOT yet applied (quorum not met)
//  6. Second approver approves - verifies proposal IS applied (quorum met)
//  7. Confirms the Control.reference_id now reflects the proposed value
//  8. Confirms the workflow instance completed successfully
//
// Why This Matters:
//
//	Ensures the quorum mechanism prevents premature application of changes and only
//	applies them when the required number of approvals is reached.
func TestApprovalFlowQuorumAppliesProposal(t *testing.T) {
	s := suite.setupTest(t)

	approver1ID, orgID, userCtx := s.SetupTestUser()
	approver2ID, approver2Ctx := s.CreateTestUserInOrg(orgID, enums.RoleMember)
	seedCtx := s.SeedContext(approver1ID, orgID)

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
		Label:         "Reference ID Approval",
		Fields:        []string{"reference_id"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "reference_id_approval",
		Params: paramsBytes,
	}

	def := s.CreateApprovalWorkflowDefinition(seedCtx, orgID, action)

	oldRef := "REF-OLD-" + ulid.Make().String()
	newRef := "REF-NEW-" + ulid.Make().String()

	control, err := s.client.Control.Create().
		SetRefCode("CTL-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetReferenceID(oldRef).
		Save(seedCtx)
	assert.NilError(t, err)

	// Trigger workflow by updating control reference ID - the hook intercepts and creates proposal
	updated, err := s.client.Control.UpdateOneID(control.ID).
		SetReferenceID(newRef).
		Save(seedCtx)
	assert.NilError(t, err)
	// Reference ID should remain unchanged since it was routed to proposal
	assert.Assert(t, updated != nil)

	// Wait for async event processing
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
	assert.Assert(t, instance != nil)
	assert.Assert(t, len(instance.WorkflowProposalID) > 0)

	proposal, err := s.client.WorkflowProposal.Get(seedCtx, instance.WorkflowProposalID)
	assert.NilError(t, err)

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

	proposal, err = s.client.WorkflowProposal.Get(userCtx, proposal.ID)
	assert.NilError(t, err)
	assert.Check(t, enums.WorkflowProposalStateApplied != proposal.State)

	updatedControl, err := s.client.Control.Get(userCtx, control.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(oldRef, updatedControl.ReferenceID))

	err = wfEngine.CompleteAssignment(approver2Ctx, assignmentsByUser[approver2ID].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	assert.NilError(t, err)

	s.WaitForEvents()

	proposal, err = s.client.WorkflowProposal.Get(userCtx, proposal.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowProposalStateApplied, proposal.State))
	assert.Check(t, len(proposal.ApprovedHash) > 0)

	updatedControl, err = s.client.Control.Get(userCtx, control.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(newRef, updatedControl.ReferenceID))

	updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStateCompleted, updatedInstance.State))
}

// TestProposalApplyDoesNotCreateRecursiveWorkflow verifies that applying an approved proposal
// does not trigger a new approval workflow for the same definition/object.
//
// Workflow Definition (Plain English):
//
//	"Require approval before Control.reference_id changes"
//
// Test Flow:
//  1. Creates an approval workflow for reference_id changes
//  2. Updates a Control.reference_id (proposal created)
//  3. Approves the assignment (proposal applied)
//  4. Verifies the Control.reference_id is updated
//  5. Confirms no additional workflow instances or proposals were created
//
// Why This Matters:
//
//	When applying proposals, the engine uses bypass context to avoid re-entering
//	approval routing. This test guards against recursive proposal creation.
func TestProposalApplyDoesNotCreateRecursiveWorkflow(t *testing.T) {
	s := suite.setupTest(t)
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
		Label:    "Reference ID Approval",
		Fields:   []string{"reference_id"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "reference_id_approval",
		Params: paramsBytes,
	}

	def := s.CreateApprovalWorkflowDefinition(seedCtx, orgID, action)

	oldRef := "REF-RECURSION-OLD-" + ulid.Make().String()
	newRef := "REF-RECURSION-NEW-" + ulid.Make().String()

	control, err := s.client.Control.Create().
		SetRefCode("CTL-RECURSION-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetReferenceID(oldRef).
		Save(seedCtx)
	assert.NilError(t, err)

	updated, err := s.client.Control.UpdateOneID(control.ID).
		SetReferenceID(newRef).
		Save(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(oldRef, updated.ReferenceID))

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
	assert.Assert(t, len(instance.WorkflowProposalID) > 0)

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(assignments, 1))

	err = wfEngine.CompleteAssignment(userCtx, assignments[0].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	assert.NilError(t, err)

	s.WaitForEvents()

	appliedControl, err := s.client.Control.Get(seedCtx, control.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(newRef, appliedControl.ReferenceID))

	proposal, err := s.client.WorkflowProposal.Get(seedCtx, instance.WorkflowProposalID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowProposalStateApplied, proposal.State))

	instanceCount, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Count(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, instanceCount))

	proposalCount, err := s.client.WorkflowProposal.Query().
		Where(
			workflowproposal.HasWorkflowObjectRefWith(workflowobjectref.ControlIDEQ(control.ID)),
			workflowproposal.OwnerIDEQ(orgID),
		).
		Count(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, proposalCount))
}

// TestApprovalFlowOptionalQuorumProceedsEarly verifies that optional (non-required) approval
// workflows apply changes as soon as any single approval is received, without waiting for all assignees.
//
// Workflow Definition (Plain English):
//
//	"Allow Control.reference_id changes after any single approval (2 approvers assigned, but not required)"
//
// Test Flow:
//  1. Creates a Control with an initial reference_id
//  2. Attempts to update the reference_id, which triggers the optional approval workflow
//  3. Verifies 2 approval assignments were created
//  4. First approver approves - verifies proposal IS immediately applied (optional = early proceed)
//  5. Confirms one assignment is APPROVED and one remains PENDING
//  6. Confirms the Control.reference_id now reflects the proposed value
//  7. Confirms the workflow instance completed successfully
//
// Why This Matters:
//
//	Demonstrates that "required=false" workflows provide a "first approval wins" behavior,
//	useful for notification-style approvals where acknowledgment is sufficient.
func TestApprovalFlowOptionalQuorumProceedsEarly(t *testing.T) {
	s := suite.setupTest(t)

	approver1ID, orgID, userCtx := s.SetupTestUser()
	approver2ID, _ := s.CreateTestUserInOrg(orgID, enums.RoleMember)
	seedCtx := s.SeedContext(approver1ID, orgID)

	wfEngine := s.Engine()

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approver1ID},
				{Type: enums.WorkflowTargetTypeUser, ID: approver2ID},
			},
		},
		Required: boolPtr(false),
		Label:    "Optional Reference ID Approval",
		Fields:   []string{"reference_id"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "optional_reference_id_approval",
		Params: paramsBytes,
	}

	def := s.CreateApprovalWorkflowDefinition(seedCtx, orgID, action)

	oldRef := "REF-OPT-OLD-" + ulid.Make().String()
	newRef := "REF-OPT-NEW-" + ulid.Make().String()

	control, err := s.client.Control.Create().
		SetRefCode("CTL-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetReferenceID(oldRef).
		Save(seedCtx)
	assert.NilError(t, err)

	// Trigger workflow by updating control reference ID
	updated, err := s.client.Control.UpdateOneID(control.ID).
		SetReferenceID(newRef).
		Save(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(oldRef, updated.ReferenceID))

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
	assert.Assert(t, len(instance.WorkflowProposalID) > 0)

	_, err = s.client.WorkflowProposal.Get(seedCtx, instance.WorkflowProposalID)
	assert.NilError(t, err)

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

	updatedControl, err := s.client.Control.Get(userCtx, control.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(newRef, updatedControl.ReferenceID))

	updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStateCompleted, updatedInstance.State))

	assignments, err = s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	assert.NilError(t, err)

	statuses := make(map[enums.WorkflowAssignmentStatus]int)
	for _, assignment := range assignments {
		statuses[assignment.Status]++
	}

	assert.Check(t, is.Equal(1, statuses[enums.WorkflowAssignmentStatusApproved]))
	assert.Check(t, is.Equal(1, statuses[enums.WorkflowAssignmentStatusPending]))
}

// TestApprovalStagingCapturesClearedField verifies that clearing (nullifying) a field value
// is correctly captured in the workflow proposal, not just setting a new value.
//
// Workflow Definition (Plain English):
//
//	"Require approval before clearing Control.reference_id (setting it to null/empty)"
//
// Test Flow:
//  1. Creates a Control with reference_id = "REF-123"
//  2. Calls ClearReferenceID() which triggers the approval workflow
//  3. Verifies the update was intercepted (Control still has "REF-123")
//  4. Verifies a proposal was created with changes["reference_id"] = nil
//  5. Confirms the proposal is in DRAFT state awaiting submission
//
// Why This Matters:
//
//	Field clearing is a distinct operation from field setting. This ensures the workflow
//	system correctly handles "clear" mutations and stages them as null values in proposals.
func TestApprovalStagingCapturesClearedField(t *testing.T) {
	s := suite.setupParallelTest(t)

	userID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Required: boolPtr(true),
		Label:    "Reference ID Approval",
		Fields:   []string{"reference_id"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	// Create definition with seedCtx (privacy bypass for setup)
	def, err := s.client.WorkflowDefinition.Create().
		SetName("Reference ID Approval " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetTriggerOperations([]string{"UPDATE"}).
		SetTriggerFields([]string{"reference_id"}).
		SetApprovalSubmissionMode(enums.WorkflowApprovalSubmissionModeManualSubmit).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			ApprovalSubmissionMode: enums.WorkflowApprovalSubmissionModeManualSubmit,
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{"reference_id"}},
			},
			Conditions: []models.WorkflowCondition{
				{Expression: "true"},
			},
			Actions: []models.WorkflowAction{
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "reference_id_approval",
					Params: paramsBytes,
				},
			},
		}).
		Save(seedCtx)
	assert.NilError(t, err)

	// Create control with seedCtx (privacy bypass for setup)
	control, err := s.client.Control.Create().
		SetRefCode("CTL-REF-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetReferenceID("REF-123").
		Save(seedCtx)
	assert.NilError(t, err)

	// Update with seedCtx - the hook should intercept and route to proposal
	// Using seedCtx since this test is testing workflow hook behavior, not privacy rules
	updated, err := s.client.Control.UpdateOneID(control.ID).
		ClearReferenceID().
		Save(seedCtx)
	assert.NilError(t, err)
	// The hook routes to proposal, so the returned entity should still have the original value
	assert.Check(t, is.Equal("REF-123", updated.ReferenceID))

	refreshed, err := s.client.Control.Get(seedCtx, control.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("REF-123", refreshed.ReferenceID))

	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	assert.NilError(t, err)
	assert.Assert(t, len(instance.WorkflowProposalID) > 0)

	proposal, err := s.client.WorkflowProposal.Query().
		Where(workflowproposal.IDEQ(instance.WorkflowProposalID)).
		WithWorkflowObjectRef().
		Only(seedCtx)
	assert.NilError(t, err)
	assert.Assert(t, proposal.Edges.WorkflowObjectRef != nil)
	assert.Check(t, is.DeepEqual(control.ID, proposal.Edges.WorkflowObjectRef.ControlID))
	// Manual submit is not yet supported; proposals are auto-submitted for now.
	assert.Check(t, is.DeepEqual(enums.WorkflowProposalStateSubmitted, proposal.State))

	value, ok := proposal.Changes["reference_id"]
	assert.Assert(t, ok)
	assert.Check(t, is.Nil(value))
}

// TestApprovalTriggerExpressionUsesCurrentObjectState verifies that trigger expressions evaluate
// against the object's current (pre-mutation) state, not the proposed values. This is critical
// for conditional workflows that should only fire when the object is in a specific state.
//
// Workflow Definition (Plain English):
//
//	"Require approval for reference_id changes, but ONLY when the current reference_id equals 'REF-EXPR-OLD-xxx'"
//	Trigger expression: object.reference_id == "REF-EXPR-OLD-xxx"
//
// Test Flow:
//  1. Creates a Control with reference_id = "REF-EXPR-OLD-xxx"
//  2. Attempts to update reference_id to "REF-EXPR-NEW-xxx"
//  3. Verifies the trigger expression evaluated against the OLD value (matches)
//  4. Confirms a workflow instance and proposal were created
//  5. Confirms an approval assignment was created (workflow triggered correctly)
//
// Why This Matters:
//
//	Trigger expressions must evaluate against the current object state to determine IF
//	a workflow should fire. The proposed changes are captured separately in the proposal.
//	This prevents circular logic where the proposed value would affect trigger evaluation.
func TestApprovalTriggerExpressionUsesCurrentObjectState(t *testing.T) {
	s := suite.setupTest(t)

	userID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Required: boolPtr(true),
		Label:    "Reference ID Approval",
		Fields:   []string{"reference_id"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	oldRef := "REF-EXPR-OLD-" + ulid.Make().String()
	newRef := "REF-EXPR-NEW-" + ulid.Make().String()

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Reference ID Approval With Expression " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetTriggerOperations([]string{"UPDATE"}).
		SetTriggerFields([]string{"reference_id"}).
		SetApprovalSubmissionMode(enums.WorkflowApprovalSubmissionModeAutoSubmit).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			ApprovalSubmissionMode: enums.WorkflowApprovalSubmissionModeAutoSubmit,
			Triggers: []models.WorkflowTrigger{
				{
					Operation:  "UPDATE",
					Fields:     []string{"reference_id"},
					Expression: "object.reference_id == \"" + oldRef + "\"",
				},
			},
			Actions: []models.WorkflowAction{
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "reference_id_approval",
					Params: paramsBytes,
				},
			},
		}).
		Save(seedCtx)
	assert.NilError(t, err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-STATUS-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetReferenceID(oldRef).
		Save(seedCtx)
	assert.NilError(t, err)

	// Update reference ID; approval hook intercepts and creates a proposal, so the control remains unchanged.
	updated, err := s.client.Control.UpdateOneID(control.ID).
		SetReferenceID(newRef).
		Save(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(oldRef, updated.ReferenceID))

	// Wait for async event processing to complete
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
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStatePaused, instance.State))
	assert.Assert(t, len(instance.WorkflowProposalID) > 0)

	proposal, err := s.client.WorkflowProposal.Get(seedCtx, instance.WorkflowProposalID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowProposalStateSubmitted, proposal.State))

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(assignments, 1))
}

// TestApprovalNoTargetsAutoApplies verifies that when an approval action resolves to zero
// assignees (e.g., an empty group), the proposal is automatically applied without blocking.
//
// Workflow Definition (Plain English):
//
//	"Require approval from members of 'Empty Approval Group' before changing Control.reference_id"
//	(Group has zero members, so no one can approve)
//
// Test Flow:
//  1. Creates an empty Group (no members)
//  2. Creates an approval workflow targeting that Group
//  3. Creates a Control and attempts to update reference_id
//  4. Verifies the workflow instance completed immediately (no blocking)
//  5. Verifies the proposal was auto-applied (state = APPLIED)
//  6. Confirms zero assignments were created (no approvers)
//  7. Confirms the Control.reference_id reflects the proposed value
//
// Why This Matters:
//
//	Empty target resolution should not block changes indefinitely. When no one CAN approve,
//	the system auto-applies to prevent deadlock situations.
func TestApprovalNoTargetsAutoApplies(t *testing.T) {
	s := suite.setupTest(t)

	userID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	group, err := s.client.Group.Create().
		SetName("Empty Approval Group " + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(seedCtx)
	assert.NilError(t, err)

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeGroup, ID: group.ID},
			},
		},
		Required: boolPtr(true),
		Label:    "Group Approval",
		Fields:   []string{"reference_id"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	def, err := s.client.WorkflowDefinition.Create().
		SetName("No Target Approval " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetTriggerOperations([]string{"UPDATE"}).
		SetTriggerFields([]string{"reference_id"}).
		SetApprovalSubmissionMode(enums.WorkflowApprovalSubmissionModeAutoSubmit).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			ApprovalSubmissionMode: enums.WorkflowApprovalSubmissionModeAutoSubmit,
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{"reference_id"}},
			},
			Actions: []models.WorkflowAction{
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "reference_id_approval",
					Params: paramsBytes,
				},
			},
		}).
		Save(seedCtx)
	assert.NilError(t, err)

	oldRef := "REF-NO-TARGET-OLD-" + ulid.Make().String()
	newRef := "REF-NO-TARGET-NEW-" + ulid.Make().String()

	control, err := s.client.Control.Create().
		SetRefCode("CTL-NOTARGET-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetReferenceID(oldRef).
		Save(seedCtx)
	assert.NilError(t, err)

	updated, err := s.client.Control.UpdateOneID(control.ID).
		SetReferenceID(newRef).
		Save(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(oldRef, updated.ReferenceID))

	// Wait for async event processing to complete
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
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStateCompleted, instance.State))
	assert.Assert(t, len(instance.WorkflowProposalID) > 0)

	proposal, err := s.client.WorkflowProposal.Get(seedCtx, instance.WorkflowProposalID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowProposalStateApplied, proposal.State))

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(assignments, 0))

	appliedControl, err := s.client.Control.Get(seedCtx, control.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(newRef, appliedControl.ReferenceID))
}

// TestApprovalHookAppliesIneligibleFields verifies that when a mutation contains both
// workflow-eligible and ineligible fields, the ineligible fields are applied immediately
// while eligible fields are staged for approval.
//
// Workflow Definition (Plain English):
//
//	"Require approval for Control.reference_id changes"
//	(Only reference_id is eligible for workflow staging)
//
// Test Flow:
//  1. Creates an approval workflow for reference_id changes
//  2. Creates a Control with an initial reference_id
//  3. Updates BOTH reference_id AND description in the same mutation
//  4. Verifies a proposal was created for reference_id
//  5. Confirms description was applied immediately, while reference_id remains unchanged
//
// Why This Matters:
//
//	Workflow staging should not block unrelated edits. Ineligible fields should pass through
//	while eligible fields are staged for approval.
func TestApprovalHookAppliesIneligibleFields(t *testing.T) {
	s := suite.setupTest(t)

	userID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Required: boolPtr(true),
		Label:    "Reference ID Approval",
		Fields:   []string{"reference_id"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "reference_id_approval",
		Params: paramsBytes,
	}
	_ = s.CreateApprovalWorkflowDefinition(seedCtx, orgID, action)

	oldRef := "REF-INELIGIBLE-OLD-" + ulid.Make().String()
	newRef := "REF-INELIGIBLE-NEW-" + ulid.Make().String()

	control, err := s.client.Control.Create().
		SetRefCode("CTL-INELIGIBLE-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetReferenceID(oldRef).
		Save(seedCtx)
	assert.NilError(t, err)

	_, err = s.client.Control.UpdateOneID(control.ID).
		SetReferenceID(newRef).
		SetDescription("not eligible").
		Save(seedCtx)
	assert.NilError(t, err)

	s.WaitForEvents()

	reloaded, err := s.client.Control.Get(seedCtx, control.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(oldRef, reloaded.ReferenceID))
	assert.Check(t, is.Equal("not eligible", reloaded.Description))

	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	assert.NilError(t, err)
	assert.Assert(t, len(instance.WorkflowProposalID) > 0)

	proposal, err := s.client.WorkflowProposal.Get(seedCtx, instance.WorkflowProposalID)
	assert.NilError(t, err)
	value, ok := proposal.Changes["reference_id"]
	assert.Assert(t, ok)
	assert.Check(t, is.DeepEqual(newRef, value))
}

// TestApprovalChangesRequestedCreatesRequesterAssignment verifies that when an approval assignment
// is marked as CHANGES_REQUESTED, a requester assignment is created with the change details.
func TestApprovalChangesRequestedCreatesRequesterAssignment(t *testing.T) {
	s := suite.setupTest(t)
	submitterID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(submitterID, orgID)

	approverID, approverCtx := s.CreateTestUserInOrg(orgID, enums.RoleMember)

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
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "status_approval",
		Params: paramsBytes,
	}

	_ = s.CreateApprovalWorkflowDefinition(seedCtx, orgID, action)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetStatus(enums.ControlStatusPreparing).
		Save(seedCtx)
	assert.NilError(t, err)

	_, err = s.client.Control.UpdateOneID(control.ID).
		SetStatus(enums.ControlStatusApproved).
		Save(seedCtx)
	assert.NilError(t, err)

	s.WaitForEvents()

	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	assert.NilError(t, err)

	assignment, err := s.client.WorkflowAssignment.Query().
		Where(
			workflowassignment.WorkflowInstanceIDEQ(instance.ID),
			workflowassignment.AssignmentKeyHasPrefix("approval_"+action.Key+"_"),
		).
		Only(seedCtx)
	assert.NilError(t, err)

	rejectionMeta := models.WorkflowAssignmentRejection{
		ActionKey:           action.Key,
		RejectionReason:     "needs more detail",
		RejectedAt:          time.Now().Format(time.RFC3339),
		RejectedByUserID:    approverID,
		ChangeRequestInputs: map[string]any{"status": "in_review"},
	}

	err = wfEngine.CompleteAssignment(approverCtx, assignment.ID, enums.WorkflowAssignmentStatusChangesRequested, nil, &rejectionMeta)
	assert.NilError(t, err)

	s.WaitForEvents()

	requesterKey := fmt.Sprintf("change_request_%s_%s", action.Key, submitterID)
	requesterAssignment, err := s.client.WorkflowAssignment.Query().
		Where(
			workflowassignment.WorkflowInstanceIDEQ(instance.ID),
			workflowassignment.AssignmentKeyEQ(requesterKey),
		).
		Only(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("REQUESTER", requesterAssignment.Role))
	assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusPending, requesterAssignment.Status))
	assert.Check(t, !requesterAssignment.Required)

	meta := requesterAssignment.Metadata
	assert.Assert(t, meta != nil)
	assert.Check(t, is.DeepEqual(action.Key, meta["change_request_action_key"]))
	assert.Check(t, is.Equal("needs more detail", meta["change_request_reason"]))
	inputs, ok := meta["change_request_inputs"].(map[string]any)
	assert.Assert(t, ok)
	assert.Check(t, is.Equal("in_review", inputs["status"]))

	exists, err := s.client.WorkflowAssignmentTarget.Query().
		Where(
			workflowassignmenttarget.WorkflowAssignmentIDEQ(requesterAssignment.ID),
			workflowassignmenttarget.TargetUserIDEQ(submitterID),
		).
		Exist(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, exists)
}

// TestApprovalChangesRequestedClosesApprovalsAndResubmitReopens verifies that when changes
// are requested, outstanding approvals are closed and a re-submission re-opens approvals.
func TestApprovalChangesRequestedClosesApprovalsAndResubmitReopens(t *testing.T) {
	s := suite.setupTest(t)

	submitterID, orgID, submitterCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(submitterID, orgID)

	approver1ID, approver1Ctx := s.CreateTestUserInOrg(orgID, enums.RoleMember)
	approver2ID, _ := s.CreateTestUserInOrg(orgID, enums.RoleMember)

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
		Label:         "Status Approval",
		Fields:        []string{"status"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "status_approval_changes",
		Params: paramsBytes,
	}

	_ = s.CreateApprovalWorkflowDefinition(seedCtx, orgID, action)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetStatus(enums.ControlStatusPreparing).
		Save(submitterCtx)
	assert.NilError(t, err)

	_, err = s.client.Control.UpdateOneID(control.ID).
		SetStatus(enums.ControlStatusApproved).
		Save(seedCtx)
	assert.NilError(t, err)

	s.WaitForEvents()

	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	assert.NilError(t, err)

	assignment1Key := fmt.Sprintf("approval_%s_%s", action.Key, approver1ID)
	assignment2Key := fmt.Sprintf("approval_%s_%s", action.Key, approver2ID)

	assignment1, err := s.client.WorkflowAssignment.Query().
		Where(
			workflowassignment.WorkflowInstanceIDEQ(instance.ID),
			workflowassignment.AssignmentKeyEQ(assignment1Key),
		).
		Only(seedCtx)
	assert.NilError(t, err)

	assignment2, err := s.client.WorkflowAssignment.Query().
		Where(
			workflowassignment.WorkflowInstanceIDEQ(instance.ID),
			workflowassignment.AssignmentKeyEQ(assignment2Key),
		).
		Only(seedCtx)
	assert.NilError(t, err)

	rejectionMeta := models.WorkflowAssignmentRejection{
		ActionKey:        action.Key,
		RejectionReason:  "needs updates",
		RejectedAt:       time.Now().Format(time.RFC3339),
		RejectedByUserID: approver1ID,
	}

	err = wfEngine.CompleteAssignment(approver1Ctx, assignment1.ID, enums.WorkflowAssignmentStatusChangesRequested, nil, &rejectionMeta)
	assert.NilError(t, err)

	s.WaitForEvents()

	reloaded1, err := s.client.WorkflowAssignment.Get(seedCtx, assignment1.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusChangesRequested, reloaded1.Status))

	reloaded2, err := s.client.WorkflowAssignment.Get(seedCtx, assignment2.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusRejected, reloaded2.Status))

	reloadedInstance, err := s.client.WorkflowInstance.Get(seedCtx, instance.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStatePaused, reloadedInstance.State))

	requesterKey := fmt.Sprintf("change_request_%s_%s", action.Key, submitterID)
	requesterAssignment, err := s.client.WorkflowAssignment.Query().
		Where(
			workflowassignment.WorkflowInstanceIDEQ(instance.ID),
			workflowassignment.AssignmentKeyEQ(requesterKey),
		).
		Only(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("REQUESTER", requesterAssignment.Role))
	assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusPending, requesterAssignment.Status))

	requesterTargetExists, err := s.client.WorkflowAssignmentTarget.Query().
		Where(
			workflowassignmenttarget.WorkflowAssignmentIDEQ(requesterAssignment.ID),
			workflowassignmenttarget.TargetUserIDEQ(submitterID),
		).
		Exist(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, requesterTargetExists)

	_, err = s.client.Control.UpdateOneID(control.ID).
		SetStatus(enums.ControlStatusArchived).
		Save(seedCtx)
	assert.NilError(t, err)

	s.WaitForEvents()

	assignment1Again, err := s.client.WorkflowAssignment.Get(seedCtx, assignment1.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusPending, assignment1Again.Status))
	assert.Check(t, len(assignment1Again.InvalidationMetadata.Reason) > 0)

	assignment2Again, err := s.client.WorkflowAssignment.Get(seedCtx, assignment2.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusPending, assignment2Again.Status))
	assert.Check(t, len(assignment2Again.InvalidationMetadata.Reason) > 0)
}

// TestApprovalHookCreatesInstancesForAllMatchingDefinitions verifies that when multiple workflow
// definitions match a single mutation, ALL of them create instances (not just the first match).
//
// Workflow Definition (Plain English):
//
//	Definition 1: "Require approval for Control.reference_id changes (approval key: one)"
//	Definition 2: "Require approval for Control.reference_id changes (approval key: two)"
//	Both definitions trigger on the same field change.
//
// Test Flow:
//  1. Creates two separate approval workflow definitions for reference_id changes
//  2. Creates a Control and attempts to update reference_id
//  3. Verifies TWO workflow instances were created (one per definition)
//  4. Confirms both definitions matched the same trigger criteria
//
// Why This Matters:
//
//	Multiple teams or compliance requirements may have overlapping approval workflows.
//	All matching workflows must execute to ensure complete policy enforcement.
func TestApprovalHookCreatesInstancesForAllMatchingDefinitions(t *testing.T) {
	s := suite.setupTest(t)

	userID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Required: boolPtr(true),
		Label:    "Reference ID Approval",
		Fields:   []string{"reference_id"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action1 := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "reference_id_approval_one",
		Params: paramsBytes,
	}
	action2 := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "reference_id_approval_two",
		Params: paramsBytes,
	}

	def1 := s.CreateApprovalWorkflowDefinition(seedCtx, orgID, action1)
	def2 := s.CreateApprovalWorkflowDefinition(seedCtx, orgID, action2)

	oldRef := "REF-MULTI-OLD-" + ulid.Make().String()
	newRef := "REF-MULTI-NEW-" + ulid.Make().String()

	control, err := s.client.Control.Create().
		SetRefCode("CTL-MULTI-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetReferenceID(oldRef).
		Save(seedCtx)
	assert.NilError(t, err)

	updated, err := s.client.Control.UpdateOneID(control.ID).
		SetReferenceID(newRef).
		Save(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(oldRef, updated.ReferenceID))

	// Wait for async event processing to complete
	s.WaitForEvents()

	instances, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.WorkflowDefinitionIDIn(def1.ID, def2.ID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		All(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(instances, 2))
}

// TestApprovalActionWhenUsesProposedChanges verifies that action-level "when" expressions can
// access both the current object state AND the proposed changes for conditional action execution.
//
// Workflow Definition (Plain English):
//
//	"Require approval for reference_id changes, but ONLY when:
//	 - Current reference_id equals 'REF-PROPOSED-OLD-xxx' AND
//	 - Proposed new reference_id equals 'REF-PROPOSED-NEW-xxx'"
//	When expression: object.reference_id == "OLD" && proposed_changes['reference_id'] == "NEW"
//
// Test Flow:
//  1. Creates a workflow with an action-level "when" clause checking both current and proposed values
//  2. Creates a Control with reference_id = "REF-PROPOSED-OLD-xxx"
//  3. Attempts to update reference_id to "REF-PROPOSED-NEW-xxx"
//  4. Verifies the "when" expression evaluated to true (both conditions met)
//  5. Confirms an approval assignment was created (action was not skipped)
//
// Why This Matters:
//
//	Action-level "when" clauses provide fine-grained control over when specific actions execute.
//	Access to proposed_changes allows conditional logic based on the intended change, not just
//	the current state.
func TestApprovalActionWhenUsesProposedChanges(t *testing.T) {
	s := suite.setupTest(t)

	userID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Required: boolPtr(true),
		Label:    "Reference ID Approval",
		Fields:   []string{"reference_id"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	oldRef := "REF-PROPOSED-OLD-" + ulid.Make().String()
	newRef := "REF-PROPOSED-NEW-" + ulid.Make().String()

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Reference ID Approval With Proposed Changes " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetTriggerOperations([]string{"UPDATE"}).
		SetTriggerFields([]string{"reference_id"}).
		SetApprovalSubmissionMode(enums.WorkflowApprovalSubmissionModeAutoSubmit).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			ApprovalSubmissionMode: enums.WorkflowApprovalSubmissionModeAutoSubmit,
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{"reference_id"}},
			},
			Actions: []models.WorkflowAction{
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "reference_id_approval",
					When:   "object.reference_id == \"" + oldRef + "\" && proposed_changes['reference_id'] == \"" + newRef + "\"",
					Params: paramsBytes,
				},
			},
		}).
		Save(seedCtx)
	assert.NilError(t, err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-PROPOSED-WHEN-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetReferenceID(oldRef).
		Save(seedCtx)
	assert.NilError(t, err)

	updated, err := s.client.Control.UpdateOneID(control.ID).
		SetReferenceID(newRef).
		Save(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(oldRef, updated.ReferenceID))

	// Wait for async event processing to complete
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
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStatePaused, instance.State))
	assert.Assert(t, len(instance.WorkflowProposalID) > 0)

	proposal, err := s.client.WorkflowProposal.Get(seedCtx, instance.WorkflowProposalID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowProposalStateSubmitted, proposal.State))

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(assignments, 1))
	assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusPending, assignments[0].Status))
}

// TestApprovalActionWhenSkipsWhenProposedChangesDoNotMatch verifies that actions are skipped
// when their "when" expression evaluates to false based on proposed_changes mismatch.
//
// Workflow Definition (Plain English):
//
//	"Require approval for reference_id changes, but ONLY when proposed value equals 'REF-SKIP-ARCHIVE-xxx'"
//	When expression: proposed_changes['reference_id'] == "REF-SKIP-ARCHIVE-xxx"
//
// Test Flow:
//  1. Creates a workflow with a "when" clause checking for a specific proposed value
//  2. Creates a Control with reference_id = "REF-SKIP-OLD-xxx"
//  3. Attempts to update reference_id to "REF-SKIP-NEW-xxx" (different from the expected archive value)
//  4. Verifies the "when" expression evaluated to false (proposed value doesn't match)
//  5. Confirms the workflow completed immediately with state = COMPLETED
//  6. Confirms the proposal was auto-applied (no approval needed since action was skipped)
//  7. Confirms zero assignments were created
//
// Why This Matters:
//
//	Demonstrates that "when" expressions can selectively skip approval requirements based on
//	the actual proposed values, enabling value-based routing of changes.
func TestApprovalActionWhenSkipsWhenProposedChangesDoNotMatch(t *testing.T) {
	s := suite.setupTest(t)

	userID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Required: boolPtr(true),
		Label:    "Reference ID Approval",
		Fields:   []string{"reference_id"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	oldRef := "REF-SKIP-OLD-" + ulid.Make().String()
	newRef := "REF-SKIP-NEW-" + ulid.Make().String()
	skipRef := "REF-SKIP-ARCHIVE-" + ulid.Make().String()

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Reference ID Approval With Proposed Changes Skip " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetTriggerOperations([]string{"UPDATE"}).
		SetTriggerFields([]string{"reference_id"}).
		SetApprovalSubmissionMode(enums.WorkflowApprovalSubmissionModeAutoSubmit).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			ApprovalSubmissionMode: enums.WorkflowApprovalSubmissionModeAutoSubmit,
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{"reference_id"}},
			},
			Actions: []models.WorkflowAction{
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "reference_id_approval",
					When:   "proposed_changes['reference_id'] == \"" + skipRef + "\"",
					Params: paramsBytes,
				},
			},
		}).
		Save(seedCtx)
	assert.NilError(t, err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-PROPOSED-WHEN-SKIP-" + ulids.New().String()).
		SetOwnerID(orgID).
		SetReferenceID(oldRef).
		Save(seedCtx)
	assert.NilError(t, err)

	updated, err := s.client.Control.UpdateOneID(control.ID).
		SetReferenceID(newRef).
		Save(seedCtx)
	assert.NilError(t, err)
	assert.Assert(t, updated != nil)

	// Wait for async event processing to complete
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
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStateCompleted, instance.State))
	assert.Assert(t, len(instance.WorkflowProposalID) > 0)

	proposal, err := s.client.WorkflowProposal.Get(seedCtx, instance.WorkflowProposalID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowProposalStateApplied, proposal.State))

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(assignments, 0))
}

// TestSkippedApprovalActionAdvancesWorkflow verifies that when an approval action is skipped
// (via when="false"), subsequent actions still execute and the workflow completes normally.
//
// Workflow Definition (Plain English):
//
//	Actions:
//	  1. Approval action with when="false" (always skipped)
//	  2. Notification action (should still execute)
//
// Test Flow:
//  1. Creates a workflow with an approval action that is always skipped (when="false")
//  2. Adds a subsequent notification action after the skipped approval
//  3. Triggers the workflow on a Control
//  4. Verifies zero approval assignments were created (approval was skipped)
//  5. Confirms the workflow instance completed successfully
//  6. Verifies that a "skipped" action event was recorded
//
// Why This Matters:
//
//	Skipped approval actions should not block the workflow. The engine must recognize that
//	a skipped action contributes zero pending work and advance to subsequent actions.
func TestSkippedApprovalActionAdvancesWorkflow(t *testing.T) {
	s := suite.setupTest(t)

	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	approvalParams := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Required: boolPtr(true),
		Label:    "Reference ID Approval",
		Fields:   []string{"reference_id"},
	}
	approvalParamsBytes, err := json.Marshal(approvalParams)
	assert.NilError(t, err)

	notificationParams := workflows.NotificationActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Title: "Skipped approval notification",
	}
	notificationParamsBytes, err := json.Marshal(notificationParams)
	assert.NilError(t, err)

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Skipped Approval " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetOwnerID(orgID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{"reference_id"}},
			},
			Conditions: []models.WorkflowCondition{
				{Expression: "true"},
			},
			Actions: []models.WorkflowAction{
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "reference_id_approval",
					When:   "false",
					Params: approvalParamsBytes,
				},
				{
					Type:   enums.WorkflowActionTypeNotification.String(),
					Key:    "notify",
					Params: notificationParamsBytes,
				},
			},
		}).
		Save(userCtx)
	assert.NilError(t, err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-SKIP-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetReferenceID("REF-SKIP-" + ulid.Make().String()).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}

	instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"reference_id"},
	})
	assert.NilError(t, err)
	assert.Assert(t, instance != nil)

	s.WaitForEvents()

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(assignments, 0))

	updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStateCompleted, updatedInstance.State))

	events, err := s.client.WorkflowEvent.Query().
		Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	assert.NilError(t, err)
	assert.Check(t, hasSkippedAction(events))
}

// TestApprovalFlowEditSubmittedProposalInvalidatesApprovals verifies that editing a SUBMITTED proposal
// invalidates existing APPROVED assignments, requiring re-approval before the workflow can complete.
// This prevents approved changes from being silently modified after approval.
//
// Workflow Definition (Plain English):
//
//	"Require 2 approvals for Control.reference_id changes"
//
// Test Flow:
//  1. Creates a 2-approver workflow for reference_id changes
//  2. Triggers a proposal for a reference_id change
//  3. Submits the proposal (state = SUBMITTED)
//  4. First approver approves their assignment (status = APPROVED)
//  5. Someone EDITS the proposal, changing the proposed value
//  6. Verifies the first approver's assignment was INVALIDATED (reset to PENDING)
//  7. Verifies invalidation metadata records the reason: "proposal changes edited after approval"
//
// Why This Matters:
//
//	Approvers approve a specific set of changes (identified by hash). If those changes are
//	modified after approval, the previous approval is no longer valid for the new content.
//	This maintains approval integrity and prevents bait-and-switch scenarios.
func TestApprovalFlowEditSubmittedProposalInvalidatesApprovals(t *testing.T) {
	s := suite.setupTest(t)

	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)
	approver2ID, _ := s.CreateTestUserInOrg(orgID, enums.RoleMember)

	// Use engine with listeners so workflow infrastructure is created automatically
	wfEngine := s.Engine()

	// Create approval workflow definition
	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
				{Type: enums.WorkflowTargetTypeUser, ID: approver2ID},
			},
		},
		Required: boolPtr(true),
		// Require both approvals so the proposal remains SUBMITTED after the first approval.
		RequiredCount: 2,
		Label:         "Reference ID Approval",
		Fields:        []string{"reference_id"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "reference_id_approval",
		Params: paramsBytes,
	}

	def := s.CreateApprovalWorkflowDefinition(seedCtx, orgID, action)

	// Create control
	control, err := s.client.Control.Create().
		SetRefCode("CTL-INVALIDATION-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(seedCtx)
	assert.NilError(t, err)

	// Trigger workflow by attempting to update control reference ID
	// The hook will intercept and create proposal + instance
	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	proposedRef := "REF-INVALIDATE-" + ulid.Make().String()

	instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
		EventType:       "UPDATE",
		ChangedFields:   []string{"reference_id"},
		ProposedChanges: map[string]any{"reference_id": proposedRef},
	})
	assert.NilError(t, err)
	assert.Assert(t, instance != nil)
	assert.Assert(t, len(instance.WorkflowProposalID) > 0)

	// Get the proposal that was created by TriggerWorkflow
	proposal, err := s.client.WorkflowProposal.Get(seedCtx, instance.WorkflowProposalID)
	assert.NilError(t, err)

	// Submit the proposal
	_, err = s.client.WorkflowProposal.UpdateOneID(proposal.ID).
		SetState(enums.WorkflowProposalStateSubmitted).
		Save(seedCtx)
	assert.NilError(t, err)

	// Submitting the proposal resumes the paused instance asynchronously and creates assignments
	s.WaitForEvents()

	// Get the assignment that was created
	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(assignments, 2))

	assignmentsByUser := map[string]*generated.WorkflowAssignment{}
	for _, assignment := range assignments {
		switch {
		case strings.HasSuffix(assignment.AssignmentKey, userID):
			assignmentsByUser[userID] = assignment
		case strings.HasSuffix(assignment.AssignmentKey, approver2ID):
			assignmentsByUser[approver2ID] = assignment
		}
	}
	assert.Assert(t, assignmentsByUser[userID] != nil)
	assert.Assert(t, assignmentsByUser[approver2ID] != nil)

	// Approve the assignment
	err = wfEngine.CompleteAssignment(userCtx, assignmentsByUser[userID].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	assert.NilError(t, err)

	// Verify assignment is now APPROVED
	approvedAssignment, err := s.client.WorkflowAssignment.Get(seedCtx, assignmentsByUser[userID].ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusApproved, approvedAssignment.Status))

	// Now edit the SUBMITTED proposal - this should invalidate the approval
	newChanges := map[string]any{"reference_id": "REF-INVALIDATE-EDIT-" + ulid.Make().String()}
	newHash, err := workflows.ComputeProposalHash(newChanges)
	assert.NilError(t, err)

	_, err = s.client.WorkflowProposal.UpdateOneID(proposal.ID).
		SetChanges(newChanges).
		SetProposedHash(newHash).
		Save(seedCtx)
	assert.NilError(t, err)

	// The approval should be invalidated (set back to PENDING)
	invalidatedAssignment, err := s.client.WorkflowAssignment.Get(seedCtx, assignmentsByUser[userID].ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusPending, invalidatedAssignment.Status))

	// Verify invalidation metadata was recorded
	assert.Assert(t, len(invalidatedAssignment.InvalidationMetadata.Reason) > 0)
	assert.Check(t, is.Equal("proposal changes edited after approval", invalidatedAssignment.InvalidationMetadata.Reason))
}

// TestInternalPolicyDetailsApprovalFlow verifies the complete workflow proposal lifecycle
// for InternalPolicy entities, demonstrating that approval workflows work across different
// entity types (not just Controls).
//
// Workflow Definition (Plain English):
//
//	"Require approval before changing InternalPolicy.details (policy content)"
//
// Test Flow:
//  1. Creates an approval workflow for InternalPolicy.details changes
//  2. Creates an InternalPolicy with initial content: "This is the initial policy content."
//  3. Attempts to update details to "This is the UPDATED policy content."
//  4. Verifies the update was intercepted (original content unchanged)
//  5. Verifies a proposal was created with the new content staged
//  6. Resumes the paused instance (simulating proposal submission)
//  7. Verifies an approval assignment was created
//  8. Approves the assignment
//  9. Confirms the proposal was applied and the InternalPolicy.details now shows updated content
//
// Why This Matters:
//
//	Demonstrates that the workflow system is not Control-specific. InternalPolicy (and other
//	entity types) can have their own approval workflows for sensitive field changes like
//	policy content modifications.
func TestInternalPolicyDetailsApprovalFlow(t *testing.T) {
	s := suite.setupTest(t)

	userID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	wfEngine := s.Engine()

	// Create approval workflow definition for InternalPolicy details field
	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Required: boolPtr(true),
		Label:    "Policy Content Approval",
		Fields:   []string{"details"},
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Policy Details Approval " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("InternalPolicy").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetTriggerOperations([]string{"UPDATE"}).
		SetTriggerFields([]string{"details"}).
		SetApprovalSubmissionMode(enums.WorkflowApprovalSubmissionModeAutoSubmit).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{"details"}},
			},
			Actions: []models.WorkflowAction{
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "details_approval",
					Params: paramsBytes,
				},
			},
		}).
		Save(seedCtx)
	assert.NilError(t, err)

	// Create InternalPolicy with initial content
	initialContent := "This is the initial policy content."
	policy, err := s.client.InternalPolicy.Create().
		SetName("Data Security Policy " + ulid.Make().String()).
		SetOwnerID(orgID).
		SetDetails(initialContent).
		Save(seedCtx)
	assert.NilError(t, err)

	// Update the policy details - hook should intercept and create proposal
	proposedContent := "This is the UPDATED policy content."
	updated, err := s.client.InternalPolicy.UpdateOneID(policy.ID).
		SetDetails(proposedContent).
		Save(seedCtx)
	assert.NilError(t, err)

	// Wait for async event processing to complete
	s.WaitForEvents()

	// Original content should be unchanged (mutation was intercepted)
	assert.Check(t, is.DeepEqual(initialContent, updated.Details))

	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.InternalPolicyIDEQ(policy.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowInstanceStatePaused, instance.State))
	assert.Assert(t, len(instance.WorkflowProposalID) > 0)

	proposal, err := s.client.WorkflowProposal.Get(seedCtx, instance.WorkflowProposalID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowProposalStateSubmitted, proposal.State))
	assert.Check(t, is.DeepEqual(proposedContent, proposal.Changes["details"]))

	// Get the workflow definition
	def, err = s.client.WorkflowDefinition.Query().
		Where(workflowdefinition.IDEQ(instance.WorkflowDefinitionID)).
		Only(seedCtx)
	assert.NilError(t, err)

	// Resume the paused instance (simulates what happens when proposal is submitted)
	obj := &workflows.Object{ID: policy.ID, Type: enums.WorkflowObjectTypeInternalPolicy}
	err = wfEngine.TriggerExistingInstance(seedCtx, instance, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"details"},
	})
	assert.NilError(t, err)

	s.WaitForEvents()

	// Verify assignment was created
	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(assignments, 1))
	assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusPending, assignments[0].Status))

	// Complete the approval
	err = wfEngine.CompleteAssignment(seedCtx, assignments[0].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	assert.NilError(t, err)

	s.WaitForEvents()

	// Verify proposal is applied and content is updated
	appliedProposal, err := s.client.WorkflowProposal.Get(seedCtx, proposal.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.WorkflowProposalStateApplied, appliedProposal.State))

	finalPolicy, err := s.client.InternalPolicy.Get(seedCtx, policy.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(proposedContent, finalPolicy.Details))
}

// TestActionWhenExpressionUsesTriggerContext verifies that action-level "when" expressions
// have access to trigger context variables like changed_edges, allowing conditional execution
// based on which edges were modified.
//
// Workflow Definition (Plain English):
//
//	"Send notification, but ONLY when the 'evidence' edge was modified"
//	When expression: 'evidence' in changed_edges
//
// Test Flow:
//
//	Subtest "when expression true executes":
//	  1. Triggers workflow with changed_edges = ["evidence"]
//	  2. Verifies the notification action executed (not skipped)
//
//	Subtest "when expression false skips":
//	  1. Triggers workflow with changed_edges = ["other_edge"]
//	  2. Verifies the notification action was skipped
//	  3. Confirms a "skipped" action event was recorded
//
// Why This Matters:
//
//	Actions can be conditionally executed based on trigger context. This enables workflows
//	that only fire certain actions when specific edges change, providing fine-grained control.
func TestActionWhenExpressionUsesTriggerContext(t *testing.T) {
	s := suite.setupTest(t)
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	action := models.WorkflowAction{
		Type: enums.WorkflowActionTypeNotification.String(),
		Key:  "notify_action",
		When: "'evidence' in changed_edges",
	}

	def, err := s.client.WorkflowDefinition.Create().
		SetName("When Workflow " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindNotification).
		SetSchemaType("Control").
		SetActive(true).
		SetOwnerID(orgID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{"reference_id"}},
			},
			Actions: []models.WorkflowAction{action},
		}).
		Save(userCtx)
	assert.NilError(t, err)

	t.Run("when expression true executes", func(t *testing.T) {
		s := s.withT(t)
		// Each subtest uses its own control to avoid duplicate instance guards
		control, err := s.client.Control.Create().
			SetRefCode("CTL-WHEN-TRUE-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetReferenceID("REF-WHEN-TRUE-" + ulid.Make().String()).
			Save(userCtx)
		assert.NilError(t, err)

		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
		}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"reference_id"},
			ChangedEdges:  []string{"evidence"},
			AddedIDs:      map[string][]string{"evidence": {"evidence-1"}},
		})
		assert.NilError(t, err)
		assert.Assert(t, instance != nil)

		s.WaitForEvents()

		events, err := s.client.WorkflowEvent.Query().
			Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		assert.NilError(t, err)
		assert.Check(t, !(hasSkippedAction(events)))
	})

	t.Run("when expression false skips", func(t *testing.T) {
		s := s.withT(t)
		// Each subtest uses its own control to avoid duplicate instance guards
		control, err := s.client.Control.Create().
			SetRefCode("CTL-WHEN-FALSE-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetReferenceID("REF-WHEN-FALSE-" + ulid.Make().String()).
			Save(userCtx)
		assert.NilError(t, err)

		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
		}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"reference_id"},
			ChangedEdges:  []string{"other_edge"},
			AddedIDs:      map[string][]string{"evidence": {"evidence-1"}},
		})
		assert.NilError(t, err)
		assert.Assert(t, instance != nil)

		s.WaitForEvents()

		events, err := s.client.WorkflowEvent.Query().
			Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		assert.NilError(t, err)
		assert.Check(t, hasSkippedAction(events))
	})
}
