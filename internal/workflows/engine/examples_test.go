//go:build test

package engine_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// TestObjectFieldCondition verifies conditions that evaluate object field values
// Based on: conditional-approval.json, webhook-notification.json
func (s *WorkflowEngineTestSuite) TestObjectFieldCondition() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.SetupWorkflowEngineWithListeners()

	s.Run("condition passes when object field matches", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Object Field Condition " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindNotification).
			SetSchemaType("Control").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"status"}).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", Fields: []string{"status"}},
				},
				Conditions: []models.WorkflowCondition{
					{Expression: "object.status == \"APPROVED\""},
				},
				Actions: []models.WorkflowAction{
					{
						Type: enums.WorkflowActionTypeNotification.String(),
						Key:  "status_notification",
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-OBJ-FIELD-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetStatus(enums.ControlStatusApproved).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
			Node: control,
		}

		result, err := wfEngine.EvaluateConditions(userCtx, def, obj, "UPDATE", []string{"status"}, nil, nil, nil)
		s.NoError(err)
		s.True(result)
	})

	s.Run("condition fails when object field does not match", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Object Field Condition Fail " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindNotification).
			SetSchemaType("Control").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"status"}).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", Fields: []string{"status"}},
				},
				Conditions: []models.WorkflowCondition{
					{Expression: "object.status == \"APPROVED\""},
				},
				Actions: []models.WorkflowAction{
					{
						Type: enums.WorkflowActionTypeNotification.String(),
						Key:  "status_notification",
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-OBJ-FIELD-FAIL-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetStatus(enums.ControlStatusNotImplemented).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
			Node: control,
		}

		result, err := wfEngine.EvaluateConditions(userCtx, def, obj, "UPDATE", []string{"status"}, nil, nil, nil)
		s.NoError(err)
		s.False(result)
	})

	s.Run("condition with category field check", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Category Condition " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindApproval).
			SetSchemaType("Control").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"status"}).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", Fields: []string{"status"}},
				},
				Conditions: []models.WorkflowCondition{
					{Expression: "object.category == \"Technical\" && 'status' in changed_fields"},
				},
				Actions: []models.WorkflowAction{
					{
						Type: enums.WorkflowActionTypeNotification.String(),
						Key:  "category_notification",
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-CATEGORY-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetCategory("Technical").
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
			Node: control,
		}

		result, err := wfEngine.EvaluateConditions(userCtx, def, obj, "UPDATE", []string{"status"}, nil, nil, nil)
		s.NoError(err)
		s.True(result)

		controlNonTechnical, err := s.client.Control.Create().
			SetRefCode("CTL-CATEGORY-OTHER-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetCategory("Administrative").
			Save(userCtx)
		s.Require().NoError(err)

		objNonTechnical := &workflows.Object{
			ID:   controlNonTechnical.ID,
			Type: enums.WorkflowObjectTypeControl,
			Node: controlNonTechnical,
		}

		result, err = wfEngine.EvaluateConditions(userCtx, def, objNonTechnical, "UPDATE", []string{"status"}, nil, nil, nil)
		s.NoError(err)
		s.False(result)
	})
}

// TestNotificationWorkflowKind verifies NOTIFICATION workflow kind execution
// Based on: webhook-notification.json
func (s *WorkflowEngineTestSuite) TestNotificationWorkflowKind() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.SetupWorkflowEngineWithListeners()

	s.Run("notification workflow triggers and completes without approval", func() {
		s.ClearWorkflowDefinitions()

		webhookCalled := make(chan struct{}, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			webhookCalled <- struct{}{}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		webhookParams := workflows.WebhookActionParams{
			URL: server.URL,
		}
		webhookParamsBytes, err := json.Marshal(webhookParams)
		s.Require().NoError(err)

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Notification Workflow " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindNotification).
			SetSchemaType("Control").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"status"}).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", Fields: []string{"status"}},
				},
				Conditions: []models.WorkflowCondition{
					{Expression: "true"},
				},
				Actions: []models.WorkflowAction{
					{
						Type:   enums.WorkflowActionTypeWebhook.String(),
						Key:    "slack_webhook",
						Params: webhookParamsBytes,
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-NOTIFICATION-" + ulid.Make().String()).
			SetOwnerID(orgID).
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

		<-webhookCalled

		updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
		s.Require().NoError(err)
		s.Equal(enums.WorkflowInstanceStateCompleted, updatedInstance.State)

		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Len(assignments, 0)
	})
}

