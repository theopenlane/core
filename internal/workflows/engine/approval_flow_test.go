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
	"github.com/theopenlane/core/internal/ent/generated/workflowdefinition"
	"github.com/theopenlane/core/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/ent/generated/workflowproposal"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
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

// TestApprovalFlowQuorumAppliesProposal verifies quorum approval applies proposals
func (s *WorkflowEngineTestSuite) TestApprovalFlowQuorumAppliesProposal() {
	approver1ID, orgID, userCtx := s.SetupTestUser()
	approver2ID, approver2Ctx := s.CreateTestUserInOrg(orgID, enums.RoleMember)

	wfEngine := s.SetupWorkflowEngineWithListeners()

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
	s.Require().NoError(err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "status_approval",
		Params: paramsBytes,
	}

	def := s.CreateApprovalWorkflowDefinition(userCtx, orgID, action)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetStatus(enums.ControlStatusNotImplemented).
		Save(userCtx)
	s.Require().NoError(err)

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
	s.Require().NoError(err)
	s.Require().NotNil(instance)
	s.Require().NotEmpty(instance.WorkflowProposalID)

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	s.Require().NoError(err)
	s.Require().Len(assignments, 2)

	assignmentsByUser := map[string]*generated.WorkflowAssignment{}
	for _, assignment := range assignments {
		switch {
		case strings.HasSuffix(assignment.AssignmentKey, approver1ID):
			assignmentsByUser[approver1ID] = assignment
		case strings.HasSuffix(assignment.AssignmentKey, approver2ID):
			assignmentsByUser[approver2ID] = assignment
		}
	}
	s.Require().NotNil(assignmentsByUser[approver1ID])
	s.Require().NotNil(assignmentsByUser[approver2ID])

	err = wfEngine.CompleteAssignment(userCtx, assignmentsByUser[approver1ID].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	s.Require().NoError(err)

	proposal, err := s.client.WorkflowProposal.Get(userCtx, instance.WorkflowProposalID)
	s.Require().NoError(err)
	s.NotEqual(enums.WorkflowProposalStateApplied, proposal.State)

	updatedControl, err := s.client.Control.Get(userCtx, control.ID)
	s.Require().NoError(err)
	s.Equal(enums.ControlStatusNotImplemented, updatedControl.Status)

	err = wfEngine.CompleteAssignment(approver2Ctx, assignmentsByUser[approver2ID].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	s.Require().NoError(err)

	proposal, err = s.client.WorkflowProposal.Get(userCtx, instance.WorkflowProposalID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowProposalStateApplied, proposal.State)
	s.NotEmpty(proposal.ApprovedHash)

	updatedControl, err = s.client.Control.Get(userCtx, control.ID)
	s.Require().NoError(err)
	s.Equal(enums.ControlStatusApproved, updatedControl.Status)

	updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowInstanceStateCompleted, updatedInstance.State)
}

// TestApprovalFlowOptionalQuorumProceedsEarly verifies optional approvals can proceed early
func (s *WorkflowEngineTestSuite) TestApprovalFlowOptionalQuorumProceedsEarly() {
	approver1ID, orgID, userCtx := s.SetupTestUser()
	approver2ID, _ := s.CreateTestUserInOrg(orgID, enums.RoleMember)

	wfEngine := s.SetupWorkflowEngineWithListeners()

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approver1ID},
				{Type: enums.WorkflowTargetTypeUser, ID: approver2ID},
			},
		},
		Required: boolPtr(false),
		Label:    "Optional Status Approval",
		Fields:   []string{"status"},
	}
	paramsBytes, err := json.Marshal(params)
	s.Require().NoError(err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "optional_status_approval",
		Params: paramsBytes,
	}

	def := s.CreateApprovalWorkflowDefinition(userCtx, orgID, action)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetStatus(enums.ControlStatusNotImplemented).
		Save(userCtx)
	s.Require().NoError(err)

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
	s.Require().NoError(err)
	s.Require().NotNil(instance)

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	s.Require().NoError(err)
	s.Require().Len(assignments, 2)

	assignmentsByUser := map[string]*generated.WorkflowAssignment{}
	for _, assignment := range assignments {
		switch {
		case strings.HasSuffix(assignment.AssignmentKey, approver1ID):
			assignmentsByUser[approver1ID] = assignment
		case strings.HasSuffix(assignment.AssignmentKey, approver2ID):
			assignmentsByUser[approver2ID] = assignment
		}
	}
	s.Require().NotNil(assignmentsByUser[approver1ID])
	s.Require().NotNil(assignmentsByUser[approver2ID])

	err = wfEngine.CompleteAssignment(userCtx, assignmentsByUser[approver1ID].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	s.Require().NoError(err)

	updatedControl, err := s.client.Control.Get(userCtx, control.ID)
	s.Require().NoError(err)
	s.Equal(enums.ControlStatusApproved, updatedControl.Status)

	updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowInstanceStateCompleted, updatedInstance.State)

	assignments, err = s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	s.Require().NoError(err)

	statuses := make(map[enums.WorkflowAssignmentStatus]int)
	for _, assignment := range assignments {
		statuses[assignment.Status]++
	}

	s.Equal(1, statuses[enums.WorkflowAssignmentStatusApproved])
	s.Equal(1, statuses[enums.WorkflowAssignmentStatusPending])
}

