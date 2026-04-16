//go:build test

package engine_test

import (
	"encoding/json"
	"errors"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/emailtemplate"
	"github.com/theopenlane/core/internal/ent/generated/notificationtemplate"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// mustMarshalSendEmailParams encodes SendEmailActionParams or fails the suite.
func (s *WorkflowEngineTestSuite) mustMarshalSendEmailParams(params workflows.SendEmailActionParams) []byte {
	b, err := json.Marshal(params)
	s.Require().NoError(err)

	return b
}

// clearEmailState truncates river tables and resets the mock email sender
func (s *WorkflowEngineTestSuite) clearEmailState() {
	err := s.client.Job.TruncateRiverTables(s.ctx)
	s.Require().NoError(err)

	s.mockEmailSender().Reset()
}

// createLinkedEmailTemplates creates an EmailTemplate and NotificationTemplate (email channel)
// linked together, owned by orgID. Returns the notification template key.
func (s *WorkflowEngineTestSuite) createLinkedEmailTemplates(userID, orgID, key, subject, body string) {
	seedCtx := s.SeedContext(userID, orgID)

	emailRecord, err := s.client.EmailTemplate.Create().
		SetOwnerID(orgID).
		SetKey(key).
		SetName("Email: " + key).
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetSubjectTemplate(subject).
		SetBodyTemplate("<p>" + body + "</p>").
		SetTextTemplate(body).
		SetActive(true).
		Save(seedCtx)
	s.Require().NoError(err)

	_, err = s.client.NotificationTemplate.Create().
		SetOwnerID(orgID).
		SetKey(key).
		SetName("Notification: " + key).
		SetChannel(enums.ChannelEmail).
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetTopicPattern("workflow.email").
		SetEmailTemplateID(emailRecord.ID).
		SetActive(true).
		Save(seedCtx)
	s.Require().NoError(err)
}

// cleanupEmailTemplates removes the email and notification templates created for a test key.
func (s *WorkflowEngineTestSuite) cleanupEmailTemplates(userID, orgID, key string) {
	seedCtx := s.SeedContext(userID, orgID)

	_, err := s.client.NotificationTemplate.Delete().
		Where(notificationtemplate.KeyEQ(key)).
		Exec(seedCtx)
	s.Require().NoError(err)

	_, err = s.client.EmailTemplate.Delete().
		Where(emailtemplate.KeyEQ(key)).
		Exec(seedCtx)
	s.Require().NoError(err)
}

// TestExecuteSendEmail_ByKey verifies that a send_email action composes and queues a message
// when referencing a notification template by key, and confirms the queued job carries the
// correct recipient and rendered subject.
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_ByKey() {
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	templateKey := "wf-send-email-bykey-" + ulid.Make().String()
	s.createLinkedEmailTemplates(userID, orgID, templateKey, "Hello Workflow", "Body text")
	defer s.cleanupEmailTemplates(userID, orgID, templateKey)

	s.clearEmailState()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-BYKEY-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	params := workflows.SendEmailActionParams{
		TemplateKey: templateKey,
		To:          []string{"recipient@example.com"},
		From:        "sender@example.com",
	}

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeSendEmail.String(),
		Key:    "send_test_email",
		Params: s.mustMarshalSendEmailParams(params),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().NoError(err)

	msgs := s.mockEmailSender().Messages()
	s.Require().Len(msgs, 1)
	s.Equal([]string{"recipient@example.com"}, msgs[0].To)
	s.Equal("Hello Workflow", msgs[0].Subject)
	s.Equal("sender@example.com", msgs[0].From)
}

// TestExecuteSendEmail_ByID verifies that a send_email action resolves a notification template
// by ID and queues a correctly addressed message.
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_ByID() {
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	templateKey := "wf-send-email-byid-" + ulid.Make().String()
	s.createLinkedEmailTemplates(userID, orgID, templateKey, "Resolved by ID", "Content by ID")
	defer s.cleanupEmailTemplates(userID, orgID, templateKey)

	s.clearEmailState()

	seedCtx := s.SeedContext(userID, orgID)
	notifRecord, err := s.client.NotificationTemplate.Query().
		Where(notificationtemplate.KeyEQ(templateKey)).
		Only(seedCtx)
	s.Require().NoError(err)

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-BYID-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	params := workflows.SendEmailActionParams{
		TemplateID: notifRecord.ID,
		To:         []string{"byid@example.com"},
		From:       "sender@example.com",
	}

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeSendEmail.String(),
		Key:    "send_by_id",
		Params: s.mustMarshalSendEmailParams(params),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().NoError(err)

	msgs := s.mockEmailSender().Messages()
	s.Require().Len(msgs, 1)
	s.Equal([]string{"byid@example.com"}, msgs[0].To)
	s.Equal("Resolved by ID", msgs[0].Subject)
}

// TestExecuteSendEmail_WithTargetUser verifies that recipients are resolved from Targets when
// To is omitted. The resolved user email is asserted in the queued job.
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_WithTargetUser() {
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	templateKey := "wf-send-email-target-" + ulid.Make().String()
	s.createLinkedEmailTemplates(userID, orgID, templateKey, "Target User Email", "Body for target")
	defer s.cleanupEmailTemplates(userID, orgID, templateKey)

	s.clearEmailState()

	// Load the test user's email to assert it appears in the queued job
	seedCtx := s.SeedContext(userID, orgID)
	testUser, err := s.client.User.Get(seedCtx, userID)
	s.Require().NoError(err)

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-TARGET-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	params := workflows.SendEmailActionParams{
		TemplateKey: templateKey,
		From:        "sender@example.com",
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{Type: enums.WorkflowTargetTypeUser, ID: userID},
			},
		},
	}

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeSendEmail.String(),
		Key:    "send_to_user",
		Params: s.mustMarshalSendEmailParams(params),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().NoError(err)

	msgs := s.mockEmailSender().Messages()
	s.Require().Len(msgs, 1)
	s.Equal([]string{testUser.Email}, msgs[0].To)
	s.Equal("Target User Email", msgs[0].Subject)
}