// TestMultiStepParallelApprovals verifies multiple approval actions with field-conditional when clauses
// Based on: multi-step-approval.json
func (s *WorkflowEngineTestSuite) TestMultiStepParallelApprovals() {
	userID, orgID, userCtx := s.SetupTestUser()
	user2ID, _ := s.CreateTestUserInOrg(orgID, enums.RoleMember)

	wfEngine := s.SetupWorkflowEngineWithListeners()

	categoryApprovalParams := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Required: boolPtr(true),
		Label:    "Category Review",
		Fields:   []string{"category"},
	}
	categoryParamsBytes, err := json.Marshal(categoryApprovalParams)
	s.Require().NoError(err)

	statusApprovalParams := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: user2ID},
			},
		},
		Required: boolPtr(true),
		Label:    "Compliance Approval",
		Fields:   []string{"status"},
	}
	statusParamsBytes, err := json.Marshal(statusApprovalParams)
	s.Require().NoError(err)

	s.Run("only category approval when category changes", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Multi-Step Approval " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindApproval).
			SetSchemaType("Control").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"category", "status"}).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", Fields: []string{"category", "status"}},
				},
				Conditions: []models.WorkflowCondition{},
				Actions: []models.WorkflowAction{
					{
						Type:   enums.WorkflowActionTypeApproval.String(),
						Key:    "technical_review",
						When:   "'category' in changed_fields",
						Params: categoryParamsBytes,
					},
					{
						Type:   enums.WorkflowActionTypeApproval.String(),
						Key:    "compliance_review",
						When:   "'status' in changed_fields",
						Params: statusParamsBytes,
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-MULTI-STEP-CAT-" + ulid.Make().String()).
			SetOwnerID(orgID).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
		}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:       "UPDATE",
			ChangedFields:   []string{"category"},
			ProposedChanges: map[string]any{"category": "Technical"},
		})
		s.Require().NoError(err)
		s.Require().NotNil(instance)

		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Len(assignments, 1)
		s.Equal("Category Review", assignments[0].Label)
	})

	s.Run("only status approval when status changes", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Multi-Step Approval Status " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindApproval).
			SetSchemaType("Control").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"category", "status"}).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", Fields: []string{"category", "status"}},
				},
				Conditions: []models.WorkflowCondition{},
				Actions: []models.WorkflowAction{
					{
						Type:   enums.WorkflowActionTypeApproval.String(),
						Key:    "technical_review",
						When:   "'category' in changed_fields",
						Params: categoryParamsBytes,
					},
					{
						Type:   enums.WorkflowActionTypeApproval.String(),
						Key:    "compliance_review",
						When:   "'status' in changed_fields",
						Params: statusParamsBytes,
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-MULTI-STEP-STATUS-" + ulid.Make().String()).
			SetOwnerID(orgID).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
		}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:       "UPDATE",
			ChangedFields:   []string{"status"},
			ProposedChanges: map[string]any{"status": enums.ControlStatusApproved.String()},
		})
		s.Require().NoError(err)
		s.Require().NotNil(instance)

		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Len(assignments, 1)
		s.Equal("Compliance Approval", assignments[0].Label)
	})

	s.Run("both approvals when both fields change", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Multi-Step Approval Both " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindApproval).
			SetSchemaType("Control").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"category", "status"}).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", Fields: []string{"category", "status"}},
				},
				Conditions: []models.WorkflowCondition{},
				Actions: []models.WorkflowAction{
					{
						Type:   enums.WorkflowActionTypeApproval.String(),
						Key:    "technical_review",
						When:   "'category' in changed_fields",
						Params: categoryParamsBytes,
					},
					{
						Type:   enums.WorkflowActionTypeApproval.String(),
						Key:    "compliance_review",
						When:   "'status' in changed_fields",
						Params: statusParamsBytes,
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-MULTI-STEP-BOTH-" + ulid.Make().String()).
			SetOwnerID(orgID).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
		}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:       "UPDATE",
			ChangedFields:   []string{"category", "status"},
			ProposedChanges: map[string]any{"category": "Technical", "status": enums.ControlStatusApproved.String()},
		})
		s.Require().NoError(err)
		s.Require().NotNil(instance)

		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Len(assignments, 2)

		labels := make([]string, len(assignments))
		for i, a := range assignments {
			labels[i] = a.Label
		}
		s.Contains(labels, "Category Review")
		s.Contains(labels, "Compliance Approval")
	})
}

