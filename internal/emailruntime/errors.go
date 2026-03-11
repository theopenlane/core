package emailruntime

import "errors"

var (
	// ErrMissingTemplateReference is returned when no notification template identifier is provided.
	ErrMissingTemplateReference = errors.New("missing notification template reference")
	// ErrTemplateReferenceConflict is returned when both template ID and key are provided.
	ErrTemplateReferenceConflict = errors.New("template reference conflict")
	// ErrMissingRecipientAddress is returned when no recipient email address is provided.
	ErrMissingRecipientAddress = errors.New("missing recipient email address")
	// ErrMissingSenderAddress is returned when no sender email address is provided.
	ErrMissingSenderAddress = errors.New("missing sender email address")
	// ErrNotificationTemplateNotFound is returned when an email notification template cannot be found.
	ErrNotificationTemplateNotFound = errors.New("notification template not found")
	// ErrEmailTemplateNotFound is returned when an email template cannot be found.
	ErrEmailTemplateNotFound = errors.New("email template not found")
	// ErrTemplateDataInvalid is returned when provided data does not satisfy template jsonschema.
	ErrTemplateDataInvalid = errors.New("template data invalid")
	// ErrTemplateRenderFailed is returned when a template cannot be rendered.
	ErrTemplateRenderFailed = errors.New("template render failed")
	// ErrTemplateSeedFailed is returned when default system templates cannot be seeded.
	ErrTemplateSeedFailed = errors.New("template seed failed")
	// ErrJobClientRequired is returned when queueing is requested without a job client.
	ErrJobClientRequired = errors.New("job client required")
	// ErrEmailQueueInsertFailed is returned when enqueueing a composed email job fails.
	ErrEmailQueueInsertFailed = errors.New("email queue insert failed")
	// ErrEmailerNotConfigured is returned when the ent client has no emailer config attached.
	ErrEmailerNotConfigured = errors.New("emailer not configured on client")
	// ErrUnsupportedTemplateURLKey is returned when a template URL key is not recognized.
	ErrUnsupportedTemplateURLKey = errors.New("unsupported template URL key")
)
