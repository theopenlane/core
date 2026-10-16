package email

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestResendWebhookEvent_ValidPayload(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"type":"email.delivered","data":{"email_id":"e_123","created_at":"2025-01-01T00:00:00Z","tags":{}}}`)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("svix-id", "msg_abc")

	event, err := ResendWebhook{}.Event(types.WebhookInboundRequest{
		Request: req,
		Payload: payload,
	})
	require.NoError(t, err)
	assert.Equal(t, "email.delivered", event.Name)
	assert.Equal(t, "msg_abc", event.DeliveryID)
	assert.Equal(t, payload, event.Payload)
}

func TestResendWebhookEvent_ObjectTags(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{
  "created_at": "2026-08-10T14:39:49.535Z",
  "data": {
    "created_at": "2026-08-10T14:39:49.381Z",
    "email_id": "e671610b-f9f9-4903-be4c-c9d008920ac3",
    "from": "email@updates.example.com",
    "message_id": "<0100019fec1d951f-e56ad6b4-e9e4-40b2-8fae-3127162540e7-000000@email.amazonses.com>",
    "subject": "Access Vendor Security Assessment Questionnaire from demo-org-1786368515",
    "tags": {
      "assessment_resp_id": "01KZP1H01G9J21QT2XHSZ5BR98",
      "is_test": "true"
    },
    "to": [
      "user1@gmail.com"
    ]
  },
  "type": "email.sent"
}`)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("svix-id", "random_value")

	event, err := ResendWebhook{}.Event(types.WebhookInboundRequest{
		Request: req,
		Payload: payload,
	})
	require.NoError(t, err)
	assert.Equal(t, "email.sent", event.Name)
	assert.Equal(t, "random_value", event.DeliveryID)
	assert.Equal(t, payload, event.Payload)

	var decoded resendWebhookEvent
	require.NoError(t, json.Unmarshal(event.Payload, &decoded))
	assert.Equal(t, "01KZP1H01G9J21QT2XHSZ5BR98", decoded.Data.Tags[TagAssessmentResponseID])
	assert.Equal(t, "true", decoded.Data.Tags[TagIsTest])
}

func TestResendWebhookEvent_InvalidJSON(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	_, err := ResendWebhook{}.Event(types.WebhookInboundRequest{
		Request: req,
		Payload: json.RawMessage(`not json`),
	})
	require.ErrorIs(t, err, ErrWebhookPayloadInvalid)
}

func TestResendWebhookVerify_MissingSvixID(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	err := ResendWebhook{Secret: "whsec_test"}.Verify(types.WebhookInboundRequest{
		Request: req,
		Payload: json.RawMessage(`{}`),
	})
	require.ErrorIs(t, err, ErrWebhookMissingID)
}

func TestResendWebhookVerify_MissingSecret(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("svix-id", "msg_abc")

	err := ResendWebhook{Secret: ""}.Verify(types.WebhookInboundRequest{
		Request: req,
		Payload: json.RawMessage(`{}`),
	})
	require.ErrorIs(t, err, ErrWebhookSecretMissing)
}

func TestParseEventTime_RFC3339(t *testing.T) {
	t.Parallel()

	result := parseEventTime("2025-01-15T10:30:00Z")
	require.NotNil(t, result)
	assert.Equal(t, 2025, result.Year())
	assert.Equal(t, 10, result.Hour())
}

func TestParseEventTime_RFC3339Nano(t *testing.T) {
	t.Parallel()

	result := parseEventTime("2025-01-15T10:30:00.123456789Z")
	require.NotNil(t, result)
	assert.Equal(t, 123456789, result.Nanosecond())
}

func TestParseEventTime_Empty(t *testing.T) {
	t.Parallel()

	assert.Nil(t, parseEventTime(""))
}

func TestParseEventTime_Invalid(t *testing.T) {
	t.Parallel()

	assert.Nil(t, parseEventTime("not-a-date"))
}