// TestApprovalStatusBasedNotifications verifies notifications based on approval status
// Based on: multi-approver-with-quorum-notifications.json, approval-with-notifications.json
func (s *WorkflowEngineTestSuite) TestApprovalStatusBasedNotifications() {
	approver1ID, orgID, userCtx := s.SetupTestUser()
	approver2ID, approver2Ctx := s.CreateTestUserInOrg(orgID, enums.RoleMember)

	wfEngine := s.SetupWorkflowEngineWithListeners()

	approvalParams := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approver1ID},
				{Type: enums.WorkflowTargetTypeUser, ID: approver2ID},
			},
		},
		Required:      boolPtr(true),
		RequiredCount: 2,
		Label:         "Team Review",
		Fields:        []string{"status"},
	}
	approvalParamsBytes, err := json.Marshal(approvalParams)
	s.Require().NoError(err)

	notifyFirstApprovalParams := workflows.NotificationActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approver1ID},
			},
		},
		Title: "First Approval Received",
	}
	notifyFirstParamsBytes, err := json.Marshal(notifyFirstApprovalParams)
	s.Require().NoError(err)

	notifyQuorumParams := workflows.NotificationActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approver1ID},
			},
		},
		Title: "Quorum Reached",
	}
	notifyQuorumParamsBytes, err := json.Marshal(notifyQuorumParams)
	s.Require().NoError(err)

	notifyRejectionParams := workflows.NotificationActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: approver1ID},
			},
		},
		Title: "Request Rejected",
	}
	notifyRejectionParamsBytes, err := json.Marshal(notifyRejectionParams)
	s.Require().NoError(err)

	s.Run("first approval triggers partial notification", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Approval Status Notifications " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindApproval).
			SetSchemaType("Control").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"status"}).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", Fields: []string{"status"}},
				},
				Actions: []models.WorkflowAction{
					{
						Type:   enums.WorkflowActionTypeApproval.String(),
						Key:    "team_review",
						Params: approvalParamsBytes,
					},
					{
						Type:   enums.WorkflowActionTypeNotification.String(),
						Key:    "notify_first_approval",
						When:   "assignments.by_action[\"team_review\"].approved == 1 && assignments.by_action[\"team_review\"].pending > 0",
						Params: notifyFirstParamsBytes,
					},
					{
						Type:   enums.WorkflowActionTypeNotification.String(),
						Key:    "notify_quorum_reached",
						When:   "assignments.by_action[\"team_review\"].approved >= 2",
						Params: notifyQuorumParamsBytes,
					},
					{
						Type:   enums.WorkflowActionTypeNotification.String(),
						Key:    "notify_rejection",
						When:   "assignments.by_action[\"team_review\"].rejected > 0",
						Params: notifyRejectionParamsBytes,
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-APPROVAL-NOTIF-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetStatus(enums.ControlStatusNotImplemented).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
		}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:       "UPDATE",
			ChangedFields:   []string{"status"},
			ProposedChanges: map[string]any{"status": enums.ControlStatusApproved.String()},
		})
		s.Require().NoError(err)
		s.Require().NotNil(instance)

		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Require().Len(assignments, 2)

		assignmentsByUser := map[string]*generated.WorkflowAssignment{}
		for _, a := range assignments {
			switch {
			case strings.HasSuffix(a.AssignmentKey, approver1ID):
				assignmentsByUser[approver1ID] = a
			case strings.HasSuffix(a.AssignmentKey, approver2ID):
				assignmentsByUser[approver2ID] = a
			}
		}
		s.Require().NotNil(assignmentsByUser[approver1ID])

		err = wfEngine.CompleteAssignment(userCtx, assignmentsByUser[approver1ID].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
		s.Require().NoError(err)

		// Check the assignment was actually updated
		updatedAssignment, err := s.client.WorkflowAssignment.Get(userCtx, assignmentsByUser[approver1ID].ID)
		s.Require().NoError(err)
		s.T().Logf("Assignment after completion: status=%s", updatedAssignment.Status)

		events, err := s.client.WorkflowEvent.Query().
			Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.T().Logf("Total events: %d", len(events))

		hasFirstApprovalEvent := false
		for _, e := range events {
			s.T().Logf("Event: type=%s, payload=%s", e.EventType, string(e.Payload.Details))
			if e.EventType == enums.WorkflowEventTypeActionCompleted {
				var payload map[string]any
				if err := json.Unmarshal(e.Payload.Details, &payload); err == nil {
					if actionKey, ok := payload["action_key"].(string); ok && actionKey == "notify_first_approval" {
						hasFirstApprovalEvent = true
					}
					if triggeredBy, ok := payload["triggered_by"].(string); ok && triggeredBy == "assignment_state_change" {
						s.T().Logf("Found assignment_state_change event: %v", payload)
						hasFirstApprovalEvent = true
					}
				}
			}
		}
		s.True(hasFirstApprovalEvent, "expected first approval notification event")

		updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
		s.Require().NoError(err)
		// Workflow should remain PAUSED until all required approvals are completed
		// Only 1 of 2 required approvals has been completed
		s.Equal(enums.WorkflowInstanceStatePaused, updatedInstance.State)
	})

	s.Run("rejection triggers rejection notification", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Rejection Notifications " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindApproval).
			SetSchemaType("Control").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"status"}).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", Fields: []string{"status"}},
				},
				Actions: []models.WorkflowAction{
					{
						Type:   enums.WorkflowActionTypeApproval.String(),
						Key:    "team_review",
						Params: approvalParamsBytes,
					},
					{
						Type:   enums.WorkflowActionTypeNotification.String(),
						Key:    "notify_rejection",
						When:   "assignments.by_action[\"team_review\"].rejected > 0",
						Params: notifyRejectionParamsBytes,
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-REJECTION-NOTIF-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetStatus(enums.ControlStatusNotImplemented).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
		}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:       "UPDATE",
			ChangedFields:   []string{"status"},
			ProposedChanges: map[string]any{"status": enums.ControlStatusApproved.String()},
		})
		s.Require().NoError(err)
		s.Require().NotNil(instance)

		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Require().Len(assignments, 2)

		assignmentsByUser := map[string]*generated.WorkflowAssignment{}
		for _, a := range assignments {
			switch {
			case strings.HasSuffix(a.AssignmentKey, approver1ID):
				assignmentsByUser[approver1ID] = a
			case strings.HasSuffix(a.AssignmentKey, approver2ID):
				assignmentsByUser[approver2ID] = a
			}
		}
		s.Require().NotNil(assignmentsByUser[approver2ID])

		err = wfEngine.CompleteAssignment(approver2Ctx, assignmentsByUser[approver2ID].ID, enums.WorkflowAssignmentStatusRejected, nil, nil)
		s.Require().NoError(err)

		events, err := s.client.WorkflowEvent.Query().
			Where(workflowevent.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)

		hasRejectionEvent := false
		for _, e := range events {
			s.T().Logf("Event: type=%s, payload=%s", e.EventType, string(e.Payload.Details))
			if e.EventType == enums.WorkflowEventTypeActionCompleted {
				var payload map[string]any
				if err := json.Unmarshal(e.Payload.Details, &payload); err == nil {
					if actionKey, ok := payload["action_key"].(string); ok && actionKey == "notify_rejection" {
						hasRejectionEvent = true
					}
					if triggeredBy, ok := payload["triggered_by"].(string); ok && triggeredBy == "assignment_state_change" {
						s.T().Logf("Found assignment_state_change event: %v", payload)
						hasRejectionEvent = true
					}
				}
			}
		}
		s.True(hasRejectionEvent, "expected rejection notification event")
	})
}

