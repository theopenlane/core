package email

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const (
	testPostmarkSecret  = "pm_webhook_password"
	testPostmarkTraceID = "8f4b0c1e-3a2d-4b7e-9c61-2d5e7f0a1b3c"
)

var (
	testPostmarkDeliveryPayload = json.RawMessage(`{
  "RecordType": "Delivery",
  "MessageStream": "outbound",
  "ServerID": 23,
  "MessageID": "00000000-0000-0000-0000-000000000000",
  "Recipient": "john@example.com",
  "Tag": "welcome-email",
  "DeliveredAt": "2026-11-05T16:33:54.9070259Z",
  "Details": "Test delivery webhook details",
  "Metadata": {
    "campaign_target_id": "01KZP1H01G9J21QT2XHSZ5BR00",
    "assessment_resp_id": "01KZP1H01G9J21QT2XHSZ5BR98",
    "is_test": "false"
  }
}`)

	testPostmarkBouncePayload = json.RawMessage(`{
  "RecordType": "Bounce",
  "MessageStream": "outbound",
  "ID": 4323372036854775807,
  "Type": "HardBounce",
  "TypeCode": 1,
  "Name": "Hard bounce",
  "Tag": "Test",
  "MessageID": "883953f4-6105-42a2-a16a-77a8eac79483",
  "Metadata": {
    "campaign_target_id": "01KZP1H01G9J21QT2XHSZ5BR00"
  },
  "ServerID": 23,
  "Description": "The server was unable to deliver your message (ex: unknown user, mailbox not found).",
  "Details": "Test bounce details",
  "Email": "john@example.com",
  "From": "sender@example.com",
  "BouncedAt": "2026-11-05T16:33:54.9070259Z",
  "DumpAvailable": true,
  "Inactive": true,
  "CanActivate": true,
  "Subject": "Test subject",
  "Content": "Test content"
}`)

	testPostmarkSpamComplaintPayload = json.RawMessage(`{
  "RecordType": "SpamComplaint",
  "MessageStream": "outbound",
  "ID": 42,
  "Type": "SpamComplaint",
  "TypeCode": 100001,
  "Name": "Spam complaint",
  "Tag": "Test",
  "MessageID": "00000000-0000-0000-0000-000000000000",
  "Metadata": {
    "campaign_target_id": "01KZP1H01G9J21QT2XHSZ5BR00"
  },
  "ServerID": 1234,
  "Description": "",
  "Details": "Test spam complaint details",
  "Email": "john@example.com",
  "From": "sender@example.com",
  "BouncedAt": "2026-11-05T16:33:54.9070259Z",
  "DumpAvailable": true,
  "Inactive": true,
  "CanActivate": false,
  "Subject": "Test subject",
  "Content": "<Abuse report dump content>"
}`)

	testPostmarkOpenPayload = json.RawMessage(`{
  "RecordType": "Open",
  "MessageStream": "outbound",
  "FirstOpen": true,
  "Client": {"Name": "Chrome 35.0.1916.153", "Company": "Google", "Family": "Chrome"},
  "OS": {"Name": "OS X 10.7 Lion", "Company": "Apple Computer, Inc.", "Family": "OS X 10"},
  "Platform": "WebMail",
  "UserAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5)",
  "ReadSeconds": 5,
  "Geo": {"CountryISOCode": "RS", "Country": "Serbia"},
  "MessageID": "00000000-0000-0000-0000-000000000000",
  "Metadata": {
    "campaign_target_id": "01KZP1H01G9J21QT2XHSZ5BR00"
  },
  "ReceivedAt": "2026-11-05T16:33:54.9070259Z",
  "Tag": "welcome-email",
  "Recipient": "john@example.com"
}`)

	testPostmarkClickPayload = json.RawMessage(`{
  "RecordType": "Click",
  "MessageStream": "outbound",
  "ClickLocation": "HTML",
  "Client": {"Name": "Chrome 35.0.1916.153", "Company": "Google", "Family": "Chrome"},
  "OS": {"Name": "OS X 10.7 Lion", "Company": "Apple Computer, Inc.", "Family": "OS X 10"},
  "Platform": "Desktop",
  "UserAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5)",
  "OriginalLink": "https://example.com",
  "Geo": {"CountryISOCode": "RS", "Country": "Serbia"},
  "MessageID": "00000000-0000-0000-0000-000000000000",
  "Metadata": {
    "campaign_target_id": "01KZP1H01G9J21QT2XHSZ5BR00"
  },
  "ReceivedAt": "2026-11-05T16:33:54.9070259Z",
  "Tag": "welcome-email",
  "Recipient": "john@example.com"
}`)

	testPostmarkSubscriptionChangePayload = json.RawMessage(`{
  "RecordType": "SubscriptionChange",
  "MessageID": null,
  "ServerID": 23,
  "MessageStream": "outbound",
  "ChangedAt": "2026-11-05T16:33:54.9070259Z",
  "Recipient": "john@example.com",
  "Origin": "Recipient",
  "SuppressSending": true,
  "SuppressionReason": "HardBounce",
  "Tag": "welcome-email",
  "Metadata": {
    "campaign_target_id": "01KZP1H01G9J21QT2XHSZ5BR00"
  }
}`)
)

