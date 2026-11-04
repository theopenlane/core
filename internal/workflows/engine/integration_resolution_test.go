//go:build test

package engine_test

import (
	"encoding/json"
	"testing"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationrun"
	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	slackdef "github.com/theopenlane/core/v2/internal/integrations/definitions/slack"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// TestExecuteIntegrationAction_ConnectedIntegration verifies that executeIntegrationAction
// resolves a connected integration and successfully queues the operation
func TestExecuteIntegrationAction_ConnectedIntegration(t *testing.T) {
	s := suite.setupParallelTest(t)
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	seedCtx := s.SeedContext(userID, orgID)

	integrationRecord, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Connected Email Integration").
		SetDefinitionID(emaildef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusConnected).
		Save(seedCtx)
	assert.NilError(t, err)

	templateKey := "integ-resolution-connected-" + ulid.Make().String()
	emailTemplateID := s.createTestEmailTemplate(userID, orgID, templateKey, "Connected Test", "body")
	t.Cleanup(func() { s.cleanupTestEmailTemplate(userID, orgID, templateKey) })

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-CONNECTED-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	configBytes, err := json.Marshal(emaildef.SendEmailRequest{
		TemplateID: emailTemplateID,
		OwnerID:    orgID,
		To:         "connected@example.com",
	})
	assert.NilError(t, err)

	params := workflows.IntegrationActionParams{
		InstallationID: integrationRecord.ID,
		DefinitionID:   emaildef.DefinitionID.ID(),
		OperationName:  emaildef.SendEmailOp.Name(),
		Config:         configBytes,
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeIntegration.String(),
		Key:    "send_connected",
		Params: paramsBytes,
	}

	// Execute should succeed -- integration is connected and the operation is queued
	err = wfEngine.Execute(userCtx, action, instance, obj)
	assert.NilError(t, err)
}

// TestExecuteIntegrationAction_DisconnectedReResolved verifies that when the stored
// installation is disabled but a replacement connected installation exists for the same
// definition, the engine re-resolves to the new installation and queues successfully
func TestExecuteIntegrationAction_DisconnectedReResolved(t *testing.T) {
	s := suite.setupParallelTest(t)
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	seedCtx := s.SeedContext(userID, orgID)

	// Create the original integration and disable it
	oldIntegration, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Old Email Integration").
		SetDefinitionID(emaildef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusDisabled).
		Save(seedCtx)
	assert.NilError(t, err)

	// Create a replacement connected integration for the same definition
	_, err = s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("New Email Integration").
		SetDefinitionID(emaildef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusConnected).
		Save(seedCtx)
	assert.NilError(t, err)

	templateKey := "integ-resolution-reresolved-" + ulid.Make().String()
	emailTemplateID := s.createTestEmailTemplate(userID, orgID, templateKey, "Re-Resolved Test", "body")
	t.Cleanup(func() { s.cleanupTestEmailTemplate(userID, orgID, templateKey) })

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-RERESOLVE-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	configBytes, err := json.Marshal(emaildef.SendEmailRequest{
		TemplateID: emailTemplateID,
		OwnerID:    orgID,
		To:         "reresolved@example.com",
	})
	assert.NilError(t, err)

	// Reference the OLD (disabled) installation ID -- engine should re-resolve
	params := workflows.IntegrationActionParams{
		InstallationID: oldIntegration.ID,
		DefinitionID:   emaildef.DefinitionID.ID(),
		OperationName:  emaildef.SendEmailOp.Name(),
		Config:         configBytes,
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeIntegration.String(),
		Key:    "send_reresolved",
		Params: paramsBytes,
	}

	// Execute should succeed -- engine re-resolves from disabled to new connected installation
	err = wfEngine.Execute(userCtx, action, instance, obj)
	assert.NilError(t, err)
}

// TestExecuteIntegrationAction_DisconnectedNoReplacement verifies that when the stored
// installation is disabled and no replacement exists, the action is skipped gracefully
func TestExecuteIntegrationAction_DisconnectedNoReplacement(t *testing.T) {
	s := suite.setupParallelTest(t)
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	seedCtx := s.SeedContext(userID, orgID)

	// Create an integration and disable it -- no replacement
	disconnectedIntegration, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Disconnected Email Integration").
		SetDefinitionID(emaildef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusDisabled).
		Save(seedCtx)
	assert.NilError(t, err)

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-DISCONNECTED-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	configBytes, err := json.Marshal(map[string]any{
		"templateId": "fake-template-id",
		"ownerId":    orgID,
		"to":         "disconnected@example.com",
	})
	assert.NilError(t, err)

	params := workflows.IntegrationActionParams{
		InstallationID: disconnectedIntegration.ID,
		DefinitionID:   emaildef.DefinitionID.ID(),
		OperationName:  emaildef.SendEmailOp.Name(),
		Config:         configBytes,
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeIntegration.String(),
		Key:    "send_disconnected",
		Params: paramsBytes,
	}

	// Should succeed (skip gracefully) rather than error
	err = wfEngine.Execute(userCtx, action, instance, obj)
	assert.NilError(t, err)

	// No integration runs should have been created
	runs, err := s.client.IntegrationRun.Query().
		Where(integrationrun.IntegrationIDEQ(disconnectedIntegration.ID)).
		All(seedCtx)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(runs, 0))
}

