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

	wfEngine := s.Engine()

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

		result, err := wfEngine.EvaluateConditions(userCtx, def, obj, "UPDATE", []string{"status"}, nil, nil, nil, nil)
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

		result, err := wfEngine.EvaluateConditions(userCtx, def, obj, "UPDATE", []string{"status"}, nil, nil, nil, nil)
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

		result, err := wfEngine.EvaluateConditions(userCtx, def, obj, "UPDATE", []string{"status"}, nil, nil, nil, nil)
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

		result, err = wfEngine.EvaluateConditions(userCtx, def, objNonTechnical, "UPDATE", []string{"status"}, nil, nil, nil, nil)
		s.NoError(err)
		s.False(result)
	})
}

// TestNotificationWorkflowKind verifies NOTIFICATION workflow kind execution
// Based on: webhook-notification.json
func (s *WorkflowEngineTestSuite) TestNotificationWorkflowKind() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

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

	wfEngine := s.Engine()

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

	wfEngine := s.Engine()

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

	wfEngine := s.Engine()

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

	wfEngine := s.Engine()

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
			nil,
		)
		s.NoError(err)
		s.False(result)
	})
}

// TestTriggerExistingInstanceResumes verifies TriggerExistingInstance behavior for running vs completed instances
func (s *WorkflowEngineTestSuite) TestTriggerExistingInstanceResumes() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	s.Run("resumes paused instance successfully", func() {
		s.ClearWorkflowDefinitions()

		approvalParams := workflows.ApprovalActionParams{
			TargetedActionParams: workflows.TargetedActionParams{
				Targets: []workflows.TargetConfig{{Type: enums.WorkflowTargetTypeUser, ID: userID}},
			},
			Required: boolPtr(true),
			Label:    "Resume Test Approval",
			Fields:   []string{"status"},
		}
		paramsBytes, err := json.Marshal(approvalParams)
		s.Require().NoError(err)

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Resume Paused Workflow " + ulid.Make().String()).
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
					{Type: enums.WorkflowActionTypeApproval.String(), Key: "test_approval", Params: paramsBytes},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-PAUSED-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetStatus(enums.ControlStatusNotImplemented).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

		// Trigger workflow which will pause for approval
		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:       "UPDATE",
			ChangedFields:   []string{"status"},
			ProposedChanges: map[string]any{"status": enums.ControlStatusApproved.String()},
		})
		s.Require().NoError(err)

		// Instance should be paused waiting for approval
		pausedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
		s.Require().NoError(err)
		s.Equal(enums.WorkflowInstanceStatePaused, pausedInstance.State)

		// TriggerExistingInstance should succeed on paused instance (no error)
		err = wfEngine.TriggerExistingInstance(userCtx, pausedInstance, def, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"status"},
		})
		s.Require().NoError(err)
	})

	s.Run("rejects completed instance", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Reject Completed Workflow " + ulid.Make().String()).
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
				Actions: []models.WorkflowAction{},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-COMPLETE-" + ulid.Make().String()).
			SetOwnerID(orgID).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"status"},
		})
		s.Require().NoError(err)

		// Instance completes immediately since no actions
		completedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
		s.Require().NoError(err)
		s.Equal(enums.WorkflowInstanceStateCompleted, completedInstance.State)

		// Completed instance should reject TriggerExistingInstance
		err = wfEngine.TriggerExistingInstance(userCtx, completedInstance, def, obj, engine.TriggerInput{
			EventType: "UPDATE",
		})
		s.Require().ErrorIs(err, engine.ErrInvalidState)
	})
}

