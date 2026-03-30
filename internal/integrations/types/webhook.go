package types //nolint:revive

import (
	"context"
	"encoding/json"
	"net/http"

	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

// WebhookInboundRequest captures the inputs for verifying and resolving an inbound webhook request
type WebhookInboundRequest struct {
	// Integration is the installed integration receiving the webhook
	Integration *generated.Integration
	// Webhook is the persisted webhook configuration for the installation
	Webhook *generated.IntegrationWebhook
	// Request is the inbound HTTP request
	Request *http.Request
	// Payload is the raw request body
	Payload json.RawMessage
}

// WebhookReceivedEvent is the normalized event emitted into gala for inbound webhooks
type WebhookReceivedEvent struct {
	// Name is the stable event identifier for the definition
	Name string `json:"name"`
	// DeliveryID is the upstream delivery identifier used for idempotency when present
	DeliveryID string `json:"deliveryId,omitempty"`
	// Payload is the raw provider payload
	Payload json.RawMessage `json:"payload"`
	// Headers captures the inbound HTTP headers as normalized strings
	Headers map[string]string `json:"headers,omitempty"`
}

// WebhookHandleRequest captures the inputs required to process one webhook event
type WebhookHandleRequest struct {
	// Integration is the installed integration receiving the event
	Integration *generated.Integration
	// Webhook is the persisted webhook configuration for the installation
	Webhook *generated.IntegrationWebhook
	// DB is the Ent client used for persistence during webhook processing
	DB *generated.Client
	// Event is the normalized webhook event envelope
	Event WebhookReceivedEvent
	// Ingest processes mapped provider payloads directly through the shared ingest pipeline
	Ingest func(context.Context, []IngestPayloadSet) error
	// DispatchOperation queues one integration operation for this installation
	DispatchOperation func(context.Context, string, json.RawMessage) error
	// CleanupInstallation removes the installation and persisted credentials when a provider event
	// represents authoritative external teardown (for example, a GitHub App uninstall webhook)
	CleanupInstallation func(context.Context) error
}

// WebhookVerifyFunc verifies authenticity of one inbound webhook request
type WebhookVerifyFunc func(request WebhookInboundRequest) error

// WebhookEventFunc resolves one inbound webhook request into a registered event
type WebhookEventFunc func(request WebhookInboundRequest) (WebhookReceivedEvent, error)

// WebhookHandleFunc processes one normalized webhook event
type WebhookHandleFunc func(ctx context.Context, request WebhookHandleRequest) error

// WebhookEventRegistration declares one supported inbound event for a definition webhook
type WebhookEventRegistration struct {
	// Name is the stable event identifier within the webhook contract
	Name string `json:"name"`
	// Topic is the gala topic used to dispatch the event
	Topic gala.TopicName `json:"topic"`
	// Ingest declares the ingest contracts supported by this webhook event
	Ingest []IngestContract `json:"ingest,omitempty"`
	// Handle processes the event
	Handle WebhookHandleFunc `json:"-"`
}

// WebhookRegistration declares one inbound webhook contract for a definition
type WebhookRegistration struct {
	// Name is the stable webhook identifier within the definition
	Name string `json:"name"`
	// EndpointURLTemplate overrides the persisted endpoint URL path
	// Use "{endpointID}" as the placeholder for the generated endpoint identifier
	EndpointURLTemplate string `json:"endpointUrlTemplate,omitempty"`
	// Verify authenticates the inbound webhook request
	Verify WebhookVerifyFunc `json:"-"`
	// Event resolves the inbound request into a supported event
	Event WebhookEventFunc `json:"-"`
	// Events lists the supported event handlers for the webhook
	Events []WebhookEventRegistration `json:"events,omitempty"`
}
