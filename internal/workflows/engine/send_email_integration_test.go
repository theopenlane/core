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
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// mustMarshalSendEmailParams encodes SendEmailActionParams or fails the suite
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

// createTestEmailTemplate creates an EmailTemplate owned by orgID and returns its ID
func (s *WorkflowEngineTestSuite) createTestEmailTemplate(userID, orgID, key, subject, body string) string {
	seedCtx := s.SeedContext(userID, orgID)

	emailRecord, err := s.client.EmailTemplate.Create().
		SetOwnerID(orgID).
		SetKey(emaildef.BrandedMessageOp.Name()).
		SetName("Email: " + key).
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetTemplateContext(enums.TemplateContextWorkflowAction).
		SetDefaults(map[string]any{
			"subject": subject,
			"title":   subject,
			"intros":  []any{body},
		}).
		SetActive(true).
		Save(seedCtx)
	s.Require().NoError(err)

	return emailRecord.ID
}

// cleanupTestEmailTemplate removes the email template created for a test key
func (s *WorkflowEngineTestSuite) cleanupTestEmailTemplate(userID, orgID, key string) {
	seedCtx := s.SeedContext(userID, orgID)

	_, err := s.client.EmailTemplate.Delete().
		Where(emailtemplate.NameEQ("Email: " + key)).
		Exec(seedCtx)
	s.Require().NoError(err)
}

// TestExecuteSendEmail_ByKey verifies that a send_email action composes and queues a message
// when referencing an email template by key
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_ByKey() {
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	templateKey := "wf-send-email-bykey-" + ulid.Make().String()
	s.createTestEmailTemplate(userID, orgID, templateKey, "Hello Workflow", "Body text")
	defer s.cleanupTestEmailTemplate(userID, orgID, templateKey)

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
		EmailTemplateKey: emaildef.BrandedMessageOp.Name(),
		To:               []string{"recipient@example.com"},
		From:             "sender@example.com",
	}

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeSendEmail.String(),
		Key:    "send_test_email",
		Params: s.mustMarshalSendEmailParams(params),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().NoError(err)
	s.WaitForEvents()

	msgs := s.mockEmailSender().Messages()
	s.Require().Len(msgs, 1)
	s.Equal([]string{"recipient@example.com"}, msgs[0].To)
	s.Equal("Hello Workflow", msgs[0].Subject)
	s.Equal("sender@example.com", msgs[0].From)
}

// TestExecuteSendEmail_ByID verifies that a send_email action resolves an email template
// by ID and queues a correctly addressed message
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_ByID() {
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	templateKey := "wf-send-email-byid-" + ulid.Make().String()
	emailTemplateID := s.createTestEmailTemplate(userID, orgID, templateKey, "Resolved by ID", "Content by ID")
	defer s.cleanupTestEmailTemplate(userID, orgID, templateKey)

	s.clearEmailState()

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
		EmailTemplateID: emailTemplateID,
		To:              []string{"byid@example.com"},
		From:            "sender@example.com",
	}

	action := models.WorkflowAction{
		Type:   enums.WorkflowActionTypeSendEmail.String(),
		Key:    "send_by_id",
		Params: s.mustMarshalSendEmailParams(params),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().NoError(err)
	s.WaitForEvents()

	msgs := s.mockEmailSender().Messages()
	s.Require().Len(msgs, 1)
	s.Equal([]string{"byid@example.com"}, msgs[0].To)
	s.Equal("Resolved by ID", msgs[0].Subject)
}

// TestExecuteSendEmail_WithTargetUser verifies that recipients are resolved from Targets when
// To is omitted
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_WithTargetUser() {
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	templateKey := "wf-send-email-target-" + ulid.Make().String()
	s.createTestEmailTemplate(userID, orgID, templateKey, "Target User Email", "Body for target")
	defer s.cleanupTestEmailTemplate(userID, orgID, templateKey)

	s.clearEmailState()

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
		EmailTemplateKey: emaildef.BrandedMessageOp.Name(),
		From:             "sender@example.com",
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
	s.WaitForEvents()

	msgs := s.mockEmailSender().Messages()
	s.Require().Len(msgs, 1)
	s.Equal([]string{testUser.Email}, msgs[0].To)
	s.Equal("Target User Email", msgs[0].Subject)
}

