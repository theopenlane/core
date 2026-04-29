//go:build test

package engine_test

import (
	"context"
	"encoding/json"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/internal/ent/generated/notification"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/registry"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/gala"
)

const (
	testDefinitionID  = "def_01K0TEST000000000000000001"
	testMessageTopic  = gala.TopicName("test.notification.message.send")
	testOperationName = "message.send"
)

func notificationTestDefinition() types.Definition {
	return types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID: testDefinitionID,
		},
		Operations: []types.OperationRegistration{
			{
				Name:  testOperationName,
				Topic: testMessageTopic,
				Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
					return json.RawMessage(`{"ok":true}`), nil
				},
			},
		},
	}
}

// registerNotificationTestTopics registers noop listeners for test notification topics in the gala registry
func registerNotificationTestTopics(registry *gala.Registry) {
	_, _ = gala.RegisterListeners(registry, gala.Definition[operations.Envelope]{
		Topic: gala.Topic[operations.Envelope]{Name: testMessageTopic},
		Name:  "test.noop.message.send",
		Handle: func(gala.HandlerContext, operations.Envelope) error {
			return nil
		},
	})
}

func (s *WorkflowEngineTestSuite) newNotificationTestRuntime() *integrationsruntime.Runtime {
	reg := registry.New()
	s.Require().NoError(reg.Register(notificationTestDefinition()))

	credStore, err := keystore.NewStore(s.client)
	s.Require().NoError(err)

	rt, err := integrationsruntime.New(integrationsruntime.Config{
		DB:       s.client,
		Gala:     s.galaRuntime,
		Registry: reg,
		Keystore: credStore,
	})
	s.Require().NoError(err)

	return rt
}

// TestExecuteNotificationWithTemplateIntegration verifies template based integration dispatch
func (s *WorkflowEngineTestSuite) TestExecuteNotificationWithTemplateIntegration() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()
	rt := s.newNotificationTestRuntime()

	registerNotificationTestTopics(s.galaRuntime.Registry())

	err := wfEngine.SetIntegrationDeps(engine.IntegrationDeps{Runtime: rt})
	s.Require().NoError(err)

	seedCtx := s.SeedContext(userID, orgID)

	integrationRecord, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Test Notification Integration").
		SetDefinitionID(testDefinitionID).
		Save(seedCtx)
	s.Require().NoError(err)

	template, err := s.client.NotificationTemplate.Create().
		SetOwnerID(orgID).
		SetKey("workflow.notify.integration").
		SetName("Integration Notify").
		SetChannel(enums.ChannelSlack).
		SetTopicPattern("workflow.notification").
		SetIntegrationID(integrationRecord.ID).
		SetBodyTemplate("Hello {{.review_url}}").
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
		TemplateKey:   template.Key,
		OperationName: testOperationName,
		Data: map[string]any{
			"review_url": "https://example.com/review",
		},
	}
	paramsBytes, err := json.Marshal(params)
	s.Require().NoError(err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeNotification.String(),
		Key:    "notify_integration",
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

	runs, err := s.client.IntegrationRun.Query().
		Where(integrationrun.IntegrationIDEQ(integrationRecord.ID)).
		All(seedCtx)
	s.Require().NoError(err)
	s.Require().Len(runs, 1)
}

// TestNotificationTemplateIntegrationFromMutation verifies mutation-driven notification integration execution
func (s *WorkflowEngineTestSuite) TestNotificationTemplateIntegrationFromMutation() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()
	rt := s.newNotificationTestRuntime()

	registerNotificationTestTopics(s.galaRuntime.Registry())

	err := wfEngine.SetIntegrationDeps(engine.IntegrationDeps{Runtime: rt})
	s.Require().NoError(err)

	seedCtx := s.SeedContext(userID, orgID)

	integrationRecord, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Test Notification Integration").
		SetDefinitionID(testDefinitionID).
		Save(seedCtx)
	s.Require().NoError(err)

	template, err := s.client.NotificationTemplate.Create().
		SetOwnerID(orgID).
		SetKey("workflow.notify.create").
		SetName("Notify Create").
		SetChannel(enums.ChannelSlack).
		SetTopicPattern("workflow.notification").
		SetIntegrationID(integrationRecord.ID).
		SetBodyTemplate("Created {{.ref_code}}").
		Save(seedCtx)
	s.Require().NoError(err)

	params := workflows.NotificationActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
		TemplateKey:   template.Key,
		OperationName: testOperationName,
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
				Key:    "notify_create",
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

	runs, err := s.client.IntegrationRun.Query().
		Where(integrationrun.IntegrationIDEQ(integrationRecord.ID)).
		All(seedCtx)
	s.Require().NoError(err)
	s.Require().Len(runs, 1)
}

// TestExecuteNotificationWithTemplateDestinationsIntegration verifies template destinations dispatch without user targets
func (s *WorkflowEngineTestSuite) TestExecuteNotificationWithTemplateDestinationsIntegration() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()
	rt := s.newNotificationTestRuntime()

	registerNotificationTestTopics(s.galaRuntime.Registry())

	err := wfEngine.SetIntegrationDeps(engine.IntegrationDeps{Runtime: rt})
	s.Require().NoError(err)

	seedCtx := s.SeedContext(userID, orgID)

	integrationRecord, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Test Notification Integration").
		SetDefinitionID(testDefinitionID).
		Save(seedCtx)
	s.Require().NoError(err)

	template, err := s.client.NotificationTemplate.Create().
		SetOwnerID(orgID).
		SetKey("workflow.notify.destinations." + ulid.Make().String()).
		SetName("Template Destinations").
		SetChannel(enums.ChannelSlack).
		SetTopicPattern("workflow.notification").
		SetIntegrationID(integrationRecord.ID).
		SetDestinations([]string{"C11111", "C22222"}).
		SetBodyTemplate("Hello {{.review_url}}").
		Save(seedCtx)
	s.Require().NoError(err)

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-NOTIFY-TEMPLATE-DEST-" + ulid.Make().String()).
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
		TemplateKey:   template.Key,
		OperationName: testOperationName,
		Data: map[string]any{
			"review_url": "https://example.com/review",
		},
	}
	paramsBytes, err := json.Marshal(params)
	s.Require().NoError(err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeNotification.String(),
		Key:    "notify_template_destinations",
		Params: paramsBytes,
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().NoError(err)

	// No user targets, so no in-app notifications
	notifications, err := s.client.Notification.Query().
		Where(
			notification.OwnerIDEQ(orgID),
			notification.TemplateIDEQ(template.ID),
		).
		All(userCtx)
	s.Require().NoError(err)
	s.Require().Len(notifications, 0)

	// Template has integration_id, so one integration run
	runs, err := s.client.IntegrationRun.Query().
		Where(integrationrun.IntegrationIDEQ(integrationRecord.ID)).
		All(seedCtx)
	s.Require().NoError(err)
	s.Require().Len(runs, 1)

	// Verify destinations are passed through in the config
	destinationsValue, ok := runs[0].OperationConfig["destinations"]
	s.Require().True(ok)
	s.ElementsMatch([]string{"C11111", "C22222"}, stringSliceValue(destinationsValue))
}

func stringSliceValue(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			if str, ok := item.(string); ok {
				values = append(values, str)
			}
		}
		return values
	default:
		return nil
	}
}