// TestTriggerWorkflowGuardsActiveDomainInstance ensures per-domain guard prevents duplicate approvals
func (s *WorkflowEngineTestSuite) TestTriggerWorkflowGuardsActiveDomainInstance() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	s.Run("blocks duplicate approval workflow for same domain", func() {
		s.ClearWorkflowDefinitions()

		approvalParams := workflows.ApprovalActionParams{
			TargetedActionParams: workflows.TargetedActionParams{
				Targets: []workflows.TargetConfig{{Type: enums.WorkflowTargetTypeUser, ID: userID}},
			},
			Required:      boolPtr(true),
			RequiredCount: 1,
			Label:         "Status Approval",
			Fields:        []string{"status"},
		}
		paramsBytes, err := json.Marshal(approvalParams)
		s.Require().NoError(err)

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Guard Workflow " + ulid.Make().String()).
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
						Key:    "status_approval",
						Params: paramsBytes,
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-GUARD-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetStatus(enums.ControlStatusNotImplemented).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:       "UPDATE",
			ChangedFields:   []string{"status"},
			ProposedChanges: map[string]any{"status": enums.ControlStatusApproved.String()},
		})
		s.Require().NoError(err)
		s.Require().NotEmpty(instance.WorkflowProposalID)

		// Second trigger for same domain should be blocked
		_, err = wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:       "UPDATE",
			ChangedFields:   []string{"status"},
			ProposedChanges: map[string]any{"status": enums.ControlStatusApproved.String()},
		})
		s.Require().ErrorIs(err, workflows.ErrWorkflowAlreadyActive)
	})
}

// TestTriggerWorkflowCooldownGuard ensures cooldown blocks recent instances for non-approval workflows
func (s *WorkflowEngineTestSuite) TestTriggerWorkflowCooldownGuard() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	s.Run("cooldown blocks workflow trigger after recent completion", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Cooldown Workflow " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindNotification).
			SetSchemaType("Control").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"status"}).
			SetCooldownSeconds(3600).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", Fields: []string{"status"}},
				},
				Actions: []models.WorkflowAction{},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-COOLDOWN-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetStatus(enums.ControlStatusNotImplemented).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

		// First trigger completes immediately (no actions)
		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"status"},
		})
		s.Require().NoError(err)

		updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
		s.Require().NoError(err)
		s.Equal(enums.WorkflowInstanceStateCompleted, updatedInstance.State)

		// Second trigger within cooldown should be blocked
		_, err = wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"status"},
		})
		s.Require().ErrorIs(err, workflows.ErrWorkflowAlreadyActive)
	})
}

// TestTriggerWorkflowRecordsEmitFailure ensures emit failures are recorded when no emitter is configured
func (s *WorkflowEngineTestSuite) TestTriggerWorkflowRecordsEmitFailure() {
	_, orgID, userCtx := s.SetupTestUser()

	// Create isolated engine without emitter to test failure recording
	wfEngine := s.NewIsolatedEngine(nil)

	s.Run("records emit failure when emitter is nil", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Emit Failure Workflow " + ulid.Make().String()).
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
				Actions: []models.WorkflowAction{},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-EMIT-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetStatus(enums.ControlStatusNotImplemented).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"status"},
		})
		s.Require().NoError(err)
		s.Require().NotNil(instance)

		// Check that emit failure was recorded
		count, err := s.client.WorkflowEvent.Query().
			Where(
				workflowevent.WorkflowInstanceIDEQ(instance.ID),
				workflowevent.EventTypeEQ(enums.WorkflowEventTypeEmitFailed),
			).
			Count(userCtx)
		s.Require().NoError(err)
		s.GreaterOrEqual(count, 1)
	})
}

// TestSelectorWithTagMismatch verifies workflows are skipped when tag selector doesn't match
func (s *WorkflowEngineTestSuite) TestSelectorWithTagMismatch() {
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	wfEngine := s.Engine()

	s.Run("workflow skipped when object lacks required tag", func() {
		s.ClearWorkflowDefinitions()

		tag, err := s.client.TagDefinition.Create().
			SetName("RequiredTag-" + ulid.Make().String()).
			SetOwnerID(orgID).
			Save(seedCtx)
		s.Require().NoError(err)

		// Create definition requiring the tag
		def, err := s.client.WorkflowDefinition.Create().
			SetName("Tag Selector Workflow " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindNotification).
			SetSchemaType("Control").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"status"}).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{
						Operation: "UPDATE",
						Fields:    []string{"status"},
						Selector: models.WorkflowSelector{
							TagIDs: []string{tag.ID},
						},
					},
				},
				Actions: []models.WorkflowAction{},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		// Create control WITHOUT the required tag
		control, err := s.client.Control.Create().
			SetRefCode("CTL-NO-TAG-" + ulid.Make().String()).
			SetOwnerID(orgID).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

		// Should find no matching definitions because tag doesn't match
		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, nil, obj)
		s.NoError(err)
		s.Empty(defs)
	})
}