// TestApprovalStagingCapturesClearedField verifies cleared fields are staged
func (s *WorkflowEngineTestSuite) TestApprovalStagingCapturesClearedField() {
	userID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	// Use engine with listeners so workflow infrastructure is created automatically
	wfEngine := s.SetupWorkflowEngineWithListeners()
	s.client.WorkflowEngine = wfEngine

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
	s.Require().NoError(err)

	// Create definition with seedCtx (privacy bypass for setup)
	_, err = s.client.WorkflowDefinition.Create().
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
	s.Require().NoError(err)

	// Create control with seedCtx (privacy bypass for setup)
	control, err := s.client.Control.Create().
		SetRefCode("CTL-REF-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetReferenceID("REF-123").
		Save(seedCtx)
	s.Require().NoError(err)

	// Update with seedCtx - the hook should intercept and route to proposal
	// Using seedCtx since this test is testing workflow hook behavior, not privacy rules
	updated, err := s.client.Control.UpdateOneID(control.ID).
		ClearReferenceID().
		Save(seedCtx)
	s.Require().NoError(err)
	// The hook routes to proposal, so the returned entity should still have the original value
	s.Equal("REF-123", updated.ReferenceID)

	refreshed, err := s.client.Control.Get(seedCtx, control.ID)
	s.Require().NoError(err)
	s.Equal("REF-123", refreshed.ReferenceID)

	domainKey := workflows.DeriveDomainKey([]string{"reference_id"})
	proposal, err := s.client.WorkflowProposal.Query().
		Where(workflowproposal.DomainKeyEQ(domainKey)).
		WithWorkflowObjectRef().
		Only(seedCtx)
	s.Require().NoError(err)
	s.Require().NotNil(proposal.Edges.WorkflowObjectRef)
	s.Equal(control.ID, proposal.Edges.WorkflowObjectRef.ControlID)
	s.Equal(enums.WorkflowProposalStateDraft, proposal.State)

	value, ok := proposal.Changes["reference_id"]
	s.Require().True(ok)
	s.Nil(value)
}

// TestApprovalTriggerExpressionUsesCurrentObjectState verifies trigger expressions see the current object state,
// not the proposed value, when approvals are routed via proposals.
func (s *WorkflowEngineTestSuite) TestApprovalTriggerExpressionUsesCurrentObjectState() {
	userID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	// Enable engine + listeners so hooks and proposal submission are wired.
	wfEngine := s.SetupWorkflowEngineWithListeners()
	s.client.WorkflowEngine = wfEngine

	params := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Required: boolPtr(true),
		Label:    "Status Approval",
		Fields:   []string{"status"},
	}
	paramsBytes, err := json.Marshal(params)
	s.Require().NoError(err)

	_, err = s.client.WorkflowDefinition.Create().
		SetName("Status Approval With Expression " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetTriggerOperations([]string{"UPDATE"}).
		SetTriggerFields([]string{"status"}).
		SetApprovalSubmissionMode(enums.WorkflowApprovalSubmissionModeAutoSubmit).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			ApprovalSubmissionMode: enums.WorkflowApprovalSubmissionModeAutoSubmit,
			Triggers: []models.WorkflowTrigger{
				{
					Operation:  "UPDATE",
					Fields:     []string{"status"},
					Expression: `object.status == "APPROVED"`,
				},
			},
			Actions: []models.WorkflowAction{
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "status_approval",
					Params: paramsBytes,
				},
			},
		}).
		Save(seedCtx)
	s.Require().NoError(err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-STATUS-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetStatus(enums.ControlStatusNotImplemented).
		Save(seedCtx)
	s.Require().NoError(err)

	// Update status; approval hook intercepts and creates a proposal, so the control remains unchanged.
	updated, err := s.client.Control.UpdateOneID(control.ID).
		SetStatus(enums.ControlStatusApproved).
		Save(seedCtx)
	s.Require().NoError(err)
	s.Equal(enums.ControlStatusNotImplemented, updated.Status)

	domainKey := workflows.DeriveDomainKey([]string{"status"})
	proposal, err := s.client.WorkflowProposal.Query().
		Where(workflowproposal.DomainKeyEQ(domainKey)).
		Only(seedCtx)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowProposalStateSubmitted, proposal.State)

	instance, err := s.client.WorkflowInstance.Query().
		Where(workflowinstance.WorkflowProposalIDEQ(proposal.ID)).
		Only(seedCtx)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowInstanceStatePaused, instance.State)

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	s.Require().NoError(err)
	s.Len(assignments, 0)
}

