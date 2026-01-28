//go:build test

package engine_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// TestWorkflowEngineExecute verifies basic workflow engine initialization and action execution
// infrastructure is correctly set up.
//
// Test Flow:
//  1. Retrieves the workflow engine from the test suite
//  2. Verifies the engine is not nil (properly initialized)
//
// Why This Matters:
//   Foundational test ensuring the workflow engine is available for action execution.
func (s *WorkflowEngineTestSuite) TestWorkflowEngineExecute() {
	wfEngine := s.Engine()
	s.Require().NotNil(wfEngine)
}

// TestExecuteApproval verifies the approval action executor correctly creates WorkflowAssignment
// and WorkflowAssignmentTarget records when processing an approval action.
//
// Test Scenarios:
//   Subtest "creates approval assignment with user target":
//     1. Executes an approval action targeting a specific user
//     2. Verifies a WorkflowAssignment was created with:
//        - Status = PENDING
//        - Required = true
//        - Label = "Test Approval"
//     3. Verifies a WorkflowAssignmentTarget was created linking to the user
//
//   Subtest "creates assignment with no targets warns":
//     1. Executes an approval action with an empty targets list
//     2. Expected: ErrApprovalNoTargets error (cannot create assignment without approvers)
//
// Why This Matters:
//   The approval executor is responsible for creating the database records that represent
//   pending approval work. This test ensures the executor correctly translates action
//   parameters into database entities.
func (s *WorkflowEngineTestSuite) TestExecuteApproval() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-TEST-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	s.Run("creates approval assignment with user target", func() {
		targets := []workflows.TargetConfig{
			{
				Type: enums.WorkflowTargetTypeUser,
				ID:   userID,
			},
		}

		params := workflows.ApprovalActionParams{
			TargetedActionParams: workflows.TargetedActionParams{
				Targets: targets,
			},
			Required: boolPtr(true),
			Label:    "Test Approval",
		}

		paramsBytes, err := json.Marshal(params)
		s.Require().NoError(err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeApproval.String(),
			Key:    "test_approval",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		s.NoError(err)

		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Len(assignments, 1)

		assignment := assignments[0]
		s.Equal(enums.WorkflowAssignmentStatusPending, assignment.Status)
		s.Equal(true, assignment.Required)
		s.Equal("Test Approval", assignment.Label)

		assignmentTargets, err := s.client.WorkflowAssignmentTarget.Query().
			Where(workflowassignmenttarget.WorkflowAssignmentIDEQ(assignment.ID)).
			All(userCtx)
		s.Require().NoError(err)
		s.Len(assignmentTargets, 1)
		s.Equal(userID, assignmentTargets[0].TargetUserID)
	})

	s.Run("creates assignment with no targets warns", func() {
		params := workflows.ApprovalActionParams{
			TargetedActionParams: workflows.TargetedActionParams{
				Targets: []workflows.TargetConfig{},
			},
			Required: boolPtr(false),
			Label:    "No Targets",
		}

		paramsBytes, err := json.Marshal(params)
		s.Require().NoError(err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeApproval.String(),
			Key:    "no_targets",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		s.ErrorIs(err, engine.ErrApprovalNoTargets)
	})
}

// TestExecuteInvalidActionType verifies that the workflow engine correctly rejects unknown
// action types with an appropriate error.
//
// Test Flow:
//  1. Creates an action with Type = "INVALID_TYPE"
//  2. Attempts to execute the action
//  3. Expected: ErrInvalidActionType error
//
// Why This Matters:
//   The workflow engine must validate action types before execution. Unknown action types
//   should fail fast with a clear error rather than silently passing or causing cryptic failures.
func (s *WorkflowEngineTestSuite) TestExecuteInvalidActionType() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-FIELD-TEST-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	action := models.WorkflowAction{
		Type: "INVALID_TYPE",
		Key:  "test",
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Error(err)
	s.ErrorIs(err, engine.ErrInvalidActionType)
}

// TestExecuteNotification verifies that notification actions execute without error.
// Notifications are typically no-op or delegate to external notification systems.
//
// Test Flow:
//  1. Creates a notification action
//  2. Executes the action
//  3. Expected: No error (action completes successfully)
//
// Why This Matters:
//   Notification actions should complete without blocking the workflow, even if the
//   actual notification delivery is async or external.
func (s *WorkflowEngineTestSuite) TestExecuteNotification() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-NOTIFY-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	err = wfEngine.Execute(userCtx, models.WorkflowAction{Type: enums.WorkflowActionTypeNotification.String(), Key: "test_notification"}, instance, obj)
	s.NoError(err)
}