func TestResendWebhookEvent_ToDeliveryEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		eventType string
		kind      deliveryEventKind
	}{
		{eventType: "email.sent", kind: deliveryEventSent},
		{eventType: "email.delivered", kind: deliveryEventDelivered},
		{eventType: "email.opened", kind: deliveryEventOpened},
		{eventType: "email.clicked", kind: deliveryEventClicked},
		{eventType: "email.bounced", kind: deliveryEventBounced},
		{eventType: "email.failed", kind: deliveryEventFailed},
	}

	for _, tc := range tests {
		t.Run(tc.eventType, func(t *testing.T) {
			t.Parallel()

			tags := map[string]string{TagCampaignTargetID: "ct_1", TagIsTest: "true"}
			event := resendWebhookEvent{
				Type: tc.eventType,
				Data: resendWebhookData{EmailID: "e_123", CreatedAt: "2026-08-10T14:39:49.381Z", Tags: tags},
			}

			delivery, ok := event.toDeliveryEvent()
			require.True(t, ok)
			assert.Equal(t, providerResend, delivery.provider)
			assert.Equal(t, tc.eventType, delivery.name)
			assert.Equal(t, tc.kind, delivery.kind)
			assert.Equal(t, "e_123", delivery.emailID)
			assert.Equal(t, "2026-08-10T14:39:49.381Z", delivery.timestamp)
			require.NotNil(t, delivery.occurredAt)
			assert.Equal(t, 381000000, delivery.occurredAt.Nanosecond())
			assert.Equal(t, tags, delivery.tags)
		})
	}
}

func TestResendWebhookEvent_ToDeliveryEventUnhandled(t *testing.T) {
	t.Parallel()

	_, ok := resendWebhookEvent{Type: "email.delivery_delayed"}.toDeliveryEvent()
	assert.False(t, ok)

	_, ok = resendWebhookEvent{}.toDeliveryEvent()
	assert.False(t, ok)
}

func TestGetTargetMetadata_KeysByKind(t *testing.T) {
	t.Parallel()

	existing := map[string]any{"sent": map[string]string{"email_id": "e_1", "timestamp": "t1"}}

	metadata := getTargetMetadata(existing, deliveryEvent{
		provider:  providerResend,
		name:      "email.delivered",
		kind:      deliveryEventDelivered,
		emailID:   "e_2",
		timestamp: "t2",
	})

	assert.Equal(t, map[string]any{
		"sent":      map[string]string{"email_id": "e_1", "timestamp": "t1"},
		"delivered": map[string]string{"email_id": "e_2", "timestamp": "t2"},
	}, metadata)
}

func TestResendDeliveryEventHandle_EmptyType(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"type":"","data":{"email_id":"e_123","created_at":"","tags":{}}}`)
	err := ResendDeliveryEvent{}.Handle(t.Context(), types.WebhookHandleRequest{
		Event: types.WebhookReceivedEvent{
			Payload: payload,
		},
	})
	require.NoError(t, err)
}

func TestResendDeliveryEventHandle_InvalidJSON(t *testing.T) {
	t.Parallel()

	err := ResendDeliveryEvent{}.Handle(t.Context(), types.WebhookHandleRequest{
		Event: types.WebhookReceivedEvent{
			Payload: json.RawMessage(`not json`),
		},
	})
	require.ErrorIs(t, err, ErrWebhookPayloadInvalid)
}

// signedResendRequest builds a Resend webhook request whose Svix headers are signed with secret at the given time
func signedResendRequest(secret string, at time.Time, payload []byte) *http.Request {
	timestamp := strconv.FormatInt(at.Unix(), 10)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("svix-id", "msg_abc")
	req.Header.Set("svix-timestamp", timestamp)
	req.Header.Set("svix-signature", SvixSignature(secret, "msg_abc", timestamp, payload))

	return req
}

func TestResendWebhookVerify_ValidSignature(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"type":"email.delivered","data":{}}`)

	err := ResendWebhook{Secret: MockResendWebhookSecret}.Verify(types.WebhookInboundRequest{
		Request: signedResendRequest(MockResendWebhookSecret, time.Now(), payload),
		Payload: payload,
	})
	require.NoError(t, err)
}

func TestResendWebhookVerify_SignatureMismatch(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"type":"email.delivered","data":{}}`)

	err := ResendWebhook{Secret: MockResendWebhookSecret}.Verify(types.WebhookInboundRequest{
		Request: signedResendRequest("whsec_d3Jvbmctc2VjcmV0", time.Now(), payload),
		Payload: payload,
	})
	require.Error(t, err)
}

func TestResendWebhookVerify_TamperedPayload(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"type":"email.delivered","data":{}}`)

	err := ResendWebhook{Secret: MockResendWebhookSecret}.Verify(types.WebhookInboundRequest{
		Request: signedResendRequest(MockResendWebhookSecret, time.Now(), payload),
		Payload: []byte(`{"type":"email.bounced","data":{}}`),
	})
	require.Error(t, err)
}

func TestResendWebhookVerify_StaleTimestamp(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"type":"email.delivered","data":{}}`)

	err := ResendWebhook{Secret: MockResendWebhookSecret}.Verify(types.WebhookInboundRequest{
		Request: signedResendRequest(MockResendWebhookSecret, time.Now().Add(-10*time.Minute), payload),
		Payload: payload,
	})
	require.Error(t, err)
}