// TestSkippedApprovalActionAdvancesWorkflow verifies skipped approvals advance workflow
func (s *WorkflowEngineTestSuite) TestSkippedApprovalActionAdvancesWorkflow() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.SetupWorkflowEngineWithListeners()

	approvalParams := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Required: boolPtr(true),
		Label:    "Status Approval",
		Fields:   []string{"status"},
	}
	approvalParamsBytes, err := json.Marshal(approvalParams)
	s.Require().NoError(err)

	notificationParams := workflows.NotificationActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Title: "Skipped approval notification",
	}
	notificationParamsBytes, err := json.Marshal(notificationParams)
	s.Require().NoError(err)

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Skipped Approval " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetOwnerID(orgID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{"status"}},
			},
			Conditions: []models.WorkflowCondition{
				{Expression: "true"},
			},
			Actions: []models.WorkflowAction{
				{
					Type:   enums.WorkflowActionTypeApproval.String(),
					Key:    "status_approval",
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
	s.Require().NoError(err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-SKIP-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetStatus(enums.ControlStatusNotImplemented).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}

	instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})
	s.Require().NoError(err)
	s.Require().NotNil(instance)

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	s.Require().NoError(err)
	s.Len(assignments, 0)

	updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowInstanceStateCompleted, updatedInstance.State)

	events, err := s.client.WorkflowEvent.Query().
		Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	s.Require().NoError(err)
	s.True(hasSkippedAction(events))
}

