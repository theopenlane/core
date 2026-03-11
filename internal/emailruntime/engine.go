package emailruntime

import (
	"context"
	"fmt"
	"maps"

	"github.com/samber/lo"
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
)

// Service composes and seeds notification-driven email templates.
type Service struct {
	// client is the ent client used to load and mutate templates.
	client *generated.Client
}

// NewService creates an email template runtime service.
func NewService(client *generated.Client) *Service {
	return &Service{client: client}
}

// ComposeFromNotificationTemplate composes a send-ready EmailMessage from notification/email template records.
func (s *Service) ComposeFromNotificationTemplate(ctx context.Context, request ComposeRequest) (*newman.EmailMessage, error) {
	request.Template = request.Template.Normalize()
	if err := request.Template.Validate(); err != nil {
		return nil, err
	}

	if len(request.To) == 0 {
		return nil, ErrMissingRecipientAddress
	}

	if request.From == "" {
		return nil, ErrMissingSenderAddress
	}

	notificationRecord, err := s.loadNotificationTemplate(ctx, request.OwnerID, request.Template.ID, request.Template.Key, request.OwnerOnly)
	if err != nil {
		return nil, err
	}

	emailRecord, err := s.loadEmailTemplate(ctx, request.OwnerID, notificationRecord, request.OwnerOnly)
	if err != nil {
		return nil, err
	}

	rendered, err := s.renderTemplateEnvelope(ctx, notificationRecord, emailRecord, request.Data)
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

	if len(request.Tags) > 0 {
		message.Tags = append(message.Tags, request.Tags...)
	}

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

// ComposeAndQueueFromNotificationTemplate composes an email message and inserts a queue job.
func (s *Service) ComposeAndQueueFromNotificationTemplate(ctx context.Context, request ComposeRequest, jobClient riverqueue.JobClient) (*newman.EmailMessage, error) {
	if jobClient == nil {
		return nil, ErrJobClientRequired
	}

	message, err := s.ComposeFromNotificationTemplate(ctx, request)
	if err != nil {
		return nil, err
	}

	if _, err := jobClient.Insert(ctx, jobs.EmailArgs{Message: *message}, nil); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEmailQueueInsertFailed, err)
	}

	return message, nil
}

// loadNotificationTemplate resolves an active notification template by ID or key for email channel delivery.
// When ownerOnly is true, only owner-scoped templates are returned; system-owned templates are excluded.
func (s *Service) loadNotificationTemplate(ctx context.Context, ownerID string, templateID string, templateKey TemplateKey, ownerOnly bool) (*generated.NotificationTemplate, error) {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	query := s.client.NotificationTemplate.Query().
		Where(
			notificationtemplate.ChannelEQ(enums.ChannelEmail),
			notificationtemplate.ActiveEQ(true),
		)

	switch {
	case ownerOnly:
		query = query.Where(notificationtemplate.OwnerIDEQ(ownerID))
	case ownerID != "":
		query = query.Where(
			notificationtemplate.Or(
				notificationtemplate.OwnerIDEQ(ownerID),
				notificationtemplate.SystemOwnedEQ(true),
			),
		)
	default:
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

	records, err := query.Where(notificationtemplate.KeyEQ(templateKey.String())).All(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("template_key", templateKey.String()).Msg("failed loading notification templates by key")
		return nil, ErrNotificationTemplateNotFound
	}

	if len(records) == 0 {
		return nil, ErrNotificationTemplateNotFound
	}

	if ownerID == "" {
		return records[0], nil
	}

	if matched, ok := lo.Find(records, func(item *generated.NotificationTemplate) bool {
		return item.OwnerID == ownerID
	}); ok {
		return matched, nil
	}

	return records[0], nil
}

// loadEmailTemplate resolves the email template referenced by a notification template.
// When ownerOnly is true, only owner-scoped email templates are returned; system-owned templates are excluded.
func (s *Service) loadEmailTemplate(ctx context.Context, ownerID string, notificationRecord *generated.NotificationTemplate, ownerOnly bool) (*generated.EmailTemplate, error) {
	if notificationRecord.EmailTemplateID == "" {
		// Synthesize an EmailTemplate from the notification record fields when no dedicated email template is linked.
		// New fields added to generated.EmailTemplate will be zero-valued here.
		return &generated.EmailTemplate{
			Key:             notificationRecord.Key,
			Name:            notificationRecord.Name,
			Format:          notificationRecord.Format,
			Locale:          notificationRecord.Locale,
			SubjectTemplate: notificationRecord.SubjectTemplate,
			BodyTemplate:    notificationRecord.BodyTemplate,
			Jsonconfig:      maps.Clone(notificationRecord.Jsonconfig),
			Uischema:        maps.Clone(notificationRecord.Uischema),
			Metadata:        maps.Clone(notificationRecord.Metadata),
			Active:          notificationRecord.Active,
			SystemOwned:     notificationRecord.SystemOwned,
		}, nil
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	query := s.client.EmailTemplate.Query().
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

	switch {
	case ownerOnly:
		query = query.Where(emailtemplate.OwnerIDEQ(ownerID))
	case ownerID != "":
		query = query.Where(
			emailtemplate.Or(
				emailtemplate.OwnerIDEQ(ownerID),
				emailtemplate.SystemOwnedEQ(true),
			),
		)
	default:
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

// loadBaseTemplateContent loads the body_template from a system-owned email template by key.
// It is used to fetch the base template shell for RAW_HTML assembly at render time.
func (s *Service) loadBaseTemplateContent(ctx context.Context, key string) (string, error) {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	record, err := s.client.EmailTemplate.Query().
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

// staticAttachmentsFromFiles converts File edge records to newman attachments.
// Only files with inline FileContents are included; files backed by external storage are skipped.
func staticAttachmentsFromFiles(ctx context.Context, files []*generated.File) []*newman.Attachment {
	attachments := make([]*newman.Attachment, 0, len(files))

	for _, f := range files {
		if len(f.FileContents) == 0 {
			logx.FromContext(ctx).Warn().Str("file_id", f.ID).Msg("skipping static attachment without inline content")

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

// Send composes and queues a notification-template-driven email using the provided ent client.
// The emailer config is read from client.Emailer, which must be non-nil.
// ownerID scopes the template lookup to the owning organization; pass empty string for system-scoped lookup.
// key identifies the NotificationTemplate record. recipient provides the To address and Recipient template variables.
// dataBuilder defines typed template data overrides; pass nil to use only the base config and recipient data.
// opts carry per-call extras (tags, reply-to, attachments, etc.) applied to the ComposeRequest before dispatch.
func Send(ctx context.Context, client *generated.Client, ownerID string, key TemplateKey, recipient compose.Recipient, dataBuilder *TemplateData, opts ...SendOption) error {
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

	_, err = NewService(client).ComposeAndQueueFromNotificationTemplate(ctx, req, client.Job)

	return err
}

// validateTemplateData validates template input against a jsonschema object.
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
