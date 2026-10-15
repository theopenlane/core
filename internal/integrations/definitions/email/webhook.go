package email

import (
	"context"
	"encoding/json"
	"time"

	"github.com/resend/resend-go/v3"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// deliveryEventKind is the provider-neutral delivery status carried by an inbound email webhook
type deliveryEventKind string

const (
	// deliveryEventSent means the provider accepted the message for sending
	deliveryEventSent deliveryEventKind = "sent"
	// deliveryEventDelivered means the recipient mail server accepted the message
	deliveryEventDelivered deliveryEventKind = "delivered"
	// deliveryEventOpened means the recipient opened the message
	deliveryEventOpened deliveryEventKind = "opened"
	// deliveryEventClicked means the recipient clicked a tracked link
	deliveryEventClicked deliveryEventKind = "clicked"
	// deliveryEventBounced means the message bounced
	deliveryEventBounced deliveryEventKind = "bounced"
	// deliveryEventComplained means the recipient marked the message as spam
	deliveryEventComplained deliveryEventKind = "complained"
	// deliveryEventFailed means the provider failed to send the message
	deliveryEventFailed deliveryEventKind = "failed"
)

const (
	// metadataEventSuffix is the metadata key suffix holding the raw provider event name
	metadataEventSuffix = "_event"
	// metadataEmailIDSuffix is the metadata key suffix holding the provider message id
	metadataEmailIDSuffix = "_email_id"
	// metadataTimestampSuffix is the metadata key suffix holding the raw provider event timestamp
	metadataTimestampSuffix = "_timestamp"
	// targetMetadataEmailIDKey is the per-event campaign target metadata key holding the provider message id
	targetMetadataEmailIDKey = "email_id"
	// targetMetadataTimestampKey is the per-event campaign target metadata key holding the raw event timestamp
	targetMetadataTimestampKey = "timestamp"
	// TagTrueValue is the tag value that marks a boolean tag such as TagIsTest as set
	TagTrueValue = "true"
)

// resendEventKinds maps handled Resend event types to delivery event kinds
var resendEventKinds = map[string]deliveryEventKind{
	resend.EventEmailSent:      deliveryEventSent,
	resend.EventEmailDelivered: deliveryEventDelivered,
	resend.EventEmailOpened:    deliveryEventOpened,
	resend.EventEmailClicked:   deliveryEventClicked,
	resend.EventEmailBounced:   deliveryEventBounced,
	resend.EventEmailFailed:    deliveryEventFailed,
}

// deliveryEvent is the provider-neutral form of one inbound email delivery webhook
type deliveryEvent struct {
	// provider is the provider name used to prefix stored metadata keys
	provider string
	// name is the raw provider event name, e.g. "email.bounced" or "Bounce"
	name string
	// kind is the provider-neutral delivery status
	kind deliveryEventKind
	// emailID is the provider message id
	emailID string
	// timestamp is the raw provider event timestamp as stored in metadata
	timestamp string
	// occurredAt is the parsed event timestamp, nil when absent or unparsable
	occurredAt *time.Time
	// tags are the correlation tags attached to the message at send time
	tags map[string]string
}

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
	// Resend sends this as:
	//
	//	"tags": {
	//	  "assessment_resp_id": "...",
	//	  "is_test": "true"
	//	}
	//
	// not:
	//
	//	"tags": [
	//	  {"name": "is_test", "value": "true"},
	//	  {"name": "assessment_resp_id", "value": "..."}
	//	]
	Tags map[string]string `json:"tags"`
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
		logx.FromContext(req.Request.Context()).Error().Err(err).Msg("failed decoding resend webhook event")

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
		logx.FromContext(ctx).Error().Err(err).Msg("failed decoding resend webhook delivery event")

		return ErrWebhookPayloadInvalid
	}

	delivery, ok := event.toDeliveryEvent()
	if !ok {
		return nil
	}

	return delivery.apply(ctx)
}

// toDeliveryEvent maps a Resend payload to a delivery event, reporting false for unhandled types
func (e resendWebhookEvent) toDeliveryEvent() (deliveryEvent, bool) {
	kind, ok := resendEventKinds[e.Type]
	if !ok {
		return deliveryEvent{}, false
	}

	return deliveryEvent{
		provider:   providerResend,
		name:       e.Type,
		kind:       kind,
		emailID:    e.Data.EmailID,
		timestamp:  e.Data.CreatedAt,
		occurredAt: parseEventTime(e.Data.CreatedAt),
		tags:       e.Data.Tags,
	}, true
}