func TestPostmarkWebhookVerify_MissingSecret(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.SetBasicAuth("postmark", testPostmarkSecret)

	err := PostmarkWebhook{Secret: ""}.Verify(types.WebhookInboundRequest{
		Request: req,
		Payload: testPostmarkDeliveryPayload,
	})
	require.ErrorIs(t, err, ErrWebhookSecretMissing)
}

func TestPostmarkWebhookVerify_MissingBasicAuth(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", nil)

	err := PostmarkWebhook{Secret: testPostmarkSecret}.Verify(types.WebhookInboundRequest{
		Request: req,
		Payload: testPostmarkDeliveryPayload,
	})
	require.ErrorIs(t, err, ErrWebhookAuthMissing)
}

func TestPostmarkWebhookVerify_WrongPassword(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.SetBasicAuth("postmark", "wrong-password")

	err := PostmarkWebhook{Secret: testPostmarkSecret}.Verify(types.WebhookInboundRequest{
		Request: req,
		Payload: testPostmarkDeliveryPayload,
	})
	require.ErrorIs(t, err, ErrWebhookAuthInvalid)
}

func TestPostmarkWebhookVerify_CorrectPassword(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.SetBasicAuth("any-user", testPostmarkSecret)

	err := PostmarkWebhook{Secret: testPostmarkSecret}.Verify(types.WebhookInboundRequest{
		Request: req,
		Payload: testPostmarkDeliveryPayload,
	})
	require.NoError(t, err)
}

func TestPostmarkWebhookEvent_HandledRecordTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		recordType string
		payload    json.RawMessage
	}{
		{recordType: "Delivery", payload: testPostmarkDeliveryPayload},
		{recordType: "Bounce", payload: testPostmarkBouncePayload},
		{recordType: "SpamComplaint", payload: testPostmarkSpamComplaintPayload},
		{recordType: "Open", payload: testPostmarkOpenPayload},
		{recordType: "Click", payload: testPostmarkClickPayload},
	}

	for _, tc := range tests {
		t.Run(tc.recordType, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set("X-PM-Webhook-Trace-Id", testPostmarkTraceID)

			event, err := PostmarkWebhook{}.Event(types.WebhookInboundRequest{
				Request: req,
				Payload: tc.payload,
			})
			require.NoError(t, err)
			assert.Equal(t, tc.recordType, event.Name)
			assert.Equal(t, testPostmarkTraceID, event.DeliveryID)
			assert.Equal(t, tc.payload, event.Payload)
		})
	}
}

func TestPostmarkWebhookEvent_SubscriptionChangeIgnored(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-PM-Webhook-Trace-Id", testPostmarkTraceID)

	event, err := PostmarkWebhook{}.Event(types.WebhookInboundRequest{
		Request: req,
		Payload: testPostmarkSubscriptionChangePayload,
	})
	require.NoError(t, err)
	assert.Empty(t, event.Name)
}

