//go:build test

package engine_test

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// TestApprovalFlowConcurrentApprovalsResumesOnce verifies single resume on concurrent approvals
func (s *WorkflowEngineTestSuite) TestApprovalFlowConcurrentApprovalsResumesOnce() {
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
		Label:         "Concurrent Approval",
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
		SetRefCode("CTL-CONCURRENT-" + ulid.Make().String()).
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
		s.Require().NoError(err)
	}

	proposal, err := s.client.WorkflowProposal.Get(userCtx, instance.WorkflowProposalID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowProposalStateApplied, proposal.State)

	updatedControl, err := s.client.Control.Get(userCtx, control.ID)
	s.Require().NoError(err)
	s.Equal(enums.ControlStatusApproved, updatedControl.Status)

	updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowInstanceStateCompleted, updatedInstance.State)

	events, err := s.client.WorkflowEvent.Query().
		Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	s.Require().NoError(err)

	completedCount := 0
	for _, event := range events {
		switch event.EventType {
		case enums.WorkflowEventTypeInstanceCompleted:
			completedCount++
		}
	}

	s.Equal(1, completedCount)
}

// TestApprovalFlowLateApprovalDoesNotReapply verifies late approvals do not reapply changes
func (s *WorkflowEngineTestSuite) TestApprovalFlowLateApprovalDoesNotReapply() {
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
		RequiredCount: 1,
		Label:         "Single Approval",
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
		SetRefCode("CTL-LATE-" + ulid.Make().String()).
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

	eventsBefore, err := s.client.WorkflowEvent.Query().
		Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	s.Require().NoError(err)

	completedBefore := 0
	for _, event := range eventsBefore {
		switch event.EventType {
		case enums.WorkflowEventTypeInstanceCompleted:
			completedBefore++
		}
	}

	err = wfEngine.CompleteAssignment(approver2Ctx, assignmentsByUser[approver2ID].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
	s.Require().NoError(err)

	eventsAfter, err := s.client.WorkflowEvent.Query().
		Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	s.Require().NoError(err)

	completedAfter := 0
	for _, event := range eventsAfter {
		switch event.EventType {
		case enums.WorkflowEventTypeInstanceCompleted:
			completedAfter++
		}
	}

	s.Equal(completedBefore, completedAfter)

	updatedControl, err := s.client.Control.Get(userCtx, control.ID)
	s.Require().NoError(err)
	s.Equal(enums.ControlStatusApproved, updatedControl.Status)

	updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowInstanceStateCompleted, updatedInstance.State)
}