// TestApprovalWithWebhook verifies approval followed by webhook execution
// Based on: evidence-review-workflow.json
func (s *WorkflowEngineTestSuite) TestApprovalWithWebhook() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.SetupWorkflowEngineWithListeners()

	webhookCalled := make(chan map[string]any, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)
		webhookCalled <- payload
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	approvalParams := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Required: boolPtr(true),
		Label:    "Evidence Review",
		Fields:   []string{"status"},
	}
	approvalParamsBytes, err := json.Marshal(approvalParams)
	s.Require().NoError(err)

	webhookParams := workflows.WebhookActionParams{
		URL: server.URL,
	}
	webhookParamsBytes, err := json.Marshal(webhookParams)
	s.Require().NoError(err)

	s.Run("webhook fires after approval completes", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Approval with Webhook " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindApproval).
			SetSchemaType("Control").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"status"}).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", Fields: []string{"status"}},
				},
				Actions: []models.WorkflowAction{
					{
						Type:   enums.WorkflowActionTypeApproval.String(),
						Key:    "evidence_review",
						Params: approvalParamsBytes,
					},
					{
						Type:   enums.WorkflowActionTypeWebhook.String(),
						Key:    "notify_review_complete",
						Params: webhookParamsBytes,
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-APPROVAL-WEBHOOK-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetStatus(enums.ControlStatusNotImplemented).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
		}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:       "UPDATE",
			ChangedFields:   []string{"status"},
			ProposedChanges: map[string]any{"status": enums.ControlStatusApproved.String()},
		})
		s.Require().NoError(err)
		s.Require().NotNil(instance)

		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Require().Len(assignments, 1)

		err = wfEngine.CompleteAssignment(userCtx, assignments[0].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
		s.Require().NoError(err)

		payload := <-webhookCalled
		s.Equal(instance.ID, payload["instance_id"])

		updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
		s.Require().NoError(err)
		s.Equal(enums.WorkflowInstanceStateCompleted, updatedInstance.State)
	})
}

