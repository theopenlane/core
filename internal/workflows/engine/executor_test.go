//go:build test

package engine_test

import (
	"testing"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated/emailtemplate"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignmenttarget"
	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
	"github.com/theopenlane/core/v2/pkg/urlx"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// TestWorkflowEngineExecute verifies basic workflow engine initialization and action execution
// infrastructure is correctly set up.
//
// Test Flow:
//  1. Retrieves the workflow engine from the test suite
//  2. Verifies the engine is not nil (properly initialized)
//
// Why This Matters:
//
//	Foundational test ensuring the workflow engine is available for action execution.
func TestWorkflowEngineExecute(t *testing.T) {
	s := suite.setupParallelTest(t)
	wfEngine := s.Engine()
	assert.Assert(t, wfEngine != nil)
}

// TestExecuteApproval verifies the approval action executor correctly creates WorkflowAssignment
// and WorkflowAssignmentTarget records when processing an approval action.
//
// Test Scenarios:
//
//	Subtest "creates approval assignment with user target":
//	  1. Executes an approval action targeting a specific user
//	  2. Verifies a WorkflowAssignment was created with:
//	     - Status = PENDING
//	     - Required = true
//	     - Label = "Test Approval"
//	  3. Verifies a WorkflowAssignmentTarget was created linking to the user
//
//	Subtest "creates assignment with no targets warns":
//	  1. Executes an approval action with an empty targets list
//	  2. Expected: ErrApprovalNoTargets error (cannot create assignment without approvers)
//
// Why This Matters:
//
//	The approval executor is responsible for creating the database records that represent
//	pending approval work. This test ensures the executor correctly translates action
//	parameters into database entities.
func TestExecuteApproval(t *testing.T) {
	s := suite.setupParallelTest(t)
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-TEST-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	t.Run("creates approval assignment with user target", func(t *testing.T) {
		s := s.withT(t)
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
		assert.NilError(t, err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeApproval.String(),
			Key:    "test_approval",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		assert.Check(t, err)

		assignments, err := s.client.WorkflowAssignment.Query().
			Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
			All(userCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Len(assignments, 1))

		assignment := assignments[0]
		assert.Check(t, is.DeepEqual(enums.WorkflowAssignmentStatusPending, assignment.Status))
		assert.Check(t, is.Equal(true, assignment.Required))
		assert.Check(t, is.Equal("Test Approval", assignment.Label))

		assignmentTargets, err := s.client.WorkflowAssignmentTarget.Query().
			Where(workflowassignmenttarget.WorkflowAssignmentIDEQ(assignment.ID)).
			All(userCtx)
		assert.NilError(t, err)
		assert.Check(t, is.Len(assignmentTargets, 1))
		assert.Check(t, is.DeepEqual(userID, assignmentTargets[0].TargetUserID))
	})

	t.Run("creates assignment with no targets warns", func(t *testing.T) {
		params := workflows.ApprovalActionParams{
			TargetedActionParams: workflows.TargetedActionParams{
				Targets: []workflows.TargetConfig{},
			},
			Required: boolPtr(false),
			Label:    "No Targets",
		}

		paramsBytes, err := json.Marshal(params)
		assert.NilError(t, err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeApproval.String(),
			Key:    "no_targets",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		assert.Check(t, is.ErrorIs(err, engine.ErrApprovalNoTargets))
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
//
//	The workflow engine must validate action types before execution. Unknown action types
//	should fail fast with a clear error rather than silently passing or causing cryptic failures.
func TestExecuteInvalidActionType(t *testing.T) {
	s := suite.setupParallelTest(t)
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-FIELD-TEST-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

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
	assert.Check(t, err != nil)
	assert.Check(t, is.ErrorIs(err, engine.ErrInvalidActionType))
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
//
//	Notification actions should complete without blocking the workflow, even if the
//	actual notification delivery is async or external.
func TestExecuteNotification(t *testing.T) {
	s := suite.setupParallelTest(t)
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-NOTIFY-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	err = wfEngine.Execute(userCtx, models.WorkflowAction{Type: enums.WorkflowActionTypeNotification.String(), Key: "test_notification"}, instance, obj)
	assert.Check(t, err)
}

// TestExecuteSendEmail verifies SEND_EMAIL action composition and queue insertion.
func TestExecuteSendEmail(t *testing.T) {
	s := suite.setupTest(t)
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()
	def := s.CreateTestWorkflowDefinition(userCtx, orgID)
	seedCtx := s.SeedContext(userID, orgID)

	err := s.client.Job.TruncateRiverTables(context.Background())
	assert.NilError(t, err)

	emailRecord, err := s.client.EmailTemplate.Create().
		SetOwnerID(orgID).
		SetKey(emaildef.BrandedMessageOp.Name()).
		SetName("Workflow Send Email").
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetTemplateContext(enums.TemplateContextWorkflowAction).
		SetDefaults(map[string]any{
			"subject": "Workflow Send Email",
			"title":   "Workflow Send Email",
			"intros":  []any{"Hi from workflow"},
		}).
		SetActive(true).
		Save(seedCtx)
	assert.NilError(t, err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-SEND-EMAIL-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	params := workflows.SendEmailActionParams{
		EmailTemplateID: emailRecord.ID,
		To:              []string{"person@example.com"},
		From:            "noreply@example.com",
	}

	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeSendEmail.String(),
		Key:    "send_email",
		Params: paramsBytes,
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	assert.NilError(t, err)
	s.WaitForEvents()

	msgs := s.mockEmailSender().Messages()
	assert.Assert(t, is.Len(msgs, 1))
	assert.Check(t, is.DeepEqual([]string{"person@example.com"}, msgs[0].To))
	assert.Check(t, is.Equal("Workflow Send Email", msgs[0].Subject))
	assert.Check(t, is.Contains(msgs[0].Text, "Hi from workflow"))

	_, err = s.client.EmailTemplate.Delete().Where(emailtemplate.NameEQ("Workflow Send Email")).Exec(seedCtx)
	assert.NilError(t, err)
}

// TestExecuteWebhook verifies the webhook action executor correctly calls external URLs
// and handles various success and failure scenarios.
//
// Test Scenarios:
//
//	Subtest "executes webhook with valid URL":
//	  1. Sets up a test HTTP server
//	  2. Executes a webhook action with URL pointing to the server
//	  3. Verifies the server received the request
//	  4. Verifies the payload contains the instance ID
//	  5. Verifies the X-Workflow-Signature header is present (when secret configured)
//
//	Subtest "fails webhook without URL":
//	  1. Executes a webhook action with no URL configured
//	  2. Expected: ErrWebhookURLRequired error
//
//	Subtest "fails webhook on non-success status":
//	  1. Sets up a server that returns HTTP 418 (Teapot)
//	  2. Executes a webhook action
//	  3. Expected: Error (non-2xx response codes are failures)
//
// Why This Matters:
//
//	Webhook actions are the primary integration point for external systems. This test
//	ensures correct HTTP behavior, error handling, and security features like HMAC signing.
func TestExecuteWebhook(t *testing.T) {
	s := suite.setupTest(t)
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-WEBHOOK-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	t.Run("executes webhook with valid URL", func(t *testing.T) {
		s := s.withT(t)
		var receivedPayload map[string]any
		serverCalled := make(chan struct{}, 1)
		customHeaderKey := "X-Custom-Header"
		customHeaderValue := "custom-value"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &receivedPayload)
			if r.Header.Get("X-Workflow-Signature") == "" {
				s.T().Errorf("expected signature header to be set")
			}
			if r.Header.Get(customHeaderKey) != customHeaderValue {
				s.T().Errorf("expected custom header %s to be set", customHeaderKey)
			}
			serverCalled <- struct{}{}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		params := workflows.WebhookActionParams{
			URL:    server.URL,
			Secret: "test-secret",
			Headers: map[string]string{
				customHeaderKey: customHeaderValue,
			},
		}

		paramsBytes, err := json.Marshal(params)
		assert.NilError(t, err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeWebhook.String(),
			Key:    "test_webhook",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		assert.NilError(t, err)

		<-serverCalled
		assert.Check(t, is.DeepEqual(instance.ID, receivedPayload["instance_id"]))
	})

	t.Run("executes webhook with payload expression", func(t *testing.T) {
		var receivedPayload map[string]any
		serverCalled := make(chan struct{}, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &receivedPayload)
			serverCalled <- struct{}{}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		params := workflows.WebhookActionParams{
			URL:         server.URL,
			PayloadExpr: `{"custom": "ok", "event_type": event_type}`,
		}

		paramsBytes, err := json.Marshal(params)
		assert.NilError(t, err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeWebhook.String(),
			Key:    "test_webhook_expr",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		assert.NilError(t, err)

		<-serverCalled
		assert.Check(t, is.DeepEqual(instance.ID, receivedPayload["instance_id"]))
		assert.Check(t, is.Equal("ok", receivedPayload["custom"]))
		assert.Check(t, is.Equal("UPDATE", receivedPayload["event_type"]))
	})

	t.Run("fails webhook with legacy payload", func(t *testing.T) {
		params := json.RawMessage(`{"url": "https://example.com/webhook", "payload": {"text": "nope"}}`)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeWebhook.String(),
			Key:    "test_webhook_legacy_payload",
			Params: params,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		assert.Check(t, err != nil)
		assert.Check(t, is.ErrorIs(err, engine.ErrWebhookPayloadUnsupported))
	})

	t.Run("fails webhook without URL", func(t *testing.T) {
		params := workflows.WebhookActionParams{}

		paramsBytes, err := json.Marshal(params)
		assert.NilError(t, err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeWebhook.String(),
			Key:    "test_webhook",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		assert.Check(t, err != nil)
		assert.Check(t, is.ErrorIs(err, engine.ErrWebhookURLRequired))
	})

	t.Run("fails webhook on non-success status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		}))
		defer server.Close()

		params := workflows.WebhookActionParams{
			URL: server.URL,
		}

		paramsBytes, err := json.Marshal(params)
		assert.NilError(t, err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeWebhook.String(),
			Key:    "test_webhook",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		assert.Check(t, err != nil)
	})

	t.Run("blocks webhook to non-public address", func(t *testing.T) {
		s := s.withT(t)
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.WriteHeader(http.StatusOK)
		}))
		s.T().Cleanup(server.Close)

		guardedEngine := s.NewIsolatedEngine(nil)

		paramsBytes, err := json.Marshal(workflows.WebhookActionParams{URL: server.URL})
		assert.NilError(t, err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeWebhook.String(),
			Key:    "test_webhook_private",
			Params: paramsBytes,
		}

		err = guardedEngine.Execute(userCtx, action, instance, obj)
		assert.Check(t, is.ErrorIs(err, engine.ErrWebhookFailed))
		assert.Check(t, is.ErrorIs(err, urlx.ErrNonPublicDestination))
		assert.Check(t, is.Equal(int32(0), calls.Load()))
	})

	t.Run("fails redirect without following or retrying", func(t *testing.T) {
		s := s.withT(t)
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			http.Redirect(w, r, "/moved", http.StatusFound)
		}))
		s.T().Cleanup(server.Close)

		paramsBytes, err := json.Marshal(workflows.WebhookActionParams{URL: server.URL})
		assert.NilError(t, err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeWebhook.String(),
			Key:    "test_webhook_redirect",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		assert.Check(t, is.ErrorIs(err, engine.ErrWebhookFailed))
		assert.Check(t, is.DeepEqual(int32(1), calls.Load()))
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
//
//	Proposal changes are stored as map[string]any where enum values are serialized as strings.
//	When applying these changes, the workflow system must coerce strings back to typed enums
//	to match the entity schema. Without coercion, database updates would fail with type errors.
func TestApplyObjectFieldUpdates_CoercesEnums(t *testing.T) {
	s := suite.setupParallelTest(t)
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-APPLY-TEST-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(seedCtx)
	assert.NilError(t, err)

	bypassCtx := auth.WithInternalCrossOrgContext(userCtx)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}

	changes := map[string]any{
		"status": enums.ControlStatusApproved.String(),
	}

	err = workflows.ApplyObjectFieldUpdates(bypassCtx, s.client, obj.Type, obj.ID, changes)
	assert.NilError(t, err)

	updated, err := s.client.Control.Get(bypassCtx, control.ID)
	assert.NilError(t, err)

	assert.Check(t, is.DeepEqual(enums.ControlStatusApproved, updated.Status))
}

// TestExecuteFieldUpdate verifies the field_update action executor correctly applies
// field changes to the target object.
//
// Test Scenarios:
//
//	Subtest "executes field update with valid params":
//	  1. Creates a Control with initial status
//	  2. Executes a field_update action with updates = { "status": "APPROVED" }
//	  3. Verifies the Control.status is now APPROVED
//
//	Subtest "fails field update without updates":
//	  1. Executes a field_update action with empty updates
//	  2. Expected: ErrMissingRequiredField error
//
// Why This Matters:
//
//	Field update actions enable workflows to programmatically modify object fields as part
//	of their execution. This is useful for automated status transitions after approval
//	or other workflow-driven field changes.
func TestExecuteFieldUpdate(t *testing.T) {
	s := suite.setupParallelTest(t)
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-FIELD-UPDATE-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	t.Run("executes field update with valid params", func(t *testing.T) {
		s := s.withT(t)
		params := workflows.FieldUpdateActionParams{
			Updates: map[string]any{
				"status": "APPROVED",
			},
		}

		paramsBytes, err := json.Marshal(params)
		assert.NilError(t, err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeFieldUpdate.String(),
			Key:    "test_update",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		assert.Check(t, err)

		updated, err := s.client.Control.Get(userCtx, obj.ID)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(enums.ControlStatusApproved, updated.Status))
	})

	t.Run("fails field update without updates", func(t *testing.T) {
		params := workflows.FieldUpdateActionParams{}

		paramsBytes, err := json.Marshal(params)
		assert.NilError(t, err)

		action := models.WorkflowAction{
			Type:   enums.WorkflowActionTypeFieldUpdate.String(),
			Key:    "test_update",
			Params: paramsBytes,
		}

		err = wfEngine.Execute(userCtx, action, instance, obj)
		assert.Check(t, err != nil)
		assert.Check(t, is.ErrorIs(err, engine.ErrMissingRequiredField))
	})
}