// TestExecuteSendEmail_WorkflowDefinitionWithSendEmail verifies end-to-end that a workflow
// definition containing a send_email action processes successfully after trigger emission
// and sends a correctly composed email through the runtime integration
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_WorkflowDefinitionWithSendEmail() {
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	templateKey := "wf-send-email-e2e-" + ulid.Make().String()
	s.createTestEmailTemplate(userID, orgID, templateKey, "E2E Subject", "E2E body text")
	defer s.cleanupTestEmailTemplate(userID, orgID, templateKey)

	s.clearEmailState()

	sendEmailAction := models.WorkflowAction{
		Key:  "notify_via_email",
		Type: enums.WorkflowActionTypeSendEmail.String(),
		Params: s.mustMarshalSendEmailParams(workflows.SendEmailActionParams{
			EmailTemplateKey: emaildef.BrandedMessageOp.Name(),
			To:               []string{"e2e@example.com"},
			From:             "noreply@example.com",
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
	s.TriggerInstance(userCtx, wfEngine, def, obj, engine.TriggerInput{
		EventType:     "UPDATE",
		ChangedFields: []string{"status"},
	})

	s.WaitForEvents()

	msgs := s.mockEmailSender().Messages()
	s.Require().Len(msgs, 1)
	s.Equal([]string{"e2e@example.com"}, msgs[0].To)
	s.Equal("E2E Subject", msgs[0].Subject)
	s.Equal("noreply@example.com", msgs[0].From)
	s.Contains(msgs[0].Text, "E2E body text")
}

// TestExecuteSendEmail_NoTemplateReference verifies that a send_email action with no template
// reference returns ErrSendEmailTemplateRequired
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

// TestExecuteSendEmail_BothTemplateAndKeyConflict verifies that providing both emailTemplateId
// and emailTemplateKey returns ErrSendEmailTemplateReferenceConflict
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
			EmailTemplateID:  "some-id",
			EmailTemplateKey: "some-key",
			To:               []string{"recipient@example.com"},
			From:             "sender@example.com",
		}),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, engine.ErrSendEmailTemplateReferenceConflict))
}

// TestExecuteSendEmail_NoRecipients verifies that a send_email action with a valid template
// but no resolved recipients returns ErrSendEmailNoRecipients
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_NoRecipients() {
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	templateKey := "wf-send-email-norecip-" + ulid.Make().String()
	s.createTestEmailTemplate(userID, orgID, templateKey, "No Recip Subject", "body")
	defer s.cleanupTestEmailTemplate(userID, orgID, templateKey)

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

	action := models.WorkflowAction{
		Type: enums.WorkflowActionTypeSendEmail.String(),
		Key:  "no_recipients",
		Params: s.mustMarshalSendEmailParams(workflows.SendEmailActionParams{
			EmailTemplateKey: emaildef.BrandedMessageOp.Name(),
			From:             "sender@example.com",
		}),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, engine.ErrSendEmailNoRecipients))
}

// TestExecuteSendEmail_TemplateNotFound verifies that a send_email action with a non-existent
// template key returns ErrSendEmailTemplateNotFound
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
			EmailTemplateKey: "nonexistent_key_" + ulid.Make().String(),
			To:               []string{"recipient@example.com"},
			From:             "sender@example.com",
		}),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, engine.ErrSendEmailTemplateNotFound))
}