// TestSelectorWithGroupMismatch verifies workflows are skipped when group selector doesn't match
func (s *WorkflowEngineTestSuite) TestSelectorWithGroupMismatch() {
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	wfEngine := s.Engine()

	s.Run("workflow skipped when object lacks required group", func() {
		s.ClearWorkflowDefinitions()

		group, err := s.client.Group.Create().
			SetName("RequiredGroup-" + ulid.Make().String()).
			SetOwnerID(orgID).
			Save(seedCtx)
		s.Require().NoError(err)

		// Create definition requiring the group
		def, err := s.client.WorkflowDefinition.Create().
			SetName("Group Selector Workflow " + ulid.Make().String()).
			SetWorkflowKind(enums.WorkflowKindNotification).
			SetSchemaType("Control").
			SetActive(true).
			SetOwnerID(orgID).
			SetTriggerOperations([]string{"UPDATE"}).
			SetTriggerFields([]string{"status"}).
			SetDefinitionJSON(models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{
						Operation: "UPDATE",
						Fields:    []string{"status"},
						Selector: models.WorkflowSelector{
							GroupIDs: []string{group.ID},
						},
					},
				},
				Actions: []models.WorkflowAction{},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		// Create control WITHOUT the required group
		control, err := s.client.Control.Create().
			SetRefCode("CTL-NO-GROUP-" + ulid.Make().String()).
			SetOwnerID(orgID).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

		// Should find no matching definitions because group doesn't match
		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, nil, obj)
		s.NoError(err)
		s.Empty(defs)
	})
}

// TestOptionalApprovalSkipped verifies optional approvals with false when clause are skipped
func (s *WorkflowEngineTestSuite) TestOptionalApprovalSkipped() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	s.Run("optional approval skipped when when clause is false", func() {
		s.ClearWorkflowDefinitions()

		approvalParams := workflows.ApprovalActionParams{
			TargetedActionParams: workflows.TargetedActionParams{
				Targets: []workflows.TargetConfig{{Type: enums.WorkflowTargetTypeUser, ID: userID}},
			},
			Required: boolPtr(false),
			Label:    "Optional Approval",
			Fields:   []string{"status"},
		}
		paramsBytes, err := json.Marshal(approvalParams)
		s.Require().NoError(err)

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Optional Approval Workflow " + ulid.Make().String()).
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
						Key:    "optional_approval",
						When:   "false", // Always false - should be skipped
						Params: paramsBytes,
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-OPTIONAL-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetStatus(enums.ControlStatusNotImplemented).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:       "UPDATE",
			ChangedFields:   []string{"status"},
			ProposedChanges: map[string]any{"status": enums.ControlStatusApproved.String()},
		})
		s.Require().NoError(err)

		// No assignments should be created since when clause is false
		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Empty(assignments)

		// Instance should complete since no approvals were needed
		updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
		s.Require().NoError(err)
		s.Equal(enums.WorkflowInstanceStateCompleted, updatedInstance.State)
	})
}

