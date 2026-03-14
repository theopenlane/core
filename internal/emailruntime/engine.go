package emailruntime

import (
	"context"
	"fmt"
	"maps"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/compose"
	"github.com/theopenlane/newman/scrubber"
	"github.com/theopenlane/riverboat/pkg/jobs"
	"github.com/theopenlane/riverboat/pkg/riverqueue"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/emailtemplate"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/notificationtemplate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/mapx"
)

// ComposeFromNotificationTemplate composes a send-ready EmailMessage from notification/email template records
func ComposeFromNotificationTemplate(ctx context.Context, client *generated.Client, request ComposeRequest) (*newman.EmailMessage, error) {
	if err := request.Template.Validate(); err != nil {
		return nil, err
	}

	if len(request.To) == 0 {
		return nil, ErrMissingRecipientAddress
	}

	if request.From == "" {
		return nil, ErrMissingSenderAddress
	}

	notificationRecord, err := loadNotificationTemplate(ctx, client, request.OwnerID, request.Template.ID, request.Template.Key, request.OwnerOnly)
	if err != nil {
		return nil, err
	}

	emailRecord, err := loadEmailTemplate(ctx, client, request.OwnerID, notificationRecord, request.OwnerOnly)
	if err != nil {
		return nil, err
	}

	// Merge defaults as the base layer; call-site data takes highest precedence.
	// Email template defaults form the base, notification template defaults override,
	// and request data overrides both.
	data := make(map[string]any, len(emailRecord.Defaults)+len(notificationRecord.Defaults)+len(request.Data))
	maps.Copy(data, emailRecord.Defaults)
	maps.Copy(data, notificationRecord.Defaults)
	maps.Copy(data, request.Data)

	if err := validateTemplateData(emailRecord.Jsonconfig, data); err != nil {
		return nil, err
	}

	rendered, err := renderTemplateEnvelope(ctx, client, notificationRecord, emailRecord, data)
	if err != nil {
		return nil, err
	}

	message := newman.NewEmailMessageWithOptions(
		newman.WithFrom(request.From),
		newman.WithTo(request.To),
		newman.WithSubject(rendered.Subject),
		newman.WithHTML(rendered.HTML),
		newman.WithText(rendered.Text),
	)

	message.SetCustomHTMLScrubber(scrubber.ScrubberFunc(renderTimeHTMLSanitize))

	if request.ReplyTo != "" {
		message.ReplyTo = request.ReplyTo
	}

	message.Tags = append(message.Tags, request.Tags...)

	if len(request.Headers) > 0 {
		if message.Headers == nil {
			message.Headers = make(map[string]string)
		}

		maps.Copy(message.Headers, request.Headers)
	}

	for _, a := range staticAttachmentsFromFiles(ctx, emailRecord.Edges.Files) {
		message.AddAttachment(a)
	}

	for _, a := range request.Attachments {
		message.AddAttachment(a)
	}

	return message, nil
}

// ComposeAndQueueFromNotificationTemplate composes an email message and inserts a queue job
func ComposeAndQueueFromNotificationTemplate(ctx context.Context, client *generated.Client, request ComposeRequest, jobClient riverqueue.JobClient) (*newman.EmailMessage, error) {
	if jobClient == nil {
		return nil, ErrJobClientRequired
	}

	message, err := ComposeFromNotificationTemplate(ctx, client, request)
	if err != nil {
		return nil, err
	}

	if _, err := jobClient.Insert(ctx, jobs.EmailArgs{Message: *message}, nil); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEmailQueueInsertFailed, err)
	}

	return message, nil
}

// loadNotificationTemplate resolves an active notification template by ID or key for email channel delivery.
// When ownerOnly is true, the lookup is scoped to owner-scoped templates only.
// When ownerOnly is false, the lookup is scoped to system-owned templates only.
func loadNotificationTemplate(ctx context.Context, client *generated.Client, ownerID string, templateID string, templateKey string, ownerOnly bool) (*generated.NotificationTemplate, error) {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	query := client.NotificationTemplate.Query().
		Where(
			notificationtemplate.ChannelEQ(enums.ChannelEmail),
			notificationtemplate.ActiveEQ(true),
		)

	if ownerOnly {
		query = query.Where(notificationtemplate.OwnerIDEQ(ownerID))
	} else {
		query = query.Where(notificationtemplate.SystemOwnedEQ(true))
	}

	if templateID != "" {
		record, err := query.Where(notificationtemplate.IDEQ(templateID)).Only(allowCtx)
		if generated.IsNotFound(err) {
			return nil, ErrNotificationTemplateNotFound
		}

		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed loading notification template by id")

			return nil, ErrNotificationTemplateNotFound
		}

		return record, nil
	}

	record, err := query.Where(notificationtemplate.KeyEQ(templateKey)).Only(allowCtx)
	if generated.IsNotFound(err) {
		if !ownerOnly {
			if fb, ok := systemFallbackTemplates[templateKey]; ok {
				return fb.toNotificationTemplate(templateKey), nil
			}
		}

		return nil, ErrNotificationTemplateNotFound
	}

	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("template_key", templateKey).Msg("failed loading notification template by key")

		return nil, ErrNotificationTemplateNotFound
	}

	return record, nil
}

