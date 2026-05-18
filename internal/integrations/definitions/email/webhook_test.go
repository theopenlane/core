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

	payload := json.RawMessage(`{"type":"email.delivered","data":{"email_id":"e_123","created_at":"2025-01-01T00:00:00Z","tags":[]}}`)
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

func TestExtractTags(t *testing.T) {
	t.Parallel()

	tags := []resendWebhookTag{
		{Name: "campaign_target_id", Value: "ct_123"},
		{Name: "is_test", Value: "true"},
		{Name: "", Value: "ignored"},
		{Name: "empty_val", Value: ""},
	}

	m := extractTags(tags)
	assert.Equal(t, "ct_123", m["campaign_target_id"])
	assert.Equal(t, "true", m["is_test"])
	assert.NotContains(t, m, "")
	assert.NotContains(t, m, "empty_val")
}

func TestExtractTags_Empty(t *testing.T) {
	t.Parallel()

	m := extractTags(nil)
	assert.Empty(t, m)
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
