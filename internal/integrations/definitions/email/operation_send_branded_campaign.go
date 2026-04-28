package email

import (
	"context"
	"encoding/json"
	"errors"
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

const (
	// batchSize is the maximum number of recipients per bulk send API call
	batchSize = 100
	// sendRateInterval is the minimum interval between individual sends to stay
	// within the provider rate limit of 20 messages per second
	sendRateInterval = 50 * time.Millisecond
)

// SendBrandedCampaignRequest is the operation config for dispatching a branded email campaign
type SendBrandedCampaignRequest struct {
	// CampaignID is the identifier of the campaign to dispatch
	CampaignID string `json:"campaignId" jsonschema:"required,description=Campaign identifier"`
	// Resend allows previously sent, incomplete targets to receive another message
	Resend bool `json:"resend,omitempty" jsonschema:"description=Whether to include previously sent targets"`
	// IncludeOverdue includes overdue targets when dispatching incomplete recipients
	IncludeOverdue bool `json:"includeOverdue,omitempty" jsonschema:"description=Whether overdue targets should be included"`
}

// SendBrandedCampaign dispatches templated branded emails to all pending campaign targets
type SendBrandedCampaign struct{}

// Handle returns the typed operation handler for builder registration
func (s SendBrandedCampaign) Handle() types.OperationHandler {
	return providerkit.WithClientRequestConfig(emailClientRef, SendBrandedCampaignOp, ErrTemplateRenderFailed, s.Run)
}

// Run loads the campaign and pending targets, creates assessment responses for
// tracking, renders all messages, then sends via batch API or rate-limited
// individual sends depending on whether attachments are present
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
		Where(campaigntarget.CampaignIDEQ(cfg.CampaignID)).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", cfg.CampaignID).Msg("failed loading campaign targets")

		return nil, err
	}

	targets = filterCampaignTargets(targets, cfg.Resend, cfg.IncludeOverdue)

	overlay := CampaignContext{
		CampaignID:          camp.ID,
		CampaignName:        camp.Name,
		CampaignDescription: camp.Description,
		CampaignDueDate:     formatCampaignDueDate(camp.DueDate),
	}

	messages, targetIDs := renderCampaignMessages(ctx, client, dispatcher, template.Defaults, camp.Metadata, overlay, targets)
	attachments := staticAttachmentsFromFiles(ctx, template.Edges.Files)

	return nil, sendCampaignMessages(ctx, req.DB, client, messages, targetIDs, attachments)
}

func filterCampaignTargets(targets []*generated.CampaignTarget, resend bool, includeOverdue bool) []*generated.CampaignTarget {
	filtered := make([]*generated.CampaignTarget, 0, len(targets))

	for _, target := range targets {
		if TargetDispatchable(target.Status, target.SentAt, resend, includeOverdue) {
			filtered = append(filtered, target)
		}
	}

	return filtered
}

// renderCampaignMessages renders all pending targets into newman messages
func renderCampaignMessages(ctx context.Context, client *EmailClient, dispatcher EmailDispatcher, defaults map[string]any, campaignData map[string]any, overlay CampaignContext, targets []*generated.CampaignTarget) ([]*newman.EmailMessage, []string) {
	messages := make([]*newman.EmailMessage, 0, len(targets))
	targetIDs := make([]string, 0, len(targets))

	for _, target := range targets {
		first, last := splitFullName(target.FullName)

		payload, err := buildDispatchPayload(defaults,
			campaignData,
			RecipientInfo{Email: target.Email, FirstName: first, LastName: last},
			overlay,
		)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("target_id", target.ID).Msg("failed building campaign payload")
			continue
		}

		msg, err := dispatcher.RenderMessage(ctx, client, payload,
			newman.WithTag(newman.Tag{Name: TagCampaignTargetID, Value: target.ID}),
		)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("target_id", target.ID).Msg("failed rendering campaign message")
			continue
		}

		messages = append(messages, msg)
		targetIDs = append(targetIDs, target.ID)
	}

	return messages, targetIDs
}

// sendCampaignMessages sends rendered messages via batch API when no attachments
// are present, falling back to rate-limited individual sends otherwise
func sendCampaignMessages(ctx context.Context, db *generated.Client, client *EmailClient, messages []*newman.EmailMessage, targetIDs []string, attachments []*newman.Attachment) error {
	if len(attachments) > 0 {
		return sendCampaignIndividual(ctx, db, client, messages, targetIDs, attachments)
	}

	for start := 0; start < len(messages); start += batchSize {
		end := min(start+batchSize, len(messages))
		batchMsgs := messages[start:end]
		batchIDs := targetIDs[start:end]

		if err := client.Sender.SendBatchEmailWithContext(ctx, batchMsgs); err != nil {
			if errors.Is(err, newman.ErrBatchNotImplemented) {
				logx.FromContext(ctx).Info().Msg("batch send not supported, falling back to individual sends")
				return sendCampaignIndividual(ctx, db, client, messages, targetIDs, nil)
			}

			return fmt.Errorf("%w: %w", ErrSendFailed, err)
		}

		for _, id := range batchIDs {
			if err := markCampaignTargetSent(ctx, db, id); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("target_id", id).Msg("failed marking target sent")
			}
		}
	}

	return nil
}

// sendCampaignIndividual sends messages one at a time with rate limiting
func sendCampaignIndividual(ctx context.Context, db *generated.Client, client *EmailClient, messages []*newman.EmailMessage, targetIDs []string, attachments []*newman.Attachment) error {
	ticker := time.NewTicker(sendRateInterval)
	defer ticker.Stop()

	for i, msg := range messages {
		if i > 0 {
			<-ticker.C
		}

		for _, a := range attachments {
			msg.Attachments = append(msg.Attachments, a)
		}

		if err := client.Sender.SendEmailWithContext(ctx, msg); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("target_id", targetIDs[i]).Msg("failed sending campaign email")
			continue
		}

		if err := markCampaignTargetSent(ctx, db, targetIDs[i]); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("target_id", targetIDs[i]).Msg("failed marking target sent")
		}
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

// formatCampaignDueDate formats a campaign due date for display in email templates
func formatCampaignDueDate(d *models.DateTime) string {
	if d == nil || d.IsZero() {
		return ""
	}

	return time.Time(*d).Format("January 2, 2006")
}
