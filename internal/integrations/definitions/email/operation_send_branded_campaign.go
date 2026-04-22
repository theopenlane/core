package email

import (
	"context"
	"encoding/json"
	"fmt"
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

// SendBrandedCampaignRequest is the operation config for dispatching a branded email campaign
type SendBrandedCampaignRequest struct {
	// CampaignID is the identifier of the campaign to dispatch
	CampaignID string `json:"campaignId" jsonschema:"required,description=Campaign identifier"`
}

// SendBrandedCampaign dispatches templated branded emails to all pending campaign targets
type SendBrandedCampaign struct{}

// Handle returns the typed operation handler for builder registration
func (s SendBrandedCampaign) Handle() types.OperationHandler {
	return providerkit.WithClientRequestConfig(emailClientRef, SendBrandedCampaignOp, ErrTemplateRenderFailed, s.Run)
}

// Run loads the campaign, iterates pending targets, and dispatches one email per target
// through the catalog dispatcher identified by the linked EmailTemplate.Key.
// Targets with sent_at already set are skipped. Failed sends are logged and processing
// continues so a single bad address does not abort the entire dispatch
func (SendBrandedCampaign) Run(ctx context.Context, req types.OperationRequest, client *EmailClient, cfg SendBrandedCampaignRequest) (json.RawMessage, error) {
	camp, err := req.DB.Campaign.Query().
		Where(campaign.IDEQ(cfg.CampaignID)).
		WithEmailTemplate(func(q *generated.EmailTemplateQuery) {
			q.WithFiles(func(fq *generated.FileQuery) {
				fq.Select(
					file.FieldProvidedFileName,
					file.FieldProvidedFileExtension,
					file.FieldDetectedMimeType,
					file.FieldFileContents)
			})
		}).Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, ErrCampaignNotFound
		}

		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", cfg.CampaignID).Msg("failed loading campaign for email dispatch")

		return nil, ErrCampaignNotFound
	}

	template := camp.Edges.EmailTemplate
	if template == nil {
		return nil, nil
	}

	dispatcher, ok := DispatcherByKey(template.Key)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDispatcherNotFound, template.Key)
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

	attachments := staticAttachmentsFromFiles(ctx, template.Edges.Files)
	campaignOverlay := CampaignContext{
		CampaignID:          camp.ID,
		CampaignName:        camp.Name,
		CampaignDescription: camp.Description,
	}

	for _, target := range targets {
		if err := sendAndMarkCampaignTarget(ctx, req, client, dispatcher, template.Defaults, campaignOverlay, target, attachments); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("campaign_id", cfg.CampaignID).Str("target_id", target.ID).Msg("failed dispatching campaign email to target")
		}
	}

	return nil, nil
}

// sendAndMarkCampaignTarget dispatches a single campaign email then marks the target as sent.
// The send is an external provider call that cannot be rolled back, so we send first and mark
// after; a failed mark means the target may be re-sent on retry, which is preferable to
// marking sent without actually sending
func sendAndMarkCampaignTarget(ctx context.Context, req types.OperationRequest, client *EmailClient, dispatcher EmailDispatcher, defaults map[string]any, campaignOverlay CampaignContext, target *generated.CampaignTarget, attachments []*newman.Attachment) error {
	first, last := splitFullName(target.FullName)

	payload, err := buildDispatchPayload(defaults,
		RecipientInfo{Email: target.Email, FirstName: first, LastName: last},
		campaignOverlay,
	)
	if err != nil {
		return err
	}

	extraOpts := []newman.MessageOption{
		newman.WithTag(newman.Tag{Name: TagCampaignTargetID, Value: target.ID}),
		newman.WithAttachments(attachments),
	}

	if err := dispatcher.SendByKey(ctx, req, client, payload, extraOpts...); err != nil {
		return err
	}

	now := models.DateTime(time.Now())
	if err := req.DB.CampaignTarget.UpdateOneID(target.ID).SetSentAt(now).Exec(ctx); err != nil {
		return fmt.Errorf("mark sent: %w", err)
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
