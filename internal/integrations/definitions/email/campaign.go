package email

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/scrubber"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/templatecontext"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/mapx"
)

// handleCampaignActivation evaluates a Campaign mutation and returns operation config
// when the campaign transitions to ACTIVE status
func handleCampaignActivation(_ context.Context, payload types.MutationPayload) (json.RawMessage, error) {
	if !lo.Contains(payload.ChangedFields, "status") {
		return nil, nil
	}

	proposed, _ := payload.ProposedChanges["status"].(string)
	if proposed != enums.CampaignStatusActive.String() {
		return nil, nil
	}

	return json.Marshal(SendCampaignRequest{CampaignID: payload.EntityID})
}

// SendCampaignEmails iterates all pending campaign targets and queues one templated email per recipient.
// Targets with sent_at already set are skipped. Failed sends are logged and processing continues
// so a single bad address does not abort the entire dispatch
func SendCampaignEmails(ctx context.Context, db *generated.Client, emailClient *EmailClient, campaignID string) error {
	if db.Job == nil {
		return ErrJobClientRequired
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

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

	targets, err := db.CampaignTarget.Query().
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
		if err := sendAndMarkTarget(ctx, db, emailClient, camp, emailRecord, notificationRecord, target); err != nil {
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
func sendAndMarkTarget(ctx context.Context, db *generated.Client, emailClient *EmailClient, camp *generated.Campaign, emailRecord *generated.EmailTemplate, notificationRecord *generated.NotificationTemplate, target *generated.CampaignTarget) error {
	tx, err := db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := sendCampaignTargetEmail(ctx, tx.Client(), emailClient, camp, emailRecord, notificationRecord, target); err != nil {
		_ = tx.Rollback()
		return err
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	now := models.DateTime(time.Now())
	if err := tx.CampaignTarget.UpdateOneID(target.ID).SetSentAt(now).Exec(allowCtx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("mark sent: %w", err)
	}

	return tx.Commit()
}

// sendCampaignTargetEmail composes and queues one email for a single campaign target
func sendCampaignTargetEmail(ctx context.Context, db *generated.Client, emailClient *EmailClient, camp *generated.Campaign, emailRecord *generated.EmailTemplate, notificationRecord *generated.NotificationTemplate, target *generated.CampaignTarget) error {
	recipient := campaignRecipient(target)

	td := buildCampaignTemplateData(emailClient.Config, camp, recipient, emailRecord, notificationRecord)

	eb := camp.Edges.EmailBranding
	if eb != nil {
		td.Branding = brandingTemplateData(eb)
	}

	data, err := td.ToTemplateData()
	if err != nil {
		return err
	}

	if err := validateTemplateData(emailRecord.Jsonconfig, data); err != nil {
		return err
	}

	rendered, err := renderDBEnvelope(ctx, db, emailRecord, data, eb)
	if err != nil {
		return err
	}

	message := newman.NewEmailMessageWithOptions(
		newman.WithFrom(emailClient.Config.FromEmail),
		newman.WithTo([]string{target.Email}),
		newman.WithSubject(rendered.Subject),
		newman.WithHTML(rendered.HTML),
		newman.WithText(rendered.Text),
	)

	message.SetCustomHTMLScrubber(scrubber.ScrubberFunc(renderTimeHTMLSanitize))

	for _, a := range staticAttachmentsFromFiles(ctx, emailRecord.Edges.Files) {
		message.AddAttachment(a)
	}

	if _, err := db.Job.Insert(ctx, jobs.EmailArgs{Message: *message}, nil); err != nil {
		return fmt.Errorf("%w: %w", ErrQueueInsertFailed, err)
	}

	return nil
}

// buildCampaignTemplateData constructs a typed CampaignData struct from config, campaign,
// recipient, and template default values. Email template defaults form the base layer;
// notification defaults override
func buildCampaignTemplateData(config RuntimeEmailConfig, camp *generated.Campaign, recipient templatecontext.RecipientData, emailRecord *generated.EmailTemplate, notificationRecord *generated.NotificationTemplate) *templatecontext.CampaignData {
	vars := make(map[string]any, len(emailRecord.Defaults)+len(notificationRecord.Defaults))
	maps.Copy(vars, emailRecord.Defaults)
	maps.Copy(vars, notificationRecord.Defaults)

	return &templatecontext.CampaignData{
		ContextData: contextDataFromConfig(config, recipient, vars),
		Campaign: templatecontext.CampaignMeta{
			Name:        camp.Name,
			Description: camp.Description,
		},
	}
}

// campaignRecipient builds a RecipientData from a campaign target
func campaignRecipient(target *generated.CampaignTarget) templatecontext.RecipientData {
	first, last := splitFullName(target.FullName)

	return templatecontext.RecipientData{
		Email:     target.Email,
		FirstName: first,
		LastName:  last,
	}
}

// contextDataFromConfig constructs a ContextData from a RuntimeEmailConfig and recipient
func contextDataFromConfig(config RuntimeEmailConfig, recipient templatecontext.RecipientData, vars map[string]any) templatecontext.ContextData {
	return templatecontext.ContextData{
		CompanyName:    config.CompanyName,
		CompanyAddress: config.CompanyAddress,
		Corporation:    config.Corporation,
		FromEmail:      config.FromEmail,
		SupportEmail:   config.SupportEmail,
		LogoURL:        config.LogoURL,
		RootURL:        config.RootURL,
		ProductURL:     config.ProductURL,
		DocsURL:        config.DocsURL,
		Recipient:      recipient,
		Vars:           vars,
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
// Used when an EmailTemplate is referenced directly (e.g. from a campaign) without an explicit NotificationTemplate record
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