// TestApprovalFlowEditSubmittedProposalInvalidatesApprovals verifies that editing a SUBMITTED proposal
// invalidates existing APPROVED assignments, requiring re-approval before the workflow can complete.
func (s *WorkflowEngineTestSuite) TestApprovalFlowEditSubmittedProposalInvalidatesApprovals() {
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)
	approver2ID, _ := s.CreateTestUserInOrg(orgID, enums.RoleMember)

	// Use engine with listeners so workflow infrastructure is created automatically
	wfEngine := s.SetupWorkflowEngineWithListeners()
	s.client.WorkflowEngine = wfEngine

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
		Label:         "Status Approval",
		Fields:        []string{"status"},
	}
	paramsBytes, err := json.Marshal(params)
	s.Require().NoError(err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "status_approval",
		Params: paramsBytes,
	}

	def := s.CreateApprovalWorkflowDefinition(seedCtx, orgID, action)

	// Create control
	control, err := s.client.Control.Create().
		SetRefCode("CTL-INVALIDATION-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(seedCtx)
	s.Require().NoError(err)

	// Trigger workflow by attempting to update control status
	// The hook will intercept and create proposal + instance
	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
		EventType:       "UPDATE",
		ChangedFields:   []string{"status"},
		ProposedChanges: map[string]any{"status": enums.ControlStatusApproved.String()},
	})
	s.Require().NoError(err)
	s.Require().NotNil(instance)
	s.Require().NotEmpty(instance.WorkflowProposalID)

	// Get the proposal that was created by TriggerWorkflow
	proposal, err := s.client.WorkflowProposal.Get(seedCtx, instance.WorkflowProposalID)
	s.Require().NoError(err)

	// Submit the proposal
	_, err = s.client.WorkflowProposal.UpdateOneID(proposal.ID).
		SetState(enums.WorkflowProposalStateSubmitted).
		Save(seedCtx)
	s.Require().NoError(err)

	// Get the assignment that was created
	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	s.Require().NoError(err)
	s.Require().Len(assignments, 2)

	assignmentsByUser := map[string]*generated.WorkflowAssignment{}
	for _, assignment := range assignments {
		switch {
		case strings.HasSuffix(assignment.AssignmentKey, userID):
			assignmentsByUser[userID] = assignment
		case strings.HasSuffix(assignment.AssignmentKey, approver2ID):
			assignmentsByUser[approver2ID] = assignment
		}
	}
	s.Require().NotNil(assignmentsByUser[userID])
	s.Require().NotNil(assignmentsByUser[approver2ID])

	// Approve the assignment
	err = wfEngine.CompleteAssignment(userCtx, assignmentsByUser[userID].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	s.Require().NoError(err)

	// Verify assignment is now APPROVED
	approvedAssignment, err := s.client.WorkflowAssignment.Get(seedCtx, assignmentsByUser[userID].ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowAssignmentStatusApproved, approvedAssignment.Status)

	// Now edit the SUBMITTED proposal - this should invalidate the approval
	newChanges := map[string]any{"status": enums.ControlStatusNeedsApproval.String()}
	newHash, err := workflows.ComputeProposalHash(newChanges)
	s.Require().NoError(err)

	_, err = s.client.WorkflowProposal.UpdateOneID(proposal.ID).
		SetChanges(newChanges).
		SetProposedHash(newHash).
		Save(seedCtx)
	s.Require().NoError(err)

	// The approval should be invalidated (set back to PENDING)
	invalidatedAssignment, err := s.client.WorkflowAssignment.Get(seedCtx, assignmentsByUser[userID].ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowAssignmentStatusPending, invalidatedAssignment.Status)

	// Verify invalidation metadata was recorded
	s.Require().NotEmpty(invalidatedAssignment.InvalidationMetadata.Reason)
	s.Equal("proposal changes edited after approval", invalidatedAssignment.InvalidationMetadata.Reason)
}