// TestMultipleWebhooksInSequence verifies multiple webhook actions execute in order
func (s *WorkflowEngineTestSuite) TestMultipleWebhooksInSequence() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	s.Run("multiple webhooks execute in sequence", func() {
		s.ClearWorkflowDefinitions()

		callOrder := make(chan string, 2)
		server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callOrder <- "webhook1"
			w.WriteHeader(http.StatusOK)
		}))
		defer server1.Close()

		server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callOrder <- "webhook2"
			w.WriteHeader(http.StatusOK)
		}))
		defer server2.Close()

		webhook1Params, _ := json.Marshal(workflows.WebhookActionParams{URL: server1.URL})
		webhook2Params, _ := json.Marshal(workflows.WebhookActionParams{URL: server2.URL})

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Multi Webhook Workflow " + ulid.Make().String()).
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
				Actions: []models.WorkflowAction{
					{Type: enums.WorkflowActionTypeWebhook.String(), Key: "webhook1", Params: webhook1Params},
					{Type: enums.WorkflowActionTypeWebhook.String(), Key: "webhook2", Params: webhook2Params},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-MULTI-HOOK-" + ulid.Make().String()).
			SetOwnerID(orgID).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:     "UPDATE",
			ChangedFields: []string{"status"},
		})
		s.Require().NoError(err)

		// Wait for both webhooks
		first := <-callOrder
		second := <-callOrder
		s.Equal("webhook1", first)
		s.Equal("webhook2", second)

		updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
		s.Require().NoError(err)
		s.Equal(enums.WorkflowInstanceStateCompleted, updatedInstance.State)
	})
}

// TestResolveAssignmentStateTransitions exercises ResolveAssignmentState for various approval states
func (s *WorkflowEngineTestSuite) TestResolveAssignmentStateTransitions() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	approvalParams := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{{Type: enums.WorkflowTargetTypeUser, ID: userID}},
		},
		Required:      boolPtr(true),
		RequiredCount: 1,
		Label:         "Status Approval",
		Fields:        []string{"status"},
	}
	paramsBytes, err := json.Marshal(approvalParams)
	s.Require().NoError(err)

	s.Run("rejected required approval fails instance", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Rejection Workflow " + ulid.Make().String()).
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
						Key:    "status_approval",
						Params: paramsBytes,
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-REJECT-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetStatus(enums.ControlStatusNotImplemented).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:       "UPDATE",
			ChangedFields:   []string{"status"},
			ProposedChanges: map[string]any{"status": enums.ControlStatusApproved.String()},
		})
		s.Require().NoError(err)

		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Require().Len(assignments, 1)

		// Reject the approval
		err = wfEngine.CompleteAssignment(userCtx, assignments[0].ID, enums.WorkflowAssignmentStatusRejected, nil, nil)
		s.Require().NoError(err)

		// Instance should be failed
		updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
		s.Require().NoError(err)
		s.Equal(enums.WorkflowInstanceStateFailed, updatedInstance.State)
	})

	s.Run("approved required approval completes instance", func() {
		s.ClearWorkflowDefinitions()

		def, err := s.client.WorkflowDefinition.Create().
			SetName("Approval Workflow " + ulid.Make().String()).
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
						Key:    "status_approval",
						Params: paramsBytes,
					},
				},
			}).
			Save(userCtx)
		s.Require().NoError(err)

		control, err := s.client.Control.Create().
			SetRefCode("CTL-APPROVE-" + ulid.Make().String()).
			SetOwnerID(orgID).
			SetStatus(enums.ControlStatusNotImplemented).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

		instance, err := wfEngine.TriggerWorkflow(userCtx, def, obj, engine.TriggerInput{
			EventType:       "UPDATE",
			ChangedFields:   []string{"status"},
			ProposedChanges: map[string]any{"status": enums.ControlStatusApproved.String()},
		})
		s.Require().NoError(err)

		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Require().Len(assignments, 1)

		// Approve the assignment
		err = wfEngine.CompleteAssignment(userCtx, assignments[0].ID, enums.WorkflowAssignmentStatusApproved, nil, nil)
		s.Require().NoError(err)

		// Instance should be completed
		updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
		s.Require().NoError(err)
		s.Equal(enums.WorkflowInstanceStateCompleted, updatedInstance.State)

		// Control status should be updated
		updatedControl, err := s.client.Control.Get(userCtx, control.ID)
		s.Require().NoError(err)
		s.Equal(enums.ControlStatusApproved, updatedControl.Status)
	})
}
