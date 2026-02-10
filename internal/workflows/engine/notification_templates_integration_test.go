//go:build test

package engine_test

import (
	"context"
	"encoding/json"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated/notification"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// integrationOpsSpy captures integration operation requests for assertions
type integrationOpsSpy struct {
	// calls stores recorded operation requests
	calls []types.OperationRequest
}

// Run records the operation request and returns an ok result
func (s *integrationOpsSpy) Run(_ context.Context, req types.OperationRequest) (types.OperationResult, error) {
	s.calls = append(s.calls, req)
	return types.OperationResult{Status: types.OperationStatusOK, Summary: "ok"}, nil
}

// TestExecuteNotificationWithTemplateIntegration verifies template based integration dispatch
func (s *WorkflowEngineTestSuite) TestExecuteNotificationWithTemplateIntegration() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	opsSpy := &integrationOpsSpy{}
	err := wfEngine.SetIntegrationDeps(engine.IntegrationDeps{
		Operations: opsSpy,
	})
	s.Require().NoError(err)

	seedCtx := s.SeedContext(userID, orgID)

	integrationRecord, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Slack").
		SetKind("slack").
		Save(seedCtx)
	s.Require().NoError(err)

	template, err := s.client.NotificationTemplate.Create().
		SetOwnerID(orgID).
		SetKey("workflow.notify.slack").
		SetName("Slack Notify").
		SetChannel(enums.ChannelSlack).
		SetTopicPattern("workflow.notification").
		SetIntegrationID(integrationRecord.ID).
		SetBodyTemplate("Hello {{review_url}}").
		Save(seedCtx)
	s.Require().NoError(err)

	_, err = s.client.NotificationPreference.Create().
		SetOwnerID(orgID).
		SetUserID(userID).
		SetChannel(enums.ChannelSlack).
		SetDestination("C12345").
		Save(seedCtx)
	s.Require().NoError(err)

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-NOTIFY-TEMPLATE-" + ulid.Make().String()).
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

	params := workflows.NotificationActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Channels:    []enums.Channel{enums.ChannelSlack},
		TemplateKey: template.Key,
		Data: map[string]any{
			"review_url": "https://example.com/review",
		},
	}
	paramsBytes, err := json.Marshal(params)
	s.Require().NoError(err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeNotification.String(),
		Key:    "notify_slack",
		Params: paramsBytes,
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().NoError(err)

	notifications, err := s.client.Notification.Query().
		Where(
			notification.OwnerIDEQ(orgID),
			notification.UserIDEQ(userID),
		).
		All(userCtx)
	s.Require().NoError(err)
	s.Require().Len(notifications, 1)
	s.Equal(template.ID, notifications[0].TemplateID)

	s.Require().Len(opsSpy.calls, 1)
	call := opsSpy.calls[0]
	s.Equal(orgID, call.OrgID)
	s.Equal(types.ProviderType("slack"), call.Provider)
	s.Equal(types.OperationName("message.send"), call.Name)
	s.Equal("C12345", call.Config["channel"])
	s.Equal("Hello https://example.com/review", call.Config["text"])
}