// apply updates the assessment response and campaign target records correlated by the event tags
func (e deliveryEvent) apply(ctx context.Context) error {
	db := ent.FromContext(ctx)

	flags := gala.WorkflowFlagsKey.GetOr(ctx, gala.WorkflowFlags{})
	flags.Bypass = true
	ctx = gala.WorkflowFlagsKey.Set(ctx, flags)

	isTest := e.tags[TagIsTest] == TagTrueValue

	assessmentResponseID := lo.CoalesceOrEmpty(e.tags[TagAssessmentResponseID], e.tags[tagAssessmentResponseIDPrior])
	if assessmentResponseID != "" {
		if err := updateAssessmentResponse(ctx, db, assessmentResponseID, e, isTest); err != nil {
			return err
		}
	}

	if campaignTargetID := e.tags[TagCampaignTargetID]; campaignTargetID != "" && !isTest {
		if err := updateCampaignTarget(ctx, db, campaignTargetID, "", e); err != nil {
			return err
		}
	}

	return nil
}

// updateAssessmentResponse applies delivery events to an assessment response record
func updateAssessmentResponse(ctx context.Context, db *ent.Client, responseID string, event deliveryEvent, isTest bool) error {
	resp, err := db.AssessmentResponse.Get(ctx, responseID)
	if err != nil {
		return err
	}

	ctx = auth.EnsureIntegrationCaller(ctx, resp.OwnerID)

	update := db.AssessmentResponse.UpdateOneID(responseID)
	eventTime := event.occurredAt

	if eventTime != nil {
		update.SetLastEmailEventAt(*eventTime)
	}

	switch event.kind {
	case deliveryEventSent:
		update.SetStatus(enums.AssessmentResponseStatusSent)
	case deliveryEventDelivered:
		if eventTime != nil {
			update.SetEmailDeliveredAt(*eventTime)
		}
	case deliveryEventOpened:
		if eventTime != nil {
			update.SetEmailOpenedAt(*eventTime)
		}

		update.AddEmailOpenCount(1)
	case deliveryEventClicked:
		if eventTime != nil {
			update.SetEmailClickedAt(*eventTime)
		}

		update.AddEmailClickCount(1)
	case deliveryEventBounced, deliveryEventComplained, deliveryEventFailed:
		metadata := resp.EmailMetadata
		if metadata == nil {
			metadata = map[string]any{}
		}

		metadata[event.provider+metadataEventSuffix] = event.name
		metadata[event.provider+metadataEmailIDSuffix] = event.emailID
		metadata[event.provider+metadataTimestampSuffix] = event.timestamp
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
func updateCampaignTarget(ctx context.Context, db *ent.Client, targetID, campaignID string, event deliveryEvent, email ...string) error {
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

	ctx = auth.EnsureIntegrationCaller(ctx, target.OwnerID)

	update := db.CampaignTarget.UpdateOneID(target.ID)

	switch event.kind {
	case deliveryEventSent:
		update.SetStatus(enums.AssessmentResponseStatusSent)

		if event.occurredAt != nil {
			update.SetSentAt(models.DateTime(*event.occurredAt))
		}
	case deliveryEventDelivered, deliveryEventOpened, deliveryEventClicked:
		update.SetStatus(enums.AssessmentResponseStatusSent)
	}

	update.SetMetadata(getTargetMetadata(target.Metadata, event))

	_, err = update.Save(ctx)

	return err
}

// getTargetMetadata records the event under its kind in the campaign target metadata
func getTargetMetadata(metadata map[string]any, event deliveryEvent) map[string]any {
	if metadata == nil {
		metadata = map[string]any{}
	}

	metadata[string(event.kind)] = map[string]string{
		targetMetadataEmailIDKey:   event.emailID,
		targetMetadataTimestampKey: event.timestamp,
	}

	return metadata
}

// parseEventTime parses a provider timestamp string and returns a time.Time pointer
func parseEventTime(value string) *time.Time {
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