// TestExecuteSendEmail_DefaultSender verifies that a send_email action without a
// From override uses the email integration's configured sender
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_DefaultSender() {
	userID, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	templateKey := "wf-send-email-nosender-" + ulid.Make().String()
	s.createTestEmailTemplate(userID, orgID, templateKey, "Subject", "body")
	defer s.cleanupTestEmailTemplate(userID, orgID, templateKey)

	s.clearEmailState()

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

	action := models.WorkflowAction{
		Type: enums.WorkflowActionTypeSendEmail.String(),
		Key:  "default_sender",
		Params: s.mustMarshalSendEmailParams(workflows.SendEmailActionParams{
			EmailTemplateKey: emaildef.BrandedMessageOp.Name(),
			To:               []string{"recipient@example.com"},
		}),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().NoError(err)
	s.WaitForEvents()

	msgs := s.mockEmailSender().Messages()
	s.Require().Len(msgs, 1)
	s.Equal("test@example.com", msgs[0].From)
}

// TestExecuteSendEmail_FullAsyncPath verifies the complete end-to-end path:
// Control mutation -> Gala mutation event -> TriggerWorkflow -> TopicWorkflowTriggered ->
// ProcessAction -> executeSendEmail -> River job inserted with correct email payload
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_FullAsyncPath() {
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	templateKey := "wf-send-email-async-" + ulid.Make().String()
	s.createTestEmailTemplate(userID, orgID, templateKey, "Async Subject", "Async body text")
	defer s.cleanupTestEmailTemplate(userID, orgID, templateKey)

	s.clearEmailState()

	sendEmailAction := models.WorkflowAction{
		Key:  "send_email_async",
		Type: enums.WorkflowActionTypeSendEmail.String(),
		Params: s.mustMarshalSendEmailParams(workflows.SendEmailActionParams{
			EmailTemplateKey: emaildef.BrandedMessageOp.Name(),
			To:               []string{"async@example.com"},
			From:             "noreply@example.com",
		}),
	}

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

	_, err = s.client.Control.UpdateOneID(control.ID).
		SetReferenceID("ref-after-" + ulid.Make().String()).
		Save(seedCtx)
	s.Require().NoError(err)

	s.WaitForEvents()
	s.WaitForEvents()

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

	msgs := s.mockEmailSender().Messages()
	s.Require().Len(msgs, 1)
	s.Equal([]string{"async@example.com"}, msgs[0].To)
	s.Equal("Async Subject", msgs[0].Subject)
	s.Equal("noreply@example.com", msgs[0].From)
	s.Contains(msgs[0].Text, "Async body text")
}

// TestExecuteSendEmail_OwnerOnlyExcludesSystemTemplate verifies that org-owned workflow definitions
// cannot reference system-owned email templates
func (s *WorkflowEngineTestSuite) TestExecuteSendEmail_OwnerOnlyExcludesSystemTemplate() {
	_, orgID, userCtx := s.SetupTestUser()
	_, _, adminCtx := s.SetupSystemAdmin()
	wfEngine := s.Engine()

	systemEmailTemplate, err := s.client.EmailTemplate.Create().
		SetKey(emaildef.BrandedMessageOp.Name()).
		SetName("System Email Template").
		SetLocale("en-US").
		SetFormat(enums.NotificationTemplateFormatHTML).
		SetTemplateContext(enums.TemplateContextWorkflowAction).
		SetDefaults(map[string]any{
			"subject": "System notification",
			"title":   "System notification",
			"intros":  []any{"System email."},
		}).
		SetActive(true).
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

	action := models.WorkflowAction{
		Type: enums.WorkflowActionTypeSendEmail.String(),
		Key:  "send_system_template",
		Params: s.mustMarshalSendEmailParams(workflows.SendEmailActionParams{
			EmailTemplateID: systemEmailTemplate.ID,
			To:              []string{"recipient@example.com"},
			From:            "sender@example.com",
		}),
	}

	err = wfEngine.Execute(userCtx, action, instance, obj)
	s.Require().Error(err)
	s.Require().True(errors.Is(err, engine.ErrSendEmailTemplateNotFound))
}