// TestExecuteSendEmail_WorkflowDefinitionWithSendEmail verifies end-to-end that a workflow
// definition containing a send_email action processes successfully via ProcessAction and
// inserts a correctly composed job into the queue.
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_WorkflowDefinitionWithSendEmail() {
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	templateKey := "wf-send-email-e2e-" + ulid.Make().String()
	s.createLinkedEmailTemplates(userID, orgID, templateKey, "E2E Subject", "E2E body text")
	defer s.cleanupEmailTemplates(userID, orgID, templateKey)

	s.clearEmailState()

	sendEmailAction := models.WorkflowAction{
		Key:  "notify_via_email",
		Type: enums.WorkflowActionTypeSendEmail.String(),
		Params: s.mustMarshalSendEmailParams(workflows.SendEmailActionParams{
			TemplateKey: templateKey,
			To:          []string{"e2e@example.com"},
			From:        "noreply@example.com",
		}),
	}

	doc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{Operation: "UPDATE", Fields: []string{"status"}},
		},
		Conditions: []models.WorkflowCondition{
			{Expression: "true"},
		},
		Actions: []models.WorkflowAction{sendEmailAction},
	}

	operations, fields := workflows.DeriveTriggerPrefilter(doc)

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Email Workflow " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindNotification).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetDefinitionJSON(doc).
		SetTriggerOperations(operations).
		SetTriggerFields(fields).
		Save(userCtx)
	s.Require().NoError(err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-E2E-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	err = wfEngine.ProcessAction(userCtx, instance, sendEmailAction)
	s.Require().NoError(err)

	msgs := s.mockEmailSender().Messages()
	s.Require().Len(msgs, 1)
	s.Equal([]string{"e2e@example.com"}, msgs[0].To)
	s.Equal("E2E Subject", msgs[0].Subject)
	s.Equal("noreply@example.com", msgs[0].From)
	s.Contains(msgs[0].Text, "E2E body text")
}

