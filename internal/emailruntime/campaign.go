package emailruntime

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/compose"
	"github.com/theopenlane/newman/scrubber"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/mapx"
)

// SendCampaignEmails iterates all pending campaign targets and queues one templated email per recipient.
// Targets with sent_at already set are skipped. Failed sends are logged and processing continues
// so a single bad address does not abort the entire dispatch.
func SendCampaignEmails(ctx context.Context, client *generated.Client, campaignID string) error {
	if client.Emailer == nil {
		return ErrEmailerNotConfigured
	}

	if client.Job == nil {
		return ErrJobClientRequired
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	camp, err := client.Campaign.Query().
		Where(campaign.IDEQ(campaignID)).
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
		Only(allowCtx)
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

	targets, err := client.CampaignTarget.Query().
		Where(
			campaigntarget.CampaignIDEQ(campaignID),
			campaigntarget.SentAtIsNil(),
		).
		All(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", campaignID).Msg("failed loading campaign targets")

		return err
	}

	notificationRecord := syntheticNotificationFromEmailTemplate(emailRecord)

	for _, target := range targets {
		if err := sendCampaignTargetEmail(ctx, client, camp, emailRecord, notificationRecord, target); err != nil {
			logx.FromContext(ctx).Error().Err(err).
				Str("campaign_id", campaignID).
				Str("target_id", target.ID).
				Msg("failed dispatching campaign email to target")

			continue
		}

		now := models.DateTime(time.Now())
		if updateErr := client.CampaignTarget.UpdateOneID(target.ID).SetSentAt(now).Exec(allowCtx); updateErr != nil {
			logx.FromContext(ctx).Error().Err(updateErr).Str("target_id", target.ID).Msg("failed marking campaign target as sent")
		}
	}

	return nil
}

// sendCampaignTargetEmail composes and queues one email for a single campaign target
func sendCampaignTargetEmail(ctx context.Context, client *generated.Client, camp *generated.Campaign, emailRecord *generated.EmailTemplate, notificationRecord *generated.NotificationTemplate, target *generated.CampaignTarget) error {
	recipient := campaignRecipient(target)

	data, err := buildCampaignTemplateData(*client.Emailer, camp, recipient, emailRecord, notificationRecord)
	if err != nil {
		return err
	}

	rendered, err := renderTemplateEnvelope(ctx, client, notificationRecord, emailRecord, data)
	if err != nil {
		return err
	}

	message := newman.NewEmailMessageWithOptions(
		newman.WithFrom(client.Emailer.FromEmail),
		newman.WithTo([]string{target.Email}),
		newman.WithSubject(rendered.Subject),
		newman.WithHTML(rendered.HTML),
		newman.WithText(rendered.Text),
	)

	message.SetCustomHTMLScrubber(scrubber.ScrubberFunc(renderTimeHTMLSanitize))

	for _, a := range staticAttachmentsFromFiles(ctx, emailRecord.Edges.Files) {
		message.AddAttachment(a)
	}

	if _, err := client.Job.Insert(ctx, jobs.EmailArgs{Message: *message}, nil); err != nil {
		return fmt.Errorf("%w: %w", ErrEmailQueueInsertFailed, err)
	}

	return nil
}

// buildCampaignTemplateData builds the merged template data map for a campaign recipient.
// Email template defaults form the base layer; notification defaults override; call-site data (recipient +
// campaign fields) takes highest precedence.
func buildCampaignTemplateData(config compose.Config, camp *generated.Campaign, recipient compose.Recipient, emailRecord *generated.EmailTemplate, notificationRecord *generated.NotificationTemplate) (map[string]any, error) {
	base, err := compose.BuildTemplateData(config, recipient, map[string]any{
		"Campaign": map[string]any{
			"Name":        camp.Name,
			"Description": camp.Description,
		},
	})
	if err != nil {
		return nil, err
	}

	data := make(map[string]any, len(emailRecord.Defaults)+len(notificationRecord.Defaults)+len(base))
	maps.Copy(data, emailRecord.Defaults)
	maps.Copy(data, notificationRecord.Defaults)
	maps.Copy(data, base)

	return data, nil
}

// campaignRecipient builds a compose.Recipient from a campaign target
func campaignRecipient(target *generated.CampaignTarget) compose.Recipient {
	first, last := splitFullName(target.FullName)

	return compose.Recipient{
		Email:     target.Email,
		FirstName: first,
		LastName:  last,
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

// syntheticNotificationFromEmailTemplate creates a minimal NotificationTemplate from an EmailTemplate.
// Used when an EmailTemplate is referenced directly (e.g. from a campaign) without an explicit NotificationTemplate record.
func syntheticNotificationFromEmailTemplate(email *generated.EmailTemplate) *generated.NotificationTemplate {
	return &generated.NotificationTemplate{
		Key:             email.Key,
		Name:            email.Name,
		Channel:         enums.ChannelEmail,
		Format:          email.Format,
		Locale:          email.Locale,
		SubjectTemplate: email.SubjectTemplate,
		BodyTemplate:    email.BodyTemplate,
		Active:          email.Active,
		SystemOwned:     email.SystemOwned,
		Jsonconfig:      mapx.DeepCloneMapAny(email.Jsonconfig),
		Defaults:        mapx.DeepCloneMapAny(email.Defaults),
	}
}
