package email

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/types"
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
      "assessment_response_id": "01KZP1H01G9J21QT2XHSZ5BR98",
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

func TestParseResendEventTime_RFC3339(t *testing.T) {
	t.Parallel()

	result := parseResendEventTime("2025-01-15T10:30:00Z")
	require.NotNil(t, result)
	assert.Equal(t, 2025, result.Year())
	assert.Equal(t, 10, result.Hour())
}

func TestParseResendEventTime_RFC3339Nano(t *testing.T) {
	t.Parallel()

	result := parseResendEventTime("2025-01-15T10:30:00.123456789Z")
	require.NotNil(t, result)
	assert.Equal(t, 123456789, result.Nanosecond())
}

func TestParseResendEventTime_Empty(t *testing.T) {
	t.Parallel()

	assert.Nil(t, parseResendEventTime(""))
}

func TestParseResendEventTime_Invalid(t *testing.T) {
	t.Parallel()

	assert.Nil(t, parseResendEventTime("not-a-date"))
}

func TestResendDeliveryEventHandle_EmptyType(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"type":"","data":{"email_id":"e_123","created_at":"","tags":[]}}`)
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
