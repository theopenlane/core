//go:build test

package engine_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowevent"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// TestObjectFieldCondition verifies that workflow conditions can evaluate object field values
// to determine whether the workflow should proceed. This enables conditional workflow execution
// based on the state of the target object.
//
// Based on: conditional-approval.json, webhook-notification.json
//
// Test Scenarios:
//
//	Subtest "condition passes when object field matches":
//	  Workflow: Trigger when Control.status == "APPROVED"
//	  Object has status = APPROVED -> Condition passes
//
//	Subtest "condition fails when object field does not match":
//	  Workflow: Trigger when Control.status == "APPROVED"
//	  Object has status = NOT_IMPLEMENTED -> Condition fails
//
//	Subtest "condition with category field check":
//	  Workflow: Trigger when Control.category == "Technical" AND 'status' in changed_fields
//	  Tests both object field evaluation AND trigger context in conditions
//
// Why This Matters:
//
//	Conditions allow workflows to be more selective about when they execute. A workflow
//	might only fire when an object reaches a certain state, preventing unnecessary
//	workflow instances for irrelevant changes.
func (s *WorkflowEngineTestSuite) TestObjectFieldCondition() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	s.Run("condition passes when object field matches", func() {

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

// TestNotificationWorkflowKind verifies that NOTIFICATION-type workflows execute immediately
// without creating approval assignments. These are fire-and-forget workflows.
//
// Based on: webhook-notification.json
//
// Workflow Definition (Plain English):
//
//	"When Control.status changes, send a webhook notification"
//	WorkflowKind = NOTIFICATION (not APPROVAL)
//
// Test Flow:
//  1. Creates a NOTIFICATION workflow with a webhook action
//  2. Sets up a test HTTP server to receive the webhook
//  3. Triggers the workflow on a Control
//  4. Verifies the webhook was called
//  5. Confirms the workflow instance completed immediately (state = COMPLETED)
//  6. Confirms zero approval assignments were created
//
// Why This Matters:
//
//	NOTIFICATION workflows are distinct from APPROVAL workflows. They execute their actions
//	immediately and complete, without pausing for human interaction. This is useful for
//	automated notifications, audit logging, and integrations.
func (s *WorkflowEngineTestSuite) TestNotificationWorkflowKind() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	s.Run("notification workflow triggers and completes without approval", func() {

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

// TestNotificationWorkflowOnCreate verifies CREATE triggers execute NOTIFICATION workflows
// and fire webhook actions without creating approval assignments.
func (s *WorkflowEngineTestSuite) TestNotificationWorkflowOnCreate() {
	_, orgID, userCtx := s.SetupTestUser()

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

	doc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{Operation: "CREATE"},
		},
		Conditions: []models.WorkflowCondition{
			{Expression: "true"},
		},
		Actions: []models.WorkflowAction{
			{
				Type:   enums.WorkflowActionTypeWebhook.String(),
				Key:    "create_webhook",
				Params: webhookParamsBytes,
			},
		},
	}

	operations, fields := workflows.DeriveTriggerPrefilter(doc)

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Create Notification Workflow " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindNotification).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetTriggerOperations(operations).
		SetTriggerFields(fields).
		SetDefinitionJSON(doc).
		Save(userCtx)
	s.Require().NoError(err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-CREATE-WEBHOOK-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	s.WaitForEvents()

	select {
	case <-webhookCalled:
	case <-time.After(5 * time.Second):
		s.FailNow("webhook not received")
	}

	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(userCtx)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowInstanceStateCompleted, instance.State)

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(userCtx)
	s.Require().NoError(err)
	s.Len(assignments, 0)
}

// TestMultiStepParallelApprovals verifies that a single workflow can contain multiple approval
// actions with different "when" clauses, enabling field-specific routing of approvals.
//
// Based on: multi-step-approval.json
//
// Workflow Definition (Plain English):
//
//	Actions:
//	  1. "Technical Review" approval (when: 'category' in changed_fields) - targets User1
//	  2. "Compliance Review" approval (when: 'status' in changed_fields) - targets User2
//
// Test Scenarios:
//
//	Subtest "only category approval when category changes":
//	  - Changed fields: ["category"]
//	  - Expected: Only Technical Review assignment created
//
//	Subtest "only status approval when status changes":
//	  - Changed fields: ["status"]
//	  - Expected: Only Compliance Review assignment created
//
//	Subtest "both approvals when both fields change":
//	  - Changed fields: ["category", "status"]
//	  - Expected: Both Technical Review AND Compliance Review assignments created
//
// Why This Matters:
//
//	Complex approval workflows may require different approvers for different field changes.
//	Using "when" clauses on actions allows a single workflow definition to route approvals
//	based on what actually changed.
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

		s.WaitForEvents()

		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Len(assignments, 1)
		s.Equal("Category Review", assignments[0].Label)
	})

	s.Run("only status approval when status changes", func() {
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

		s.WaitForEvents()

		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Len(assignments, 1)
		s.Equal("Compliance Approval", assignments[0].Label)
	})

	s.Run("both approvals when both fields change", func() {
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

		s.WaitForEvents()

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

// TestApprovalStatusBasedNotifications verifies that notification actions can be triggered
// based on approval assignment status changes, enabling real-time updates as approvals progress.
//
// Based on: multi-approver-with-quorum-notifications.json, approval-with-notifications.json
//
// Workflow Definition (Plain English):
//
//	Actions:
//	  1. "Team Review" approval (requires 2 approvers)
//	  2. "First Approval Received" notification (when: approved == 1 && pending > 0)
//	  3. "Quorum Reached" notification (when: approved >= 2)
//	  4. "Request Rejected" notification (when: rejected > 0)
//
// Test Scenarios:
//
//	Subtest "first approval triggers partial notification":
//	  - First approver approves
//	  - Expected: "First Approval Received" notification fires (1 approved, 1 pending)
//	  - Workflow remains PAUSED (quorum not met)
//
//	Subtest "rejection triggers rejection notification":
//	  - One approver rejects
//	  - Expected: "Request Rejected" notification fires
//
// Why This Matters:
//
//	Approval workflows often need to notify stakeholders as progress occurs. Using
//	assignments.by_action["action_key"].approved/rejected/pending in "when" clauses
//	enables reactive notifications based on approval state transitions.
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

		s.WaitForEvents()

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

		s.WaitForEvents()

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

// TestApprovalWithWebhook verifies that workflows can combine approval and webhook actions,
// with the webhook executing after the approval is completed.
//
// Based on: evidence-review-workflow.json
//
// Workflow Definition (Plain English):
//
//	Actions:
//	  1. "Evidence Review" approval (requires 1 approver)
//	  2. Webhook notification (fires after approval completes)
//
// Test Flow:
//  1. Creates a workflow with an approval followed by a webhook action
//  2. Sets up a test HTTP server to receive the webhook
//  3. Triggers the workflow on a Control
//  4. Verifies an approval assignment was created
//  5. Approves the assignment
//  6. Verifies the webhook was called with the instance ID in the payload
//  7. Confirms the workflow instance completed successfully
//
// Why This Matters:
//
//	Real-world workflows often need to notify external systems after approval. This test
//	confirms that actions execute in sequence and that post-approval webhooks receive
//	appropriate context about the completed workflow.
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

		s.WaitForEvents()

		updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
		s.Require().NoError(err)
		s.Equal(enums.WorkflowInstanceStateCompleted, updatedInstance.State)
	})
}