// TestExecuteSendEmail_DisconnectedFallsBackToRuntime verifies that when an email template
// references a disabled integration, dispatchSendEmail falls through to the runtime
// dispatch path and successfully sends the email
func TestExecuteSendEmail_DisconnectedFallsBackToRuntime(t *testing.T) {
	s := suite.setupTest(t)
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	seedCtx := s.SeedContext(userID, orgID)

	// Create a disabled integration
	disconnectedIntegration, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Disabled Email").
		SetDefinitionID(emaildef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusDisabled).
		Save(seedCtx)
	assert.NilError(t, err)

	// Create an email template that references the disabled integration
	templateKey := "integ-disconnected-fallback-" + ulid.Make().String()
	emailRecord, err := s.client.EmailTemplate.Create().
		SetOwnerID(orgID).
		SetKey(emaildef.BrandedMessageOp.Name()).
		SetName("Email: " + templateKey).
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetTemplateContext(enums.TemplateContextWorkflowAction).
		SetIntegrationID(disconnectedIntegration.ID).
		SetDefaults(map[string]any{
			"subject": "Fallback Subject",
			"title":   "Fallback Subject",
			"intros":  []any{"Fallback body text"},
		}).
		SetActive(true).
		Save(seedCtx)
	assert.NilError(t, err)

	t.Cleanup(func() {
		_, _ = s.client.EmailTemplate.Delete().Where().Exec(seedCtx)
	})

	s.clearEmailState()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-FALLBACK-" + ulid.Make().String()).
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
		To:              []string{"fallback@example.com"},
	}

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeSendEmail.String(),
		Key:    "send_fallback",
		Params: s.mustMarshalSendEmailParams(params),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	assert.NilError(t, err)

	s.WaitForEvents()

	// The email should have been sent via runtime dispatch (not the disabled integration)
	msgs := s.mockEmailSender().Messages()
	assert.Assert(t, is.Len(msgs, 1))
	assert.Check(t, is.DeepEqual([]string{"fallback@example.com"}, msgs[0].To))
	assert.Check(t, is.Equal("Fallback Subject", msgs[0].Subject))
}

