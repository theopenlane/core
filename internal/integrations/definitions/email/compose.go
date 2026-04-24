package email

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/theopenlane/newman"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/emailtemplate"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// loadEmailTemplate resolves an active email template by ID for the given owner, eager-loading
// the Files edge so static attachments can be included in the dispatched message
func loadEmailTemplate(ctx context.Context, client *generated.Client, ownerID string, emailTemplateID string) (*generated.EmailTemplate, error) {
	record, err := client.EmailTemplate.Query().
		Where(
			emailtemplate.IDEQ(emailTemplateID),
			emailtemplate.ActiveEQ(true),
			emailtemplate.OwnerIDEQ(ownerID),
		).
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

// buildDispatchPayload overlays the supplied struct values onto template defaults as a JSON object
// and returns the raw payload consumed by EmailDispatcher.SendByKey. Each overlay is marshaled
// through its JSON tags, so the overlay struct types (RecipientInfo, CampaignContext, etc.) remain
// the single source of truth for per-invocation field names; overlays apply in order, so later
// overlays win on key conflicts
func buildDispatchPayload(defaults map[string]any, overlays ...any) (json.RawMessage, error) {
	base, err := jsonx.ToRawMessage(defaults)
	if err != nil {
		return nil, fmt.Errorf("%w: defaults: %w", ErrTemplateRenderFailed, err)
	}

	if len(base) == 0 {
		base = json.RawMessage(`{}`)
	}

	for _, overlay := range overlays {
		patch, err := jsonx.ToRawMap(overlay)
		if err != nil {
			return nil, fmt.Errorf("%w: overlay: %w", ErrTemplateRenderFailed, err)
		}

		base, _, err = jsonx.MergeObjectMap(base, patch)
		if err != nil {
			return nil, fmt.Errorf("%w: merge: %w", ErrTemplateRenderFailed, err)
		}
	}

	return base, nil
}

// markCampaignTargetSent records the current time as the sent_at timestamp on a campaign target
func markCampaignTargetSent(ctx context.Context, db *generated.Client, targetID string) error {
	now := models.DateTime(time.Now())
	if err := db.CampaignTarget.UpdateOneID(targetID).SetSentAt(now).Exec(ctx); err != nil {
		return fmt.Errorf("mark sent: %w", err)
	}

	return nil
}

// createAssessmentResponse creates an assessment response record linking a
// campaign target to the campaign's assessment for completion tracking
func createAssessmentResponse(ctx context.Context, db *generated.Client, camp *generated.Campaign, assessmentID string, target *generated.CampaignTarget) error {
	create := db.AssessmentResponse.Create().
		SetAssessmentID(assessmentID).
		SetCampaignID(camp.ID).
		SetEmail(target.Email).
		SetOwnerID(camp.OwnerID)

	if camp.EntityID != "" {
		create.SetEntityID(camp.EntityID)
	}

	if camp.DueDate != nil && !camp.DueDate.IsZero() {
		create.SetDueDate(time.Time(*camp.DueDate))
	}

	return create.Exec(ctx)
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
