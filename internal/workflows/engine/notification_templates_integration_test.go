//go:build test

package engine_test

import (
	"encoding/json"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated/integrationrun"
	"github.com/theopenlane/core/internal/ent/generated/notification"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/gala"
)

const (
	testSlackDefinitionID = "def_01K0SLACK000000000000000001"
	testTeamsDefinitionID = "def_01K0MSTEAMS00000000000000001"
	testSlackMessageTopic = gala.TopicName("test.slack.message.send")
	testTeamsMessageTopic = gala.TopicName("test.microsoft_teams.message.send")
)

// notificationV2RegistryStub implements registry.DefinitionRegistry for notification tests
type notificationV2RegistryStub struct{}

func (notificationV2RegistryStub) Definition(id types.DefinitionID) (types.Definition, bool) {
	switch string(id) {
	case testSlackDefinitionID:
		return notificationTestSlackDefinition(), true
	case testTeamsDefinitionID:
		return notificationTestTeamsDefinition(), true
	default:
		return types.Definition{}, false
	}
}

func (notificationV2RegistryStub) Client(types.DefinitionID, types.ClientName) (types.ClientRegistration, error) {
	return types.ClientRegistration{}, nil
}

func (notificationV2RegistryStub) Operation(id types.DefinitionID, name types.OperationName) (types.OperationRegistration, error) {
	def, ok := notificationV2RegistryStub{}.Definition(id)
	if !ok {
		return types.OperationRegistration{}, engine.ErrIntegrationOperationCriteriaRequired
	}
	for _, op := range def.Operations {
		if op.Name == name {
			return op, nil
		}
	}
	return types.OperationRegistration{}, engine.ErrIntegrationOperationCriteriaRequired
}

func (notificationV2RegistryStub) OperationFromString(definitionID, name string) (types.OperationRegistration, error) {
	return notificationV2RegistryStub{}.Operation(types.DefinitionID(definitionID), types.OperationName(name))
}

func (notificationV2RegistryStub) Catalog() []types.DefinitionSpec {
	return []types.DefinitionSpec{
		notificationTestSlackDefinition().Spec,
		notificationTestTeamsDefinition().Spec,
	}
}

func (notificationV2RegistryStub) Listeners() []types.OperationRegistration {
	slack := notificationTestSlackDefinition()
	teams := notificationTestTeamsDefinition()
	return []types.OperationRegistration{slack.Operations[0], teams.Operations[0]}
}

func notificationTestSlackDefinition() types.Definition {
	return types.Definition{
		Spec: types.DefinitionSpec{
			ID:   testSlackDefinitionID,
			Slug: "slack",
		},
		Operations: []types.OperationRegistration{
			{
				Name:  "message.send",
				Topic: testSlackMessageTopic,
			},
		},
	}
}

func notificationTestTeamsDefinition() types.Definition {
	return types.Definition{
		Spec: types.DefinitionSpec{
			ID:   testTeamsDefinitionID,
			Slug: "microsoft_teams",
		},
		Operations: []types.OperationRegistration{
			{
				Name:  "message.send",
				Topic: testTeamsMessageTopic,
			},
		},
	}
}

// registerNotificationTestTopics registers noop listeners for test notification topics in the gala registry
func registerNotificationTestTopics(registry *gala.Registry) {
	_, _ = gala.RegisterListeners(registry, gala.Definition[operations.Envelope]{
		Topic: gala.Topic[operations.Envelope]{Name: testSlackMessageTopic},
		Name:  "test.noop.slack.message.send",
		Handle: func(gala.HandlerContext, operations.Envelope) error {
			return nil
		},
	})
	_, _ = gala.RegisterListeners(registry, gala.Definition[operations.Envelope]{
		Topic: gala.Topic[operations.Envelope]{Name: testTeamsMessageTopic},
		Name:  "test.noop.teams.message.send",
		Handle: func(gala.HandlerContext, operations.Envelope) error {
			return nil
		},
	})
}

// TestExecuteNotificationWithTemplateIntegration verifies template based integration dispatch
func (s *WorkflowEngineTestSuite) TestExecuteNotificationWithTemplateIntegration() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	runStore, err := operations.NewRunStore(s.client)
	s.Require().NoError(err)

	registerNotificationTestTopics(s.galaRuntime.Registry())

	dispatcher, err := operations.NewDispatcher(notificationV2RegistryStub{}, s.client, runStore, s.galaRuntime)
	s.Require().NoError(err)

	err = wfEngine.SetIntegrationDeps(engine.IntegrationDeps{
		Registry:   notificationV2RegistryStub{},
		Dispatcher: dispatcher,
	})
	s.Require().NoError(err)

	seedCtx := s.SeedContext(userID, orgID)

	integrationRecord, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Slack").
		SetDefinitionSlug("slack").
		SetDefinitionID(testSlackDefinitionID).
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

	runs, err := s.client.IntegrationRun.Query().
		Where(integrationrun.IntegrationIDEQ(integrationRecord.ID)).
		All(seedCtx)
	s.Require().NoError(err)
	s.Require().Len(runs, 1)
	s.Equal("message.send", runs[0].OperationName)
	s.Equal("C12345", runs[0].OperationConfig["channel"])
	s.Equal("Hello https://example.com/review", runs[0].OperationConfig["text"])
}

// TestNotificationTemplateIntegrationFromMutation verifies mutation-driven notification integration execution
func (s *WorkflowEngineTestSuite) TestNotificationTemplateIntegrationFromMutation() {
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	runStore, err := operations.NewRunStore(s.client)
	s.Require().NoError(err)

	registerNotificationTestTopics(s.galaRuntime.Registry())

	dispatcher, err := operations.NewDispatcher(notificationV2RegistryStub{}, s.client, runStore, s.galaRuntime)
	s.Require().NoError(err)

	err = wfEngine.SetIntegrationDeps(engine.IntegrationDeps{
		Registry:   notificationV2RegistryStub{},
		Dispatcher: dispatcher,
	})
	s.Require().NoError(err)

	seedCtx := s.SeedContext(userID, orgID)

	integrationRecord, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Slack").
		SetDefinitionSlug("slack").
		SetDefinitionID(testSlackDefinitionID).
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

	runs, err := s.client.IntegrationRun.Query().
		Where(integrationrun.IntegrationIDEQ(integrationRecord.ID)).
		All(seedCtx)
	s.Require().NoError(err)
	s.Require().Len(runs, 1)
	s.Equal("message.send", runs[0].OperationName)
}
