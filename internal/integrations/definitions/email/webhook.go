package email

import (
	"context"
	"encoding/json"
	"time"

	"github.com/resend/resend-go/v3"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/internal/integrations/types"
)

// resendWebhookEvent represents the structure of a Resend webhook event payload
type resendWebhookEvent struct {
	// Type is the event type, e.g. "email.sent", "email.delivered"
	Type string `json:"type"`
	// Data contains the event-specific data
	Data resendWebhookData `json:"data"`
}

// resendWebhookData contains the data portion of a Resend webhook event
type resendWebhookData struct {
	// EmailID is the unique identifier for the email in Resend
	EmailID string `json:"email_id"`
	// CreatedAt is the timestamp when the event occurred
	CreatedAt string `json:"created_at"`
	// Tags are the custom tags attached to the email for tracking purposes
	Tags []resendWebhookTag `json:"tags"`
}

// resendWebhookTag represents a single tag from a Resend webhook event
type resendWebhookTag struct {
	// Name is the tag key
	Name string `json:"name"`
	// Value is the tag value
	Value string `json:"value"`
}

// ResendWebhook implements verification and event resolution for inbound Resend webhooks
type ResendWebhook struct {
	// Secret is the Svix signing secret used to verify inbound webhook payloads
	Secret string
}

// ResendDeliveryEvent handles email delivery status updates from Resend
type ResendDeliveryEvent struct{}

// Verify authenticates an inbound Resend webhook using Svix signature verification
func (w ResendWebhook) Verify(req types.WebhookInboundRequest) error {
	headers := resend.WebhookHeaders{
		Id:        req.Request.Header.Get("svix-id"),
		Timestamp: req.Request.Header.Get("svix-timestamp"),
		Signature: req.Request.Header.Get("svix-signature"),
	}

	if headers.Id == "" {
		return ErrWebhookMissingID
	}

	if w.Secret == "" {
		return ErrWebhookSecretMissing
	}

	client := resend.NewClient("")

	return client.Webhooks.Verify(&resend.VerifyWebhookOptions{
		Payload:       string(req.Payload),
		Headers:       headers,
		WebhookSecret: w.Secret,
	})
}

// Event resolves an inbound Resend webhook payload into a registered event
func (ResendWebhook) Event(req types.WebhookInboundRequest) (types.WebhookReceivedEvent, error) {
	var event resendWebhookEvent
	if err := json.Unmarshal(req.Payload, &event); err != nil {
		return types.WebhookReceivedEvent{}, ErrWebhookPayloadInvalid
	}

	return types.WebhookReceivedEvent{
		Name:       event.Type,
		DeliveryID: req.Request.Header.Get("svix-id"),
		Payload:    req.Payload,
	}, nil
}

// Handle processes one Resend delivery event by updating the relevant campaign target
// and assessment response records
func (ResendDeliveryEvent) Handle(ctx context.Context, req types.WebhookHandleRequest) error {
	var event resendWebhookEvent
	if err := json.Unmarshal(req.Event.Payload, &event); err != nil {
		return ErrWebhookPayloadInvalid
	}

	if event.Type == "" {
		return nil
	}

	db := ent.FromContext(ctx)

	tagMap := extractTags(event.Data.Tags)

	isTest := tagMap[TagIsTest] == "true"

	if assessmentResponseID := tagMap[TagAssessmentResponseID]; assessmentResponseID != "" {
		if err := updateAssessmentResponse(ctx, db, assessmentResponseID, event, isTest); err != nil {
			return err
		}
	}

	if campaignTargetID := tagMap[TagCampaignTargetID]; campaignTargetID != "" && !isTest {
		if err := updateCampaignTarget(ctx, db, campaignTargetID, "", event); err != nil {
			return err
		}
	}

	return nil
}

// extractTags builds a key→value map from Resend webhook tags
func extractTags(tags []resendWebhookTag) map[string]string {
	m := make(map[string]string, len(tags))

	for _, tag := range tags {
		if tag.Name != "" && tag.Value != "" {
			m[tag.Name] = tag.Value
		}
	}

	return m
}

// updateAssessmentResponse applies delivery events to an assessment response record
func updateAssessmentResponse(ctx context.Context, db *ent.Client, responseID string, event resendWebhookEvent, isTest bool) error {
	resp, err := db.AssessmentResponse.Get(ctx, responseID)
	if err != nil {
		return err
	}

	update := db.AssessmentResponse.UpdateOneID(responseID)
	eventTime := parseResendEventTime(event.Data.CreatedAt)

	if eventTime != nil {
		update.SetLastEmailEventAt(*eventTime)
	}

	switch event.Type {
	case resend.EventEmailSent:
		update.SetStatus(enums.AssessmentResponseStatusSent)
	case resend.EventEmailDelivered:
		if eventTime != nil {
			update.SetEmailDeliveredAt(*eventTime)
		}
	case resend.EventEmailOpened:
		if eventTime != nil {
			update.SetEmailOpenedAt(*eventTime)
		}

		update.AddEmailOpenCount(1)
	case resend.EventEmailClicked:
		if eventTime != nil {
			update.SetEmailClickedAt(*eventTime)
		}

		update.AddEmailClickCount(1)
	case resend.EventEmailBounced, resend.EventEmailFailed:
		metadata := resp.EmailMetadata
		if metadata == nil {
			metadata = map[string]any{}
		}

		metadata["resend_event"] = event.Type
		metadata["resend_email_id"] = event.Data.EmailID
		metadata["resend_timestamp"] = event.Data.CreatedAt
		update.SetEmailMetadata(metadata)
	}

	if _, err := update.Save(ctx); err != nil {
		return err
	}

	if resp.CampaignID != "" && !resp.IsTest && !isTest {
		return updateCampaignTarget(ctx, db, "", resp.CampaignID, event, resp.Email)
	}

	return nil
}

// updateCampaignTarget applies delivery events to a campaign target record
func updateCampaignTarget(ctx context.Context, db *ent.Client, targetID, campaignID string, event resendWebhookEvent, email ...string) error {
	var (
		target *ent.CampaignTarget
		err    error
	)

	switch {
	case targetID != "":
		target, err = db.CampaignTarget.Get(ctx, targetID)
	case campaignID != "" && len(email) > 0 && email[0] != "":
		target, err = db.CampaignTarget.Query().
			Where(
				campaigntarget.CampaignIDEQ(campaignID),
				campaigntarget.EmailEqualFold(email[0]),
			).
			Only(ctx)
	default:
		return nil
	}

	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}

		return err
	}

	update := db.CampaignTarget.UpdateOneID(target.ID)

	eventTime := parseResendEventTime(event.Data.CreatedAt)

	switch event.Type {
	case resend.EventEmailSent, resend.EventEmailDelivered, resend.EventEmailOpened, resend.EventEmailClicked:
		update.SetStatus(enums.AssessmentResponseStatusSent)

		if eventTime != nil {
			update.SetSentAt(models.DateTime(*eventTime))
		}
	case resend.EventEmailBounced, resend.EventEmailFailed:
		metadata := target.Metadata
		if metadata == nil {
			metadata = map[string]any{}
		}

		metadata["resend_event"] = event.Type
		metadata["resend_email_id"] = event.Data.EmailID
		metadata["resend_timestamp"] = event.Data.CreatedAt
		update.SetMetadata(metadata)
	}

	_, err = update.Save(ctx)

	return err
}

// parseResendEventTime parses a Resend timestamp string and returns a time.Time pointer
func parseResendEventTime(value string) *time.Time {
	if value == "" {
		return nil
	}

	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return &parsed
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed
	}

	return nil
}
