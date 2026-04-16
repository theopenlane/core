package email

import (
	"context"
	"fmt"

	"github.com/theopenlane/newman"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/emailtemplate"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// loadEmailTemplate resolves an active email template by ID for the given owner
func loadEmailTemplate(ctx context.Context, client *generated.Client, ownerID string, emailTemplateID string) (*generated.EmailTemplate, error) {
	record, err := client.EmailTemplate.Query().
		Where(
			emailtemplate.IDEQ(emailTemplateID),
			emailtemplate.ActiveEQ(true),
			emailtemplate.OwnerIDEQ(ownerID),
		).
		WithEmailBranding().
		WithFiles(func(q *generated.FileQuery) {
			q.Select(
				file.FieldProvidedFileName,
				file.FieldProvidedFileExtension,
				file.FieldDetectedMimeType,
				file.FieldFileContents)
		}).Only(ctx)
	if generated.IsNotFound(err) {
		return nil, ErrEmailTemplateNotFound
	}

	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("email_template_id", emailTemplateID).Msg("failed loading email template")

		return nil, ErrEmailTemplateNotFound
	}

	return record, nil
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

// validateTemplateData validates template input against a jsonschema object
func validateTemplateData(schema map[string]any, payload map[string]any) error {
	if len(schema) == 0 {
		return nil
	}

	result, err := jsonx.ValidateSchema(schema, payload)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTemplateDataInvalid, err)
	}

	if result.Valid() {
		return nil
	}

	return fmt.Errorf("%w: %s", ErrTemplateDataInvalid, result.Errors())
}