// TestEdgeTriggerWithCondition verifies edge-based triggers with size() conditions
// Based on: evidence-review-workflow.json
func (s *WorkflowEngineTestSuite) TestEdgeTriggerWithCondition() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.NewTestEngine(nil)

	s.Run("edge trigger with size condition evaluates correctly", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Edge Trigger Size " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindApproval).
			SetSchemaType("Evidence").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"controls"}).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{
						Operation: "UPDATE",
						Edges:     []string{"controls"},
					},
				},
				Conditions: []models.WorkflowCondition{
					{Expression: "'controls' in changed_edges && size(added_ids['controls']) > 0"},
				},
				Actions: []models.WorkflowAction{
					{
						Type: enums.WorkflowActionTypeNotification.String(),
						Key:  "evidence_linked",
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{
			ID:   "evidence-123",
			Type: enums.WorkflowObjectTypeEvidence,
		}

		result, err := wfEngine.EvaluateConditions(
			userCtx,
			def,
			obj,
			"UPDATE",
			nil,
			[]string{"controls"},
			map[string][]string{"controls": {"control-1", "control-2"}},
			nil,
		)
		s.NoError(err)
		s.True(result)

		result, err = wfEngine.EvaluateConditions(
			userCtx,
			def,
			obj,
			"UPDATE",
			nil,
			[]string{"controls"},
			map[string][]string{"controls": {}},
			nil,
		)
		s.NoError(err)
		s.False(result)

		result, err = wfEngine.EvaluateConditions(
			userCtx,
			def,
			obj,
			"UPDATE",
			nil,
			[]string{"other_edge"},
			map[string][]string{"other_edge": {"id-1"}},
			nil,
		)
		s.NoError(err)
		s.False(result)
	})
}
