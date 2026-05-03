package email

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/theopenlane/newman"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/templatekit"
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

// CampaignDispatchInput holds the shared dispatch parameters for campaign operations
type CampaignDispatchInput struct {
	// CampaignID is the identifier of the campaign to dispatch
	CampaignID string `json:"campaignId" jsonschema:"required,description=Campaign identifier"`
	// Resend allows previously sent, incomplete targets to receive another message
	Resend bool `json:"resend,omitempty" jsonschema:"description=Whether to include previously sent targets"`
	// IncludeOverdue includes overdue targets when dispatching incomplete recipients
	IncludeOverdue bool `json:"includeOverdue,omitempty" jsonschema:"description=Whether overdue targets should be included"`
}

// SendBrandedCampaignRequest is the operation config for dispatching a branded email campaign
type SendBrandedCampaignRequest struct {
	CampaignDispatchInput
}

// SendBrandedCampaign dispatches templated branded emails to all pending campaign targets
type SendBrandedCampaign struct{}

// Handle returns the typed operation handler for builder registration
func (s SendBrandedCampaign) Handle() types.OperationHandler {
	return providerkit.WithClientRequestConfig(emailClientRef, SendCampaignOp, ErrTemplateRenderFailed, s.Run)
}

// Run loads the campaign and pending targets, renders all messages, then sends
// via batch API or rate-limited individual sends depending on whether
// attachments are present. Returns a marshaled CampaignDispatchResult with counts
func (SendBrandedCampaign) Run(ctx context.Context, req types.OperationRequest, client *Client, cfg SendBrandedCampaignRequest) (json.RawMessage, error) {
	camp, dispatchable, skipped, err := loadCampaignWithTargets(ctx, req.DB, cfg.CampaignDispatchInput, func(q *generated.CampaignQuery) {
		q.WithEmailTemplate(func(tq *generated.EmailTemplateQuery) {
			tq.WithFiles(func(fq *generated.FileQuery) {
				fq.Select(
					file.FieldProvidedFileName,
					file.FieldProvidedFileExtension,
					file.FieldDetectedMimeType,
					file.FieldFileContents)
			})
		})
	})
	if err != nil {
		return nil, err
	}

	template := camp.Edges.EmailTemplate
	if template == nil {
		return nil, fmt.Errorf("%w: no email template linked to campaign", ErrDispatcherNotFound)
	}

	dispatcher, ok := DispatcherByKey(template.Key)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDispatcherNotFound, template.Key)
	}

	overlay := CampaignContext{
		CampaignID:          camp.ID,
		CampaignName:        camp.Name,
		CampaignDescription: camp.Description,
		CampaignDueDate:     formatCampaignDueDate(camp.DueDate),
	}

	messages, targetIDs, renderFailed := renderCampaignMessages(ctx, client, dispatcher, template.Defaults, camp.Metadata, overlay, dispatchable)
	attachments := staticAttachmentsFromFiles(ctx, template.Edges.Files)

	sentCount, sendFailed := sendCampaignMessages(ctx, req.DB, client, messages, targetIDs, attachments)

	result := CampaignDispatchResult{
		SentCount:    sentCount,
		SkippedCount: skipped,
		FailedCount:  renderFailed + sendFailed,
	}

	return json.Marshal(result)
}

// filterCampaignTargets returns the subset of targets that are dispatchable based on status, sent history, and resend/overdue flags
func filterCampaignTargets(targets []*generated.CampaignTarget, resend bool, includeOverdue bool) ([]*generated.CampaignTarget, int) {
	filtered := make([]*generated.CampaignTarget, 0, len(targets))

	for _, target := range targets {
		if TargetDispatchable(target.Status, target.SentAt, resend, includeOverdue) {
			filtered = append(filtered, target)
		}
	}

	return filtered, len(targets) - len(filtered)
}

// renderCampaignMessages renders all pending targets into newman messages
func renderCampaignMessages(ctx context.Context, client *Client, dispatcher Dispatcher, defaults map[string]any, campaignData map[string]any, overlay CampaignContext, targets []*generated.CampaignTarget) ([]*newman.EmailMessage, []string, int) {
	messages := make([]*newman.EmailMessage, 0, len(targets))
	targetIDs := make([]string, 0, len(targets))
	failed := 0

	for _, target := range targets {
		first, last := splitFullName(target.FullName)

		payload, err := templatekit.BuildDispatchPayload(defaults,
			campaignData,
			RecipientInfo{Email: target.Email, FirstName: first, LastName: last},
			overlay,
		)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("target_id", target.ID).Msg("failed building campaign payload")
			failed++

			continue
		}

		msg, err := dispatcher.RenderMessage(ctx, client, payload,
			newman.WithTag(newman.Tag{Name: TagCampaignTargetID, Value: target.ID}),
		)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("target_id", target.ID).Msg("failed rendering campaign message")
			failed++

			continue
		}

		messages = append(messages, msg)
		targetIDs = append(targetIDs, target.ID)
	}

	return messages, targetIDs, failed
}

// sendCampaignMessages sends rendered messages via batch API when no attachments
// are present, falling back to rate-limited individual sends otherwise.
// Returns (sent, failed) counts
func sendCampaignMessages(ctx context.Context, db *generated.Client, client *Client, messages []*newman.EmailMessage, targetIDs []string, attachments []*newman.Attachment) (int, int) {
	if len(attachments) > 0 {
		return sendCampaignIndividual(ctx, db, client, messages, targetIDs, attachments)
	}

	sent := 0

	for start := 0; start < len(messages); start += batchSize {
		end := min(start+batchSize, len(messages))
		batchMsgs := messages[start:end]
		batchIDs := targetIDs[start:end]

		if err := client.Sender.SendBatchEmailWithContext(ctx, batchMsgs); err != nil {
			if errors.Is(err, newman.ErrBatchNotImplemented) {
				logx.FromContext(ctx).Info().Msg("batch send not supported, falling back to individual sends")
				return sendCampaignIndividual(ctx, db, client, messages, targetIDs, nil)
			}

			logx.FromContext(ctx).Error().Err(err).Msg("batch send failed")

			return sent, len(messages) - sent
		}

		for _, id := range batchIDs {
			if err := markCampaignTargetSent(ctx, db, id); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("target_id", id).Msg("failed marking target sent")
			}
		}

		sent += len(batchMsgs)
	}

	return sent, 0
}

// sendCampaignIndividual sends messages one at a time with rate limiting
func sendCampaignIndividual(ctx context.Context, db *generated.Client, client *Client, messages []*newman.EmailMessage, targetIDs []string, attachments []*newman.Attachment) (int, int) {
	ticker := time.NewTicker(sendRateInterval)
	defer ticker.Stop()

	sent, failed := 0, 0

	for i, msg := range messages {
		if i > 0 {
			<-ticker.C
		}

		msg.Attachments = append(msg.Attachments, attachments...)

		if err := client.Sender.SendEmailWithContext(ctx, msg); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("target_id", targetIDs[i]).Msg("failed sending campaign email")
			failed++

			continue
		}

		sent++

		if err := markCampaignTargetSent(ctx, db, targetIDs[i]); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("target_id", targetIDs[i]).Msg("failed marking target sent")
		}
	}

	return sent, failed
}

// formatCampaignDueDate formats a campaign due date for display in email templates
func formatCampaignDueDate(d *models.DateTime) string {
	if d == nil || d.IsZero() {
		return ""
	}

	return time.Time(*d).Format("January 2, 2006")
}
