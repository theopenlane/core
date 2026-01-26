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

// TestWorkflowEngineExecute verifies action execution
func (s *WorkflowEngineTestSuite) TestWorkflowEngineExecute() {
	wfEngine := s.Engine()
	s.Require().NotNil(wfEngine)
}

// TestExecuteApproval verifies approval action execution
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
		s.NoError(err)
	})
}

// TestExecuteInvalidActionType verifies invalid action handling
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

// TestExecuteNotification verifies notification action execution
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

// TestExecuteWebhook verifies webhook action execution
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

// TestApplyObjectFieldUpdates_CoercesEnums verifies enum coercion for updates
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

// TestExecuteFieldUpdate verifies field update action execution
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