// TestExecuteIntegrationAction_SlackMessageSendWithTemplate verifies that an integration action
// targeting Slack message.send with a template reference loads the notification template from
// the database, merges its defaults into the operation config, and delivers the message
// through the mock Slack API
func TestExecuteIntegrationAction_SlackMessageSendWithTemplate(t *testing.T) {
	s := suite.setupTest(t)
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()
	seedCtx := s.SeedContext(userID, orgID)

	slackIntegration, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Slack Test Integration").
		SetDefinitionID(slackdef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusConnected).
		Save(seedCtx)
	assert.NilError(t, err)
	assert.NilError(t, s.integrationsRT.ReconcileCredential(seedCtx, slackIntegration, slackdef.MockBotTokenConnection, slackdef.MockBotTokenCredential()))

	templateKey := "slack-tpl-" + ulid.Make().String()

	template, err := s.client.NotificationTemplate.Create().
		SetOwnerID(orgID).
		SetKey(templateKey).
		SetName("Slack Template: " + templateKey).
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetTemplateContext(enums.TemplateContextWorkflowAction).
		SetTopicPattern(slackdef.MessageSendOp.Name()).
		SetDefaults(map[string]any{
			"channel": "C-FROM-TEMPLATE",
			"text":    "Hello from template defaults",
		}).
		SetActive(true).
		Save(seedCtx)
	assert.NilError(t, err)

	t.Cleanup(func() {
		_, _ = s.client.NotificationTemplate.Delete().Where().Exec(seedCtx)
	})

	s.mockSlackRecorder().Reset()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-SLACKTPL-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	configBytes, err := json.Marshal(map[string]any{
		"templateId": template.ID,
	})
	assert.NilError(t, err)

	params := workflows.IntegrationActionParams{
		InstallationID: slackIntegration.ID,
		DefinitionID:   slackdef.DefinitionID.ID(),
		OperationName:  slackdef.MessageSendOp.Name(),
		Config:         configBytes,
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeIntegration.String(),
		Key:    "send_slack_template",
		Params: paramsBytes,
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	assert.NilError(t, err)

	s.WaitForEvents()

	msgs := s.mockSlackRecorder().Messages()
	assert.Assert(t, is.Len(msgs, 1))
	assert.Check(t, is.Equal("C-FROM-TEMPLATE", msgs[0].Channel))
	assert.Check(t, is.Equal("Hello from template defaults", msgs[0].Text))
}

// TestExecuteIntegrationAction_SlackTemplateMergesOverrides verifies that explicit config
// fields override template defaults when both are provided
func TestExecuteIntegrationAction_SlackTemplateMergesOverrides(t *testing.T) {
	s := suite.setupTest(t)
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()
	seedCtx := s.SeedContext(userID, orgID)

	slackIntegration, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Slack Override Integration").
		SetDefinitionID(slackdef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusConnected).
		Save(seedCtx)
	assert.NilError(t, err)
	assert.NilError(t, s.integrationsRT.ReconcileCredential(seedCtx, slackIntegration, slackdef.MockBotTokenConnection, slackdef.MockBotTokenCredential()))

	templateKey := "slack-override-" + ulid.Make().String()

	template, err := s.client.NotificationTemplate.Create().
		SetOwnerID(orgID).
		SetKey(templateKey).
		SetName("Slack Override: " + templateKey).
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetTemplateContext(enums.TemplateContextWorkflowAction).
		SetTopicPattern(slackdef.MessageSendOp.Name()).
		SetDefaults(map[string]any{
			"channel": "C-DEFAULT",
			"text":    "default text",
		}).
		SetActive(true).
		Save(seedCtx)
	assert.NilError(t, err)

	t.Cleanup(func() {
		_, _ = s.client.NotificationTemplate.Delete().Where().Exec(seedCtx)
	})

	s.mockSlackRecorder().Reset()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-SLACKOVERRIDE-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	configBytes, err := json.Marshal(map[string]any{
		"templateId": template.ID,
		"text":       "override text",
	})
	assert.NilError(t, err)

	params := workflows.IntegrationActionParams{
		InstallationID: slackIntegration.ID,
		DefinitionID:   slackdef.DefinitionID.ID(),
		OperationName:  slackdef.MessageSendOp.Name(),
		Config:         configBytes,
	}
	paramsBytes, err := json.Marshal(params)
	assert.NilError(t, err)

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeIntegration.String(),
		Key:    "send_slack_override",
		Params: paramsBytes,
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	assert.NilError(t, err)

	s.WaitForEvents()

	msgs := s.mockSlackRecorder().Messages()
	assert.Assert(t, is.Len(msgs, 1))
	assert.Check(t, is.Equal("C-DEFAULT", msgs[0].Channel), "channel should come from template defaults")
	assert.Check(t, is.Equal("override text", msgs[0].Text), "text should be overridden by config")
}