func TestPostmarkWebhookEvent_InvalidJSON(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", nil)

	_, err := PostmarkWebhook{}.Event(types.WebhookInboundRequest{
		Request: req,
		Payload: json.RawMessage(`not json`),
	})
	require.ErrorIs(t, err, ErrWebhookPayloadInvalid)
}

func TestPostmarkDeliveryEventHandle_EmptyRecordType(t *testing.T) {
	t.Parallel()

	err := PostmarkDeliveryEvent{}.Handle(t.Context(), types.WebhookHandleRequest{
		Event: types.WebhookReceivedEvent{
			Payload: json.RawMessage(`{"RecordType":"","MessageID":"00000000-0000-0000-0000-000000000000","Metadata":{}}`),
		},
	})
	require.NoError(t, err)
}

func TestPostmarkDeliveryEventHandle_InvalidJSON(t *testing.T) {
	t.Parallel()

	err := PostmarkDeliveryEvent{}.Handle(t.Context(), types.WebhookHandleRequest{
		Event: types.WebhookReceivedEvent{
			Payload: json.RawMessage(`not json`),
		},
	})
	require.ErrorIs(t, err, ErrWebhookPayloadInvalid)
}

func TestPostmarkWebhookEvent_ToDeliveryEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		recordType string
		payload    json.RawMessage
		kind       deliveryEventKind
		emailID    string
	}{
		{recordType: "Delivery", payload: testPostmarkDeliveryPayload, kind: deliveryEventDelivered, emailID: "00000000-0000-0000-0000-000000000000"},
		{recordType: "Bounce", payload: testPostmarkBouncePayload, kind: deliveryEventBounced, emailID: "883953f4-6105-42a2-a16a-77a8eac79483"},
		{recordType: "SpamComplaint", payload: testPostmarkSpamComplaintPayload, kind: deliveryEventComplained, emailID: "00000000-0000-0000-0000-000000000000"},
		{recordType: "Open", payload: testPostmarkOpenPayload, kind: deliveryEventOpened, emailID: "00000000-0000-0000-0000-000000000000"},
		{recordType: "Click", payload: testPostmarkClickPayload, kind: deliveryEventClicked, emailID: "00000000-0000-0000-0000-000000000000"},
	}

	for _, tc := range tests {
		t.Run(tc.recordType, func(t *testing.T) {
			t.Parallel()

			var event postmarkWebhookEvent
			require.NoError(t, json.Unmarshal(tc.payload, &event))

			delivery, ok := event.toDeliveryEvent()
			require.True(t, ok)
			assert.Equal(t, providerPostmark, delivery.provider)
			assert.Equal(t, tc.recordType, delivery.name)
			assert.Equal(t, tc.kind, delivery.kind)
			assert.Equal(t, tc.emailID, delivery.emailID)
			assert.Equal(t, "2026-11-05T16:33:54.9070259Z", delivery.timestamp)
			require.NotNil(t, delivery.occurredAt)
			assert.Equal(t, 907025900, delivery.occurredAt.Nanosecond())
			assert.Equal(t, "01KZP1H01G9J21QT2XHSZ5BR00", delivery.tags[TagCampaignTargetID])
		})
	}
}

func TestPostmarkWebhookEvent_ToDeliveryEventDeliveryTags(t *testing.T) {
	t.Parallel()

	var event postmarkWebhookEvent
	require.NoError(t, json.Unmarshal(testPostmarkDeliveryPayload, &event))

	delivery, ok := event.toDeliveryEvent()
	require.True(t, ok)
	assert.Equal(t, "01KZP1H01G9J21QT2XHSZ5BR98", delivery.tags[TagAssessmentResponseID])
	assert.Equal(t, "false", delivery.tags[TagIsTest])
}

func TestPostmarkWebhookEvent_ToDeliveryEventUnhandled(t *testing.T) {
	t.Parallel()

	var event postmarkWebhookEvent
	require.NoError(t, json.Unmarshal(testPostmarkSubscriptionChangePayload, &event))

	_, ok := event.toDeliveryEvent()
	assert.False(t, ok)
}
