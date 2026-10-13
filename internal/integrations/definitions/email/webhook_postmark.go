package email

import (
	"context"
	"crypto/subtle"
	"encoding/json"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// postmarkTraceIDHeader carries the Postmark delivery id, stable across retries of one event
	postmarkTraceIDHeader = "X-PM-Webhook-Trace-Id"
	// postmarkRecordDelivery is the Postmark record type for delivered messages
	postmarkRecordDelivery = "Delivery"
	// postmarkRecordBounce is the Postmark record type for bounced messages
	postmarkRecordBounce = "Bounce"
	// postmarkRecordSpamComplaint is the Postmark record type for spam complaints
	postmarkRecordSpamComplaint = "SpamComplaint"
	// postmarkRecordOpen is the Postmark record type for message opens
	postmarkRecordOpen = "Open"
	// postmarkRecordClick is the Postmark record type for tracked link clicks
	postmarkRecordClick = "Click"
)

// postmarkRecordTypes lists the Postmark record types dispatched to PostmarkDeliveryEvent
var postmarkRecordTypes = []string{
	postmarkRecordDelivery,
	postmarkRecordBounce,
	postmarkRecordSpamComplaint,
	postmarkRecordOpen,
	postmarkRecordClick,
}

// postmarkRecord is the minimal Postmark webhook payload needed to route an event
type postmarkRecord struct {
	// RecordType identifies the Postmark event, e.g. "Delivery" or "Bounce"
	RecordType string `json:"RecordType"`
}

// postmarkWebhookEvent is the union of Postmark webhook payload fields used for delivery tracking
type postmarkWebhookEvent struct {
	// RecordType identifies the Postmark event, e.g. "Delivery" or "Bounce"
	RecordType string `json:"RecordType"`
	// MessageID is the Postmark message id
	MessageID string `json:"MessageID"`
	// DeliveredAt is the delivery timestamp on Delivery records
	DeliveredAt string `json:"DeliveredAt"`
	// BouncedAt is the event timestamp on Bounce and SpamComplaint records
	BouncedAt string `json:"BouncedAt"`
	// ReceivedAt is the event timestamp on Open and Click records
	ReceivedAt string `json:"ReceivedAt"`
	// Metadata carries the tags attached to the message at send time
	Metadata map[string]string `json:"Metadata"`
}

// PostmarkWebhook implements verification and event resolution for inbound Postmark webhooks
type PostmarkWebhook struct {
	// Secret is the basic auth password configured on the Postmark webhook URL
	Secret string
}

// PostmarkDeliveryEvent handles email delivery status updates from Postmark
type PostmarkDeliveryEvent struct{}

// Verify authenticates an inbound Postmark webhook by comparing the basic auth password to the secret
func (w PostmarkWebhook) Verify(req types.WebhookInboundRequest) error {
	if w.Secret == "" {
		return ErrWebhookSecretMissing
	}

	_, password, ok := req.Request.BasicAuth()
	if !ok {
		return ErrWebhookAuthMissing
	}

	if subtle.ConstantTimeCompare([]byte(password), []byte(w.Secret)) != 1 {
		return ErrWebhookAuthInvalid
	}

	return nil
}

// Event resolves an inbound Postmark webhook payload, leaving Name empty for unhandled record types
func (PostmarkWebhook) Event(req types.WebhookInboundRequest) (types.WebhookReceivedEvent, error) {
	var record postmarkRecord
	if err := json.Unmarshal(req.Payload, &record); err != nil {
		logx.FromContext(req.Request.Context()).Error().Err(err).Msg("failed decoding postmark webhook event")

		return types.WebhookReceivedEvent{}, ErrWebhookPayloadInvalid
	}

	event := types.WebhookReceivedEvent{
		DeliveryID: req.Request.Header.Get(postmarkTraceIDHeader),
		Payload:    req.Payload,
	}

	if lo.Contains(postmarkRecordTypes, record.RecordType) {
		event.Name = record.RecordType
	}

	return event, nil
}

// Handle processes one Postmark delivery event by updating the relevant campaign target
// and assessment response records
func (PostmarkDeliveryEvent) Handle(ctx context.Context, req types.WebhookHandleRequest) error {
	var event postmarkWebhookEvent
	if err := json.Unmarshal(req.Event.Payload, &event); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed decoding postmark webhook delivery event")

		return ErrWebhookPayloadInvalid
	}

	delivery, ok := event.toDeliveryEvent()
	if !ok {
		return nil
	}

	return delivery.apply(ctx)
}

// toDeliveryEvent maps a Postmark payload to a delivery event, reporting false for unhandled record types
func (e postmarkWebhookEvent) toDeliveryEvent() (deliveryEvent, bool) {
	var (
		kind      deliveryEventKind
		timestamp string
	)

	switch e.RecordType {
	case postmarkRecordDelivery:
		kind, timestamp = deliveryEventDelivered, e.DeliveredAt
	case postmarkRecordOpen:
		kind, timestamp = deliveryEventOpened, e.ReceivedAt
	case postmarkRecordClick:
		kind, timestamp = deliveryEventClicked, e.ReceivedAt
	case postmarkRecordBounce:
		kind, timestamp = deliveryEventBounced, e.BouncedAt
	case postmarkRecordSpamComplaint:
		kind, timestamp = deliveryEventComplained, e.BouncedAt
	default:
		return deliveryEvent{}, false
	}

	return deliveryEvent{
		provider:   providerPostmark,
		name:       e.RecordType,
		kind:       kind,
		emailID:    e.MessageID,
		timestamp:  timestamp,
		occurredAt: parseEventTime(timestamp),
		tags:       e.Metadata,
	}, true
}