// TestExecuteSendEmail_NoTemplateReference verifies that a send_email action with no template
// reference returns ErrSendEmailTemplateRequired.
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_NoTemplateReference() {
	_, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-NOREF-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	action := models.WorkflowAction{
		Type: enums.WorkflowActionTypeSendEmail.String(),
		Key:  "no_template_ref",
		Params: s.mustMarshalSendEmailParams(workflows.SendEmailActionParams{
			To:   []string{"recipient@example.com"},
			From: "sender@example.com",
		}),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, engine.ErrSendEmailTemplateRequired))
}

// TestExecuteSendEmail_BothTemplateAndKeyConflict verifies that providing both template_id and
// template_key returns ErrSendEmailTemplateReferenceConflict.
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_BothTemplateAndKeyConflict() {
	_, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-CONFLICT-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	action := models.WorkflowAction{
		Type: enums.WorkflowActionTypeSendEmail.String(),
		Key:  "template_conflict",
		Params: s.mustMarshalSendEmailParams(workflows.SendEmailActionParams{
			TemplateID:  "some-id",
			TemplateKey: "some-key",
			To:          []string{"recipient@example.com"},
			From:        "sender@example.com",
		}),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, engine.ErrSendEmailTemplateReferenceConflict))
}

// TestExecuteSendEmail_NoRecipients verifies that a send_email action with a valid template
// but no resolved recipients returns ErrSendEmailNoRecipients.
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_NoRecipients() {
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	templateKey := "wf-send-email-norecip-" + ulid.Make().String()
	s.createLinkedEmailTemplates(userID, orgID, templateKey, "No Recip Subject", "body")
	defer s.cleanupEmailTemplates(userID, orgID, templateKey)

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-NORECIP-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	// No To or Targets
	action := models.WorkflowAction{
		Type: enums.WorkflowActionTypeSendEmail.String(),
		Key:  "no_recipients",
		Params: s.mustMarshalSendEmailParams(workflows.SendEmailActionParams{
			TemplateKey: templateKey,
			From:        "sender@example.com",
		}),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, engine.ErrSendEmailNoRecipients))
}

// TestExecuteSendEmail_TemplateNotFound verifies that a send_email action with a non-existent
// template key returns an error wrapping ErrSendEmailTemplateComposeFailed.
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_TemplateNotFound() {
	_, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-MISSING-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	action := models.WorkflowAction{
		Type: enums.WorkflowActionTypeSendEmail.String(),
		Key:  "missing_template",
		Params: s.mustMarshalSendEmailParams(workflows.SendEmailActionParams{
			TemplateKey: "nonexistent_key_" + ulid.Make().String(),
			To:          []string{"recipient@example.com"},
			From:        "sender@example.com",
		}),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, engine.ErrSendEmailTemplateComposeFailed))
}

// TestExecuteSendEmail_SenderMissing verifies that a send_email action with no From address
// and no default emailer from address returns ErrSendEmailSenderMissing.
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_SenderMissing() {
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	templateKey := "wf-send-email-nosender-" + ulid.Make().String()
	s.createLinkedEmailTemplates(userID, orgID, templateKey, "Subject", "body")
	defer s.cleanupEmailTemplates(userID, orgID, templateKey)

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-NOSEND-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	// No From; the test suite emailer has an empty FromEmail.
	action := models.WorkflowAction{
		Type: enums.WorkflowActionTypeSendEmail.String(),
		Key:  "no_sender",
		Params: s.mustMarshalSendEmailParams(workflows.SendEmailActionParams{
			TemplateKey: templateKey,
			To:          []string{"recipient@example.com"},
		}),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, engine.ErrSendEmailSenderMissing))
}