// TestEdgeTriggerWithCondition verifies that workflows can trigger on edge (relationship)
// changes and use size() functions in conditions to check how many IDs were added.
//
// Based on: evidence-review-workflow.json
//
// Workflow Definition (Plain English):
//
//	"Trigger when the 'controls' edge is modified AND at least one control was added"
//	Trigger: edges = ["controls"]
//	Condition: 'controls' in changed_edges && size(added_ids['controls']) > 0
//
// Test Scenarios:
//
//  1. Changed edges = ["controls"], added_ids = ["control-1", "control-2"]
//     -> Condition passes (edge changed AND controls were added)
//
//  2. Changed edges = ["controls"], added_ids = [] (empty)
//     -> Condition fails (edge changed but no controls added)
//
//  3. Changed edges = ["other_edge"]
//     -> Condition fails ('controls' not in changed_edges)
//
// Why This Matters:
//
//	Edge-based workflows enable triggering on relationship changes, not just field changes.
//	The size() function allows distinguishing between "edge touched" and "items actually added".
func (s *WorkflowEngineTestSuite) TestEdgeTriggerWithCondition() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	s.Run("edge trigger with size condition evaluates correctly", func() {

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

// TestTriggerExistingInstanceResumes verifies the behavior of TriggerExistingInstance, which
// allows resuming paused workflow instances but rejects already-completed instances.
//
// Test Scenarios:
//
//	Subtest "resumes paused instance successfully":
//	  1. Triggers an approval workflow (instance pauses waiting for approval)
//	  2. Calls TriggerExistingInstance on the PAUSED instance
//	  3. Expected: No error (resumption allowed)
//
//	Subtest "rejects completed instance":
//	  1. Triggers a notification workflow (instance completes immediately)
//	  2. Calls TriggerExistingInstance on the COMPLETED instance
//	  3. Expected: ErrInvalidState error (cannot resume completed work)
//
// Why This Matters:
//
//	Proposal submission uses TriggerExistingInstance to resume paused approval workflows.
//	This test ensures the API correctly handles valid (PAUSED) and invalid (COMPLETED) states.
func (s *WorkflowEngineTestSuite) TestTriggerExistingInstanceResumes() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	s.Run("resumes paused instance successfully", func() {

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

// TestTriggerWorkflowGuardsActiveDomainInstance ensures that only one active approval workflow
// instance can exist per domain (definition + object combination). This prevents duplicate
// approval requests for the same change.
//
// Workflow Definition (Plain English):
//
//	"Require approval for Control.status changes"
//
// Test Flow:
//  1. Triggers an approval workflow for a Control (instance created, proposal pending)
//  2. Attempts to trigger the SAME workflow for the SAME object again
//  3. Expected: ErrWorkflowAlreadyActive error (duplicate blocked)
//
// Why This Matters:
//
//	Without this guard, users could trigger multiple parallel approval workflows for the
//	same object change, leading to confusion and potential data inconsistencies. The guard
//	ensures at most one active approval workflow per domain.
func (s *WorkflowEngineTestSuite) TestTriggerWorkflowGuardsActiveDomainInstance() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	s.Run("blocks duplicate approval workflow for same domain", func() {

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

// TestTriggerWorkflowCooldownGuard ensures that non-approval workflows respect cooldown periods,
// preventing rapid re-triggering of the same workflow within a configured time window.
//
// Workflow Definition (Plain English):
//
//	"Send notification on Control.status change, with 1-hour cooldown"
//	CooldownSeconds = 3600
//
// Test Flow:
//  1. Triggers a notification workflow for a Control (completes immediately)
//  2. Immediately attempts to trigger the SAME workflow for the SAME object
//  3. Expected: ErrWorkflowAlreadyActive error (cooldown not elapsed)
//
// Why This Matters:
//
//	Cooldowns prevent notification spam when objects are rapidly updated. Without cooldowns,
//	a burst of updates would generate a burst of notifications. The cooldown ensures
//	workflows have a "quiet period" after execution.
func (s *WorkflowEngineTestSuite) TestTriggerWorkflowCooldownGuard() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	s.Run("cooldown blocks workflow trigger after recent completion", func() {

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

// TestTriggerWorkflowRecordsEmitFailure ensures that when the event emitter fails (e.g., queue
// unavailable), the failure is recorded as a WorkflowEvent for later reconciliation.
//
// Test Flow:
//  1. Creates a workflow engine with a nil/broken emitter
//  2. Triggers a workflow (instance created successfully)
//  3. Verifies an EMIT_FAILED event was recorded in WorkflowEvent table
//  4. The workflow instance should still be created, but the event wasn't published
//
// Why This Matters:
//
//	Event emission failures should not crash the workflow system. By recording failures,
//	a reconciliation process can later retry publishing the events, ensuring eventual
//	consistency even during infrastructure issues.
func (s *WorkflowEngineTestSuite) TestTriggerWorkflowRecordsEmitFailure() {
	_, orgID, userCtx := s.SetupTestUser()

	// Create isolated engine without emitter to test failure recording
	wfEngine := s.NewIsolatedEngine(nil)

	s.Run("records emit failure when emitter is nil", func() {

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

// TestSelectorWithTagMismatch verifies that workflows with tag selectors only match objects
// that have the required tags. Objects without the tag should not trigger the workflow.
//
// Workflow Definition (Plain English):
//
//	"Trigger on Control.status change, but ONLY for Controls tagged with 'RequiredTag'"
//	Trigger selector: { tagIDs: ["tag-123"] }
//
// Test Flow:
//  1. Creates a TagDefinition
//  2. Creates a workflow definition requiring that tag in its selector
//  3. Creates a Control WITHOUT the required tag
//  4. Calls FindMatchingDefinitions for that Control
//  5. Expected: Empty result (no matching definitions)
//
// Why This Matters:
//
//	Tag selectors enable scoping workflows to specific subsets of objects. This is essential
//	for teams that want different approval workflows for different categories of controls
//	(e.g., PCI-tagged controls require extra approval).
func (s *WorkflowEngineTestSuite) TestSelectorWithTagMismatch() {
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	wfEngine := s.Engine()

	s.Run("workflow skipped when object lacks required tag", func() {

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

// TestSelectorWithGroupMismatch verifies that workflows with group selectors only match objects
// associated with the required group. Objects not in the group should not trigger the workflow.
//
// Workflow Definition (Plain English):
//
//	"Trigger on Control.status change, but ONLY for Controls in 'RequiredGroup'"
//	Trigger selector: { groupIDs: ["group-123"] }
//
// Test Flow:
//  1. Creates a Group
//  2. Creates a workflow definition requiring that group in its selector
//  3. Creates a Control NOT associated with the required group
//  4. Calls FindMatchingDefinitions for that Control
//  5. Expected: Empty result (no matching definitions)
//
// Why This Matters:
//
//	Group selectors enable department-specific workflows. Different teams can have different
//	approval requirements for controls they own, without affecting other teams' controls.
func (s *WorkflowEngineTestSuite) TestSelectorWithGroupMismatch() {
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	wfEngine := s.Engine()

	s.Run("workflow skipped when object lacks required group", func() {

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

// TestOptionalApprovalSkipped verifies that optional approval actions with a "when" clause
// evaluating to false are completely skipped, allowing the workflow to complete without blocking.
//
// Workflow Definition (Plain English):
//
//	"Optional approval for Control.status change, but ONLY when condition is true"
//	When expression: "false" (always skips)
//	Required: false (optional)
//
// Test Flow:
//  1. Creates an optional approval workflow with when="false"
//  2. Triggers the workflow on a Control
//  3. Verifies zero approval assignments were created
//  4. Confirms the workflow completed immediately (not blocked)
//
// Why This Matters:
//
//	Optional approvals with "when" clauses allow for conditional approval gates. When the
//	condition is false, the approval is not needed and the workflow should proceed without
//	creating unnecessary assignments.
func (s *WorkflowEngineTestSuite) TestOptionalApprovalSkipped() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	s.Run("optional approval skipped when when clause is false", func() {

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

		s.WaitForEvents()

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

// TestMultipleWebhooksInSequence verifies that multiple webhook actions in a workflow execute
// in the order they are defined, ensuring predictable action sequencing.
//
// Workflow Definition (Plain English):
//
//	Actions:
//	  1. Webhook to server1 (key: "webhook1")
//	  2. Webhook to server2 (key: "webhook2")
//
// Test Flow:
//  1. Sets up two test HTTP servers that record call order
//  2. Creates a workflow with two webhook actions in sequence
//  3. Triggers the workflow
//  4. Verifies webhook1 was called before webhook2
//  5. Confirms the workflow completed successfully
//
// Why This Matters:
//
//	Action ordering is often critical for integrations. For example, you might want to
//	notify an upstream system before notifying downstream consumers. This test ensures
//	the engine respects the defined action order.
func (s *WorkflowEngineTestSuite) TestMultipleWebhooksInSequence() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	s.Run("multiple webhooks execute in sequence", func() {

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

		webhook1Params, _ := json.Marshal(workflows.WebhookActionParams{
			URL:       server1.URL,
			TimeoutMS: 1000,
			Retries:   0,
		})
		webhook2Params, _ := json.Marshal(workflows.WebhookActionParams{
			URL:       server2.URL,
			TimeoutMS: 1000,
			Retries:   0,
		})

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

		readWithTimeout := func(label string) string {
			select {
			case value := <-callOrder:
				return value
			case <-time.After(5 * time.Second):
				s.FailNow("webhook not received", label)
			}
			return ""
		}

		first := readWithTimeout("webhook1")
		second := readWithTimeout("webhook2")
		s.Equal("webhook1", first)
		s.Equal("webhook2", second)

		s.WaitForEvents()

		updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
		s.Require().NoError(err)
		s.Equal(enums.WorkflowInstanceStateCompleted, updatedInstance.State)
	})
}

// TestResolveAssignmentStateTransitions exercises the workflow engine's handling of different
// approval assignment outcomes and their effect on the workflow instance state.
//
// Test Scenarios:
//
//	Subtest "rejected required approval fails instance":
//	  Workflow: Required approval with 1 approver
//	  1. Triggers workflow, assignment created
//	  2. Rejects the assignment
//	  3. Expected: Instance state = FAILED (required approval was rejected)
//
//	Subtest "approved required approval completes instance":
//	  Workflow: Required approval with 1 approver
//	  1. Triggers workflow, assignment created
//	  2. Approves the assignment
//	  3. Expected: Instance state = COMPLETED, proposal applied
//	  4. Verifies the Control.status was updated to the proposed value
//
// Why This Matters:
//
//	The workflow engine must correctly transition instance state based on approval outcomes.
//	Rejection of required approvals should fail the workflow, while approval should complete
//	it and apply the proposed changes.
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

	def, err := s.client.WorkflowDefinition.Create().
		SetName("State Transition Workflow " + ulid.Make().String()).
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

	s.Run("rejected required approval fails instance", func() {
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

		s.WaitForEvents()

		// Instance should be failed
		updatedInstance, err := s.client.WorkflowInstance.Get(userCtx, instance.ID)
		s.Require().NoError(err)
		s.Equal(enums.WorkflowInstanceStateFailed, updatedInstance.State)
	})

	s.Run("approved required approval completes instance", func() {
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

		s.WaitForEvents()

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
