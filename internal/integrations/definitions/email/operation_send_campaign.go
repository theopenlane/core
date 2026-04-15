package email

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/theopenlane/newman"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaign"
	"github.com/theopenlane/core/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// SendCampaignRequest is the operation config for dispatching a full campaign
type SendCampaignRequest struct {
	// CampaignID is the identifier of the campaign to dispatch
	CampaignID string `json:"campaignId" jsonschema:"required,description=Campaign identifier"`
}

// SendCampaign dispatches templated emails to all pending campaign targets
type SendCampaign struct{}

// Handle returns the typed operation handler for builder registration
func (s SendCampaign) Handle() types.OperationHandler {
	return providerkit.WithClientRequestConfig(emailClientRef, SendCampaignOp, ErrTemplateRenderFailed, s.Run)
}

// Run loads the campaign, iterates pending targets, renders and sends one email per target.
// Targets with sent_at already set are skipped. Failed sends are logged and processing
// continues so a single bad address does not abort the entire dispatch
func (SendCampaign) Run(ctx context.Context, req types.OperationRequest, client *EmailClient, cfg SendCampaignRequest) (json.RawMessage, error) {
	camp, err := req.DB.Campaign.Query().
		Where(campaign.IDEQ(cfg.CampaignID)).
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
			return nil, ErrCampaignNotFound
		}

		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", cfg.CampaignID).Msg("failed loading campaign for email dispatch")

		return nil, ErrCampaignNotFound
	}

	emailRecord := camp.Edges.EmailTemplate
	if emailRecord == nil {
		return nil, nil
	}

	targets, err := req.DB.CampaignTarget.Query().
		Where(
			campaigntarget.CampaignIDEQ(cfg.CampaignID),
			campaigntarget.SentAtIsNil(),
		).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", cfg.CampaignID).Msg("failed loading campaign targets")

		return nil, err
	}

	for _, target := range targets {
		if err := sendAndMarkTarget(ctx, req.DB, client, camp, emailRecord, target); err != nil {
			logx.FromContext(ctx).Error().Err(err).
				Str("campaign_id", cfg.CampaignID).
				Str("target_id", target.ID).
				Msg("failed dispatching campaign email to target")
		}
	}

	return nil, nil
}

// sendAndMarkTarget sends a single campaign email then marks the target as sent.
// The send is an external provider call that cannot be rolled back, so we send
// first and mark after. A failed mark means the target may be re-sent on retry,
// which is preferable to marking sent without actually sending
func sendAndMarkTarget(ctx context.Context, db *generated.Client, emailClient *EmailClient, camp *generated.Campaign, emailRecord *generated.EmailTemplate, target *generated.CampaignTarget) error {
	if err := sendCampaignTargetEmail(ctx, emailClient, camp, emailRecord, target); err != nil {
		return err
	}

	now := models.DateTime(time.Now())
	if err := db.CampaignTarget.UpdateOneID(target.ID).SetSentAt(now).Exec(ctx); err != nil {
		return fmt.Errorf("mark sent: %w", err)
	}

	return nil
}

// sendCampaignTargetEmail renders and sends one email for a single campaign target
// through the integration framework's email client
func sendCampaignTargetEmail(ctx context.Context, emailClient *EmailClient, camp *generated.Campaign, emailRecord *generated.EmailTemplate, target *generated.CampaignTarget) error {
	first, last := splitFullName(target.FullName)

	vars := make(map[string]any, len(emailRecord.Defaults)+len(camp.Metadata)+5) //nolint:mnd
	maps.Copy(vars, emailRecord.Defaults)
	maps.Copy(vars, camp.Metadata)

	vars["recipientEmail"] = target.Email
	vars["recipientFirstName"] = first
	vars["recipientLastName"] = last
	vars["campaignName"] = camp.Name
	vars["campaignDescription"] = camp.Description

	data, err := buildTemplateData(emailClient.Config, vars)
	if err != nil {
		return err
	}

	if err := validateTemplateData(emailRecord.Jsonconfig, data); err != nil {
		return err
	}

	rendered, err := renderDBEnvelope(emailRecord, data, camp.Edges.EmailBranding)
	if err != nil {
		return err
	}

	opts := []newman.MessageOption{
		newman.WithFrom(emailClient.Config.FromEmail),
		newman.WithTo([]string{target.Email}),
		newman.WithSubject(rendered.Subject),
		newman.WithHTML(rendered.HTML),
		newman.WithText(rendered.Text),
		newman.WithTag(newman.Tag{Name: TagCampaignTargetID, Value: target.ID}),
		newman.WithAttachments(staticAttachmentsFromFiles(ctx, emailRecord.Edges.Files)),
	}

	message := newman.NewEmailMessageWithOptions(opts...)

	if err := emailClient.Sender.SendEmailWithContext(ctx, message); err != nil {
		logx.FromContext(ctx).Error().Err(err).
			Str("campaign_id", camp.ID).
			Str("target_id", target.ID).
			Msg("failed sending campaign email")

		return fmt.Errorf("%w: %w", ErrSendFailed, err)
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