// TestExecuteSendEmail_FullAsyncPath verifies the complete end-to-end path:
// Control mutation → Gala mutation event → TriggerWorkflow → TopicWorkflowTriggered →
// ProcessAction → executeSendEmail → River job inserted with correct email payload.
// This exercises every layer of the workflow event dispatch stack, not just Execute directly.
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_FullAsyncPath() {
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	templateKey := "wf-send-email-async-" + ulid.Make().String()
	s.createLinkedEmailTemplates(userID, orgID, templateKey, "Async Subject", "Async body text")
	defer s.cleanupEmailTemplates(userID, orgID, templateKey)

	s.clearEmailState()

	sendEmailAction := models.WorkflowAction{
		Key:  "send_email_async",
		Type: enums.WorkflowActionTypeSendEmail.String(),
		Params: s.mustMarshalSendEmailParams(workflows.SendEmailActionParams{
			TemplateKey: templateKey,
			To:          []string{"async@example.com"},
			From:        "noreply@example.com",
		}),
	}

	// Trigger on any Control UPDATE (empty Fields = match all).
	doc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{Operation: "UPDATE", Fields: []string{}},
		},
		Conditions: []models.WorkflowCondition{
			{Expression: "true"},
		},
		Actions: []models.WorkflowAction{sendEmailAction},
	}

	operations, triggerFields := workflows.DeriveTriggerPrefilter(doc)

	def, err := s.client.WorkflowDefinition.Create().
		SetName("Async Email Workflow " + ulid.Make().String()).
		SetWorkflowKind(enums.WorkflowKindNotification).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(orgID).
		SetDefinitionJSON(doc).
		SetTriggerOperations(operations).
		SetTriggerFields(triggerFields).
		Save(userCtx)
	s.Require().NoError(err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-ASYNC-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetReferenceID("ref-before-" + ulid.Make().String()).
		Save(seedCtx)
	s.Require().NoError(err)

	// Mutate the control — this fires the Gala mutation hook which dispatches the workflow.
	_, err = s.client.Control.UpdateOneID(control.ID).
		SetReferenceID("ref-after-" + ulid.Make().String()).
		Save(seedCtx)
	s.Require().NoError(err)

	// Block until gala workers finish: mutation event → TriggerWorkflow → ProcessAction.
	s.WaitForEvents()

	// Verify a workflow instance was created and completed for the control.
	instance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(def.ID),
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	s.Require().NoError(err)
	s.Require().NotNil(instance)
	s.Equal(enums.WorkflowInstanceStateCompleted, instance.State)

	// Verify the email was sent with the correct composed message.
	msgs := s.mockEmailSender().Messages()
	s.Require().Len(msgs, 1)
	s.Equal([]string{"async@example.com"}, msgs[0].To)
	s.Equal("Async Subject", msgs[0].Subject)
	s.Equal("noreply@example.com", msgs[0].From)
	s.Contains(msgs[0].Text, "Async body text")
}

// TestExecuteSendEmail_OwnerOnlyExcludesSystemTemplate verifies that org-owned workflow definitions
// cannot reference system-owned notification templates; only owner-scoped templates are accessible
// from org-owned workflow definitions.
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_OwnerOnlyExcludesSystemTemplate() {
	_, orgID, userCtx := s.SetupTestUser()
	_, _, adminCtx := s.SetupSystemAdmin()
	wfEngine := s.Engine()

	// Create a system-owned email notification template (no owner_id, system_owned=true via hook)
	systemTemplateKey := "wf-send-email-system-" + ulid.Make().String()

	_, err := s.client.NotificationTemplate.Create().
		SetKey(systemTemplateKey).
		SetName("System Email Template").
		SetChannel(enums.ChannelEmail).
		SetActive(true).
		SetTopicPattern("workflow.email").
		SetSubjectTemplate("System notification").
		SetBodyTemplate("<p>System email.</p>").
		Save(adminCtx)
	s.Require().NoError(err)

	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-SYSONLY-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	instance := s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	// The system template has no owner_id; OwnerOnly enforcement must exclude it.
	action := models.WorkflowAction{
		Type: enums.WorkflowActionTypeSendEmail.String(),
		Key:  "send_system_template",
		Params: s.mustMarshalSendEmailParams(workflows.SendEmailActionParams{
			TemplateKey: systemTemplateKey,
			To:          []string{"recipient@example.com"},
			From:        "sender@example.com",
		}),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, engine.ErrSendEmailTemplateComposeFailed))
}