// loadEmailTemplate resolves the email template referenced by a notification template.
// When ownerOnly is true, the lookup is scoped to owner-scoped templates only.
// When ownerOnly is false, the lookup is scoped to system-owned templates only.
func loadEmailTemplate(ctx context.Context, client *generated.Client, ownerID string, notificationRecord *generated.NotificationTemplate, ownerOnly bool) (*generated.EmailTemplate, error) {
	if notificationRecord.EmailTemplateID == "" {
		return &generated.EmailTemplate{
			Key:             notificationRecord.Key,
			Name:            notificationRecord.Name,
			Format:          notificationRecord.Format,
			Locale:          notificationRecord.Locale,
			SubjectTemplate: notificationRecord.SubjectTemplate,
			BodyTemplate:    notificationRecord.BodyTemplate,
			Jsonconfig:      mapx.DeepCloneMapAny(notificationRecord.Jsonconfig),
			Uischema:        mapx.DeepCloneMapAny(notificationRecord.Uischema),
			Metadata:        mapx.DeepCloneMapAny(notificationRecord.Metadata),
			Active:          notificationRecord.Active,
			SystemOwned:     notificationRecord.SystemOwned,
		}, nil
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	query := client.EmailTemplate.Query().
		Where(
			emailtemplate.IDEQ(notificationRecord.EmailTemplateID),
			emailtemplate.ActiveEQ(true),
		).
		WithFiles(func(q *generated.FileQuery) {
			q.Select(
				file.FieldProvidedFileName,
				file.FieldProvidedFileExtension,
				file.FieldDetectedMimeType,
				file.FieldFileContents,
			)
		})

	if ownerOnly {
		query = query.Where(emailtemplate.OwnerIDEQ(ownerID))
	} else {
		query = query.Where(emailtemplate.SystemOwnedEQ(true))
	}

	record, err := query.Only(allowCtx)
	if generated.IsNotFound(err) {
		return nil, ErrEmailTemplateNotFound
	}

	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("email_template_id", notificationRecord.EmailTemplateID).Msg("failed loading email template")

		return nil, ErrEmailTemplateNotFound
	}

	return record, nil
}

// loadBaseTemplateContent loads the body_template from a system-owned email template by key
func loadBaseTemplateContent(ctx context.Context, client *generated.Client, key string) (string, error) {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	record, err := client.EmailTemplate.Query().
		Where(
			emailtemplate.KeyEQ(key),
			emailtemplate.SystemOwnedEQ(true),
			emailtemplate.ActiveEQ(true),
		).
		Only(allowCtx)
	if generated.IsNotFound(err) {
		return "", ErrEmailTemplateNotFound
	}

	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("base_template_key", key).Msg("failed loading base email template")

		return "", ErrEmailTemplateNotFound
	}

	return record.BodyTemplate, nil
}

// staticAttachmentsFromFiles converts File edge records to newman attachments
func staticAttachmentsFromFiles(ctx context.Context, files []*generated.File) []*newman.Attachment {
	attachments := make([]*newman.Attachment, 0, len(files))

	for _, f := range files {
		if len(f.FileContents) == 0 {
			logx.FromContext(ctx).Debug().Str("file_id", f.ID).Msg("skipping static attachment without inline content")

			continue
		}

		filename := f.ProvidedFileName
		if f.ProvidedFileExtension != "" {
			filename = fmt.Sprintf("%s.%s", filename, f.ProvidedFileExtension)
		}

		a := newman.NewAttachment(filename, f.FileContents)
		a.ContentType = f.DetectedMimeType

		attachments = append(attachments, a)
	}

	return attachments
}

// Send composes and queues a template-driven email; opts carry per-call extras (tags, reply-to, attachments, etc.) applied to the ComposeRequest before dispatch
func Send(ctx context.Context, client *generated.Client, ownerID string, key string, recipient compose.Recipient, dataBuilder *TemplateData, opts ...SendOption) error {
	if client.Emailer == nil {
		return ErrEmailerNotConfigured
	}

	if dataBuilder == nil {
		dataBuilder = NewTemplateData()
	}

	data, err := dataBuilder.Build(*client.Emailer, recipient)
	if err != nil {
		return err
	}

	req := ComposeRequest{
		OwnerID: ownerID,
		Template: TemplateRef{
			Key: key,
		},
		To:   []string{recipient.Email},
		From: client.Emailer.FromEmail,
		Data: data,
	}

	for _, opt := range opts {
		opt(&req)
	}

	_, err = ComposeAndQueueFromNotificationTemplate(ctx, client, req, client.Job)

	return err
}

// validateTemplateData validates template input against a jsonschema object
func validateTemplateData(schema map[string]any, payload map[string]any) error {
	valid, err := ValidateJSONSchema(schema, payload)
	if err != nil {
		return ErrTemplateDataInvalid
	}

	if valid {
		return nil
	}

	return ErrTemplateDataInvalid
}
