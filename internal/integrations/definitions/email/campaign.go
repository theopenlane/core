package email

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/pkg/logx"
)

// SendCampaignEmails iterates all pending campaign targets and queues one templated email per recipient.
// Targets with sent_at already set are skipped. Failed sends are logged and processing continues
// so a single bad address does not abort the entire dispatch
func SendCampaignEmails(ctx context.Context, db *generated.Client, emailClient *EmailClient, campaignID string) error {
	if db.Job == nil {
		return ErrJobClientRequired
	}

	camp, err := db.Campaign.Query().
		Where(campaign.IDEQ(campaignID)).
		WithEmailBranding().
		WithEmailTemplate(func(q *generated.EmailTemplateQuery) {
			q.WithFiles(func(fq *generated.FileQuery) {
				fq.Select(
					file.FieldProvidedFileName,
					file.FieldProvidedFileExtension,
					file.FieldDetectedMimeType,
					file.FieldFileContents,
				)
			})
		}).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return ErrCampaignNotFound
		}

		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", campaignID).Msg("failed loading campaign for email dispatch")

		return ErrCampaignNotFound
	}

	emailRecord := camp.Edges.EmailTemplate
	if emailRecord == nil {
		return nil
	}

	targets, err := db.CampaignTarget.Query().
		Where(
			campaigntarget.CampaignIDEQ(campaignID),
			campaigntarget.SentAtIsNil(),
		).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", campaignID).Msg("failed loading campaign targets")

		return err
	}

	for _, target := range targets {
		if err := sendAndMarkTarget(ctx, db, emailClient, camp, emailRecord, target); err != nil {
			logx.FromContext(ctx).Error().Err(err).
				Str("campaign_id", campaignID).
				Str("target_id", target.ID).
				Msg("failed dispatching campaign email to target")
		}
	}

	return nil
}

// sendAndMarkTarget sends a single campaign email and marks the target as sent within a transaction.
// The transaction ensures the send and the mark are atomic — a send without a mark cannot occur
func sendAndMarkTarget(ctx context.Context, db *generated.Client, emailClient *EmailClient, camp *generated.Campaign, emailRecord *generated.EmailTemplate, target *generated.CampaignTarget) error {
	tx, err := db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := sendCampaignTargetEmail(ctx, tx.Client(), emailClient, camp, emailRecord, target); err != nil {
		_ = tx.Rollback()
		return err
	}

	now := models.DateTime(time.Now())
	if err := tx.CampaignTarget.UpdateOneID(target.ID).SetSentAt(now).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("mark sent: %w", err)
	}

	return tx.Commit()
}

// sendCampaignTargetEmail composes and queues one email for a single campaign target
func sendCampaignTargetEmail(ctx context.Context, db *generated.Client, emailClient *EmailClient, camp *generated.Campaign, emailRecord *generated.EmailTemplate, target *generated.CampaignTarget) error {
	first, last := splitFullName(target.FullName)

	vars := make(map[string]any, len(emailRecord.Defaults)+len(camp.Metadata)+5) //nolint:mnd
	maps.Copy(vars, emailRecord.Defaults)
	maps.Copy(vars, camp.Metadata)

	vars["recipientEmail"] = target.Email
	vars["recipientFirstName"] = first
	vars["recipientLastName"] = last
	vars["campaignName"] = camp.Name
	vars["campaignDescription"] = camp.Description

	eb := camp.Edges.EmailBranding

	data, err := buildTemplateData(emailClient.Config, vars)
	if err != nil {
		return err
	}

	if err := validateTemplateData(emailRecord.Jsonconfig, data); err != nil {
		return err
	}

	rendered, err := renderDBEnvelope(emailRecord, data, eb)
	if err != nil {
		return err
	}

	message := newman.NewEmailMessageWithOptions(
		newman.WithFrom(emailClient.Config.FromEmail),
		newman.WithTo([]string{target.Email}),
		newman.WithSubject(rendered.Subject),
		newman.WithHTML(rendered.HTML),
		newman.WithText(rendered.Text),
		newman.WithTag(newman.Tag{Name: TagCampaignTargetID, Value: target.ID}),
	)

	for _, a := range staticAttachmentsFromFiles(ctx, emailRecord.Edges.Files) {
		message.AddAttachment(a)
	}

	if _, err := db.Job.Insert(ctx, jobs.EmailArgs{Message: *message}, nil); err != nil {
		return fmt.Errorf("%w: %w", ErrQueueInsertFailed, err)
	}

	return nil
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