// TestInternalPolicyDetailsApprovalFlow verifies the complete workflow proposal lifecycle
// for policy content changes via actual entity mutation.
func (s *WorkflowEngineTestSuite) TestInternalPolicyDetailsApprovalFlow() {
	userID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	wfEngine := s.SetupWorkflowEngineWithListeners()
	s.client.WorkflowEngine = wfEngine

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
	s.Require().NoError(err)

	_, err = s.client.WorkflowDefinition.Create().
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
	s.Require().NoError(err)

	// Create InternalPolicy with initial content
	initialContent := "This is the initial policy content."
	policy, err := s.client.InternalPolicy.Create().
		SetName("Data Security Policy " + ulid.Make().String()).
		SetOwnerID(orgID).
		SetDetails(initialContent).
		Save(seedCtx)
	s.Require().NoError(err)

	// Update the policy details - hook should intercept and create proposal
	proposedContent := "This is the UPDATED policy content."
	updated, err := s.client.InternalPolicy.UpdateOneID(policy.ID).
		SetDetails(proposedContent).
		Save(seedCtx)
	s.Require().NoError(err)

	// Original content should be unchanged (mutation was intercepted)
	s.Equal(initialContent, updated.Details)

	// Verify proposal was created
	domainKey := workflows.DeriveDomainKey([]string{"details"})
	proposal, err := s.client.WorkflowProposal.Query().
		Where(workflowproposal.DomainKeyEQ(domainKey)).
		Only(seedCtx)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowProposalStateSubmitted, proposal.State)
	s.Equal(proposedContent, proposal.Changes["details"])

	// Verify workflow instance exists (created in PAUSED state by the hook)
	instance, err := s.client.WorkflowInstance.Query().
		Where(workflowinstance.WorkflowProposalIDEQ(proposal.ID)).
		Only(seedCtx)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowInstanceStatePaused, instance.State)

	// Get the workflow definition
	def, err := s.client.WorkflowDefinition.Query().
		Where(workflowdefinition.IDEQ(instance.WorkflowDefinitionID)).
		Only(seedCtx)
	s.Require().NoError(err)

	// Resume the paused instance (simulates what happens when proposal is submitted)
	obj := &workflows.Object{ID: policy.ID, Type: enums.WorkflowObjectTypeInternalPolicy}
	err = wfEngine.TriggerExistingInstance(seedCtx, instance, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"details"},
	})
	s.Require().NoError(err)

	// Verify assignment was created
	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	s.Require().NoError(err)
	s.Require().Len(assignments, 1)
	s.Equal(enums.WorkflowAssignmentStatusPending, assignments[0].Status)

	// Complete the approval
	err = wfEngine.CompleteAssignment(seedCtx, assignments[0].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	s.Require().NoError(err)

	// Verify proposal is applied and content is updated
	appliedProposal, err := s.client.WorkflowProposal.Get(seedCtx, proposal.ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowProposalStateApplied, appliedProposal.State)

	finalPolicy, err := s.client.InternalPolicy.Get(seedCtx, policy.ID)
	s.Require().NoError(err)
	s.Equal(proposedContent, finalPolicy.Details)
}

// TestActionWhenExpressionUsesTriggerContext verifies when expressions use trigger context
func (s *WorkflowEngineTestSuite) TestActionWhenExpressionUsesTriggerContext() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.SetupWorkflowEngineWithListeners()

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
				{Operation: "UPDATE", Fields: []string{"status"}},
			},
			Actions: []models.WorkflowAction{action},
		}).
		Save(userCtx)
	s.Require().NoError(err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-WHEN-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}

	s.Run("when expression true executes", func() {
		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"status"},
			ChangedEdges:  []string{"evidence"},
			AddedIDs:      map[string][]string{"evidence": {"evidence-1"}},
		})
		s.Require().NoError(err)
		s.Require().NotNil(instance)

		events, err := s.client.WorkflowEvent.Query().
			Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.False(hasSkippedAction(events))
	})

	s.Run("when expression false skips", func() {
		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"status"},
			ChangedEdges:  []string{"other_edge"},
			AddedIDs:      map[string][]string{"evidence": {"evidence-1"}},
		})
		s.Require().NoError(err)
		s.Require().NotNil(instance)

		events, err := s.client.WorkflowEvent.Query().
			Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.True(hasSkippedAction(events))
	})
}
