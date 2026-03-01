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
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/jsonx"
)

// integrationOpsSpy captures integration operation requests for assertions
type integrationOpsSpy struct {
	// calls stores recorded operation requests
	calls []types.OperationRequest
}

// notificationIntegrationRegistryStub provides notification operation descriptors for test providers
type notificationIntegrationRegistryStub struct{}

// OperationDescriptors returns provider operation descriptors for notification tests
func (notificationIntegrationRegistryStub) OperationDescriptors(provider types.ProviderType) []types.OperationDescriptor {
	switch provider {
	case "slack":
		return []types.OperationDescriptor{
			{
				Provider: "slack",
				Name:     types.OperationName("message.send"),
				Kind:     types.OperationKindNotify,
			},
		}
	case "microsoftteams":
		return []types.OperationDescriptor{
			{
				Provider: "microsoftteams",
				Name:     types.OperationName("message.send"),
				Kind:     types.OperationKindNotify,
			},
		}
	default:
		return nil
	}
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
		Registry:   notificationIntegrationRegistryStub{},
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
	s.Equal(integrationRecord.ID, call.IntegrationID)
	s.Equal(types.ProviderType("slack"), call.Provider)
	s.Equal(types.OperationName("message.send"), call.Name)
	config, err := jsonx.ToMap(call.Config)
	s.Require().NoError(err)
	s.Equal("C12345", config["channel"])
	s.Equal("Hello https://example.com/review", config["text"])
}

// TestNotificationTemplateIntegrationFromMutation verifies mutation-driven notification integration execution
func (s *WorkflowEngineTestSuite) TestNotificationTemplateIntegrationFromMutation() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	opsSpy := &integrationOpsSpy{}
	err := wfEngine.SetIntegrationDeps(engine.IntegrationDeps{
		Registry:   notificationIntegrationRegistryStub{},
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
		SetKey("workflow.notify.slack.create").
		SetName("Slack Notify Create").
		SetChannel(enums.ChannelSlack).
		SetTopicPattern("workflow.notification").
		SetIntegrationID(integrationRecord.ID).
		SetBodyTemplate("Created {{ref_code}}").
		Save(seedCtx)
	s.Require().NoError(err)

	_, err = s.client.NotificationPreference.Create().
		SetOwnerID(orgID).
		SetUserID(userID).
		SetChannel(enums.ChannelSlack).
		SetDestination("C12345").
		Save(seedCtx)
	s.Require().NoError(err)

	params := workflows.NotificationActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		Channels:    []enums.Channel{enums.ChannelSlack},
		TemplateKey: template.Key,
		Data: map[string]any{
			"ref_code": "CTRL-PLACEHOLDER",
		},
	}
	paramsBytes, err := json.Marshal(params)
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
				Type:   enums.WorkflowActionTypeNotification.String(),
				Key:    "notify_slack_create",
				Params: paramsBytes,
			},
		},
	}

	operations, fields := workflows.DeriveTriggerPrefilter(doc)
	def, err := s.client.WorkflowDefinition.Create().
		SetName("Mutation Notify Workflow " + ulid.Make().String()).
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
		SetRefCode("CTL-NOTIFY-CREATE-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	s.WaitForEvents()

	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.ControlIDEQ(control.ID),
			workflowinstance.OwnerIDEQ(orgID),
		).
		Only(userCtx)
	s.Require().NoError(err)
	s.Equal(enums.WorkflowInstanceStateCompleted, instance.State)

	notifications, err := s.client.Notification.Query().
		Where(
			notification.OwnerIDEQ(orgID),
			notification.UserIDEQ(userID),
			notification.TemplateIDEQ(template.ID),
		).
		All(userCtx)
	s.Require().NoError(err)
	s.Require().Len(notifications, 1)

	s.Require().Len(opsSpy.calls, 1)
	call := opsSpy.calls[0]
	s.Equal(orgID, call.OrgID)
	s.Equal(integrationRecord.ID, call.IntegrationID)
	s.Equal(types.ProviderType("slack"), call.Provider)
	s.Equal(types.OperationName("message.send"), call.Name)
	config, err := jsonx.ToMap(call.Config)
	s.Require().NoError(err)
	s.Equal("C12345", config["channel"])
	s.Equal("Created CTRL-PLACEHOLDER", config["text"])
}
