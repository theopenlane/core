package email

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/theopenlane/newman"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/emailtemplate"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// CampaignDispatchResult is the structured result returned by campaign operations
type CampaignDispatchResult struct {
	// SentCount is the number of messages successfully sent
	SentCount int `json:"sentCount"`
	// SkippedCount is the number of targets skipped (already sent, completed, filtered)
	SkippedCount int `json:"skippedCount"`
	// FailedCount is the number of targets that failed to render or send
	FailedCount int `json:"failedCount"`
}

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
	if err := db.CampaignTarget.UpdateOneID(targetID).
		SetSentAt(now).
		SetStatus(enums.AssessmentResponseStatusSent).
		Exec(privacy.DecisionContext(ctx, privacy.Allow)); err != nil {
		return fmt.Errorf("mark sent: %w", err)
	}

	return nil
}

// createAssessmentResponseForRecipient creates a new assessment response record for the campaign and recipient email
func createAssessmentResponseForRecipient(ctx context.Context, db *generated.Client, camp *generated.Campaign, assessmentID string, email string, isTest bool) (*generated.AssessmentResponse, error) {
	create := db.AssessmentResponse.Create().
		SetAssessmentID(assessmentID).
		SetCampaignID(camp.ID).
		SetEmail(email).
		SetOwnerID(camp.OwnerID)

	if isTest {
		create.SetIsTest(true)
	}

	if camp.EntityID != "" {
		create.SetEntityID(camp.EntityID)
	}

	if camp.DueDate != nil && !camp.DueDate.IsZero() {
		create.SetDueDate(time.Time(*camp.DueDate))
	}

	return create.Save(ctx)
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

// TargetDispatchable reports whether a campaign target should be included in a
// dispatch attempt for the requested action semantics
func TargetDispatchable(status enums.AssessmentResponseStatus, sentAt *models.DateTime, resend bool, includeOverdue bool) bool {
	if sentAt != nil && !sentAt.IsZero() && !resend {
		return false
	}

	switch status {
	case enums.AssessmentResponseStatusCompleted:
		return false
	case enums.AssessmentResponseStatusOverdue:
		return includeOverdue || resend
	case enums.AssessmentResponseStatusSent:
		return resend
	default:
		return true
	}
}

// splitFullName splits a full name string into first and last components on the first space
func splitFullName(fullName string) (string, string) {
	name := strings.TrimSpace(fullName)
	if name == "" {
		return "", ""
	}

	first, last, _ := strings.Cut(name, " ")

	return first, strings.TrimSpace(last)
}