// TestExecuteWebhook verifies the webhook action executor correctly calls external URLs
// and handles various success and failure scenarios.
//
// Test Scenarios:
//   Subtest "executes webhook with valid URL":
//     1. Sets up a test HTTP server
//     2. Executes a webhook action with URL pointing to the server
//     3. Verifies the server received the request
//     4. Verifies the payload contains the instance ID
//     5. Verifies the X-Workflow-Signature header is present (when secret configured)
//
//   Subtest "fails webhook without URL":
//     1. Executes a webhook action with no URL configured
//     2. Expected: ErrWebhookURLRequired error
//
//   Subtest "fails webhook on non-success status":
//     1. Sets up a server that returns HTTP 418 (Teapot)
//     2. Executes a webhook action
//     3. Expected: Error (non-2xx response codes are failures)
//
// Why This Matters:
//   Webhook actions are the primary integration point for external systems. This test
//   ensures correct HTTP behavior, error handling, and security features like HMAC signing.
func (s *WorkflowEngineTestSuite) TestExecuteWebhook() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-WEBHOOK-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	s.Run("executes webhook with valid URL", func() {
		var receivedPayload map[string]any
		serverCalled := make(chan struct{}, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &receivedPayload)
			if r.Header.Get("X-Workflow-Signature") == "" {
				s.T().Errorf("expected signature header to be set")
			}
			serverCalled <- struct{}{}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		params := workflows.WebhookActionParams{
			URL:    server.URL,
			Secret: "test-secret",
		}

		paramsBytes, err := json.Marshal(params)
		s.Require().NoError(err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeWebhook.String(),
			Key:    "test_webhook",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		s.Require().NoError(err)

		<-serverCalled
		s.Equal(instance.ID, receivedPayload["instance_id"])
	})

	s.Run("fails webhook without URL", func() {
		params := workflows.WebhookActionParams{}

		paramsBytes, err := json.Marshal(params)
		s.Require().NoError(err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeWebhook.String(),
			Key:    "test_webhook",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		s.Error(err)
		s.ErrorIs(err, engine.ErrWebhookURLRequired)
	})

	s.Run("fails webhook on non-success status", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		}))
		defer server.Close()

		params := workflows.WebhookActionParams{
			URL: server.URL,
		}

		paramsBytes, err := json.Marshal(params)
		s.Require().NoError(err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeWebhook.String(),
			Key:    "test_webhook",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		s.Error(err)
	})
}

// TestApplyObjectFieldUpdates_CoercesEnums verifies that ApplyObjectFieldUpdates correctly
// coerces string enum values to their typed enum equivalents when updating entities.
//
// Test Flow:
//  1. Creates a Control with initial status
//  2. Calls ApplyObjectFieldUpdates with status = "APPROVED" (string, not enum type)
//  3. Verifies the Control.Status is now enums.ControlStatusApproved (typed enum)
//
// Why This Matters:
//   Proposal changes are stored as map[string]any where enum values are serialized as strings.
//   When applying these changes, the workflow system must coerce strings back to typed enums
//   to match the entity schema. Without coercion, database updates would fail with type errors.
func (s *WorkflowEngineTestSuite) TestApplyObjectFieldUpdates_CoercesEnums() {
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-APPLY-TEST-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(seedCtx)
	s.Require().NoError(err)

	// Use AllowContext for workflow operations that need privacy bypass
	bypassCtx := workflows.AllowContext(userCtx)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}

	changes := map[string]any{
		"status": enums.ControlStatusApproved.String(),
	}

	err = workflows.ApplyObjectFieldUpdates(bypassCtx, s.client, obj.Type, obj.ID, changes)
	s.Require().NoError(err)

	updated, err := s.client.Control.Get(bypassCtx, control.ID)
	s.Require().NoError(err)

	s.Equal(enums.ControlStatusApproved, updated.Status)
}

// TestExecuteFieldUpdate verifies the field_update action executor correctly applies
// field changes to the target object.
//
// Test Scenarios:
//   Subtest "executes field update with valid params":
//     1. Creates a Control with initial status
//     2. Executes a field_update action with updates = { "status": "APPROVED" }
//     3. Verifies the Control.status is now APPROVED
//
//   Subtest "fails field update without updates":
//     1. Executes a field_update action with empty updates
//     2. Expected: ErrMissingRequiredField error
//
// Why This Matters:
//   Field update actions enable workflows to programmatically modify object fields as part
//   of their execution. This is useful for automated status transitions after approval
//   or other workflow-driven field changes.
func (s *WorkflowEngineTestSuite) TestExecuteFieldUpdate() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-FIELD-UPDATE-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	s.Run("executes field update with valid params", func() {
		params := workflows.FieldUpdateActionParams{
			Updates: map[string]any{
				"status": "APPROVED",
			},
		}

		paramsBytes, err := json.Marshal(params)
		s.Require().NoError(err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeFieldUpdate.String(),
			Key:    "test_update",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		s.NoError(err)

		updated, err := s.client.Control.Get(userCtx, obj.ID)
		s.Require().NoError(err)
		s.Equal(enums.ControlStatusApproved, updated.Status)
	})

	s.Run("fails field update without updates", func() {
		params := workflows.FieldUpdateActionParams{}

		paramsBytes, err := json.Marshal(params)
		s.Require().NoError(err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeFieldUpdate.String(),
			Key:    "test_update",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		s.Error(err)
		s.ErrorIs(err, engine.ErrMissingRequiredField)
	})
}
