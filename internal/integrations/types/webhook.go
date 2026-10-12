package types //nolint:revive

import (
	"context"
	"encoding/json"
	"net/http"

	generated "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// StaticWebhookEntry pairs a definition ID with a webhook registration that declares a static route
type StaticWebhookEntry struct {
	// DefinitionID is the definition that owns the webhook
	DefinitionID string
	// WebhookName is the webhook name within the definition
	WebhookName string
	// StaticRoute is the fixed URL path for the webhook
	StaticRoute string
}

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
	// Event is the normalized webhook event envelope
	Event WebhookReceivedEvent
	// Ingest processes mapped provider payloads directly through the shared ingest pipeline
	Ingest func(context.Context, []IngestPayloadSet) error
	// DispatchOperation queues one integration operation for this installation
	DispatchOperation func(context.Context, string, json.RawMessage) error
	// CleanupInstallation removes the installation and credentials on external teardown
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
	// Replaces lists the retired contract names whose persisted webhook rows move onto this contract
	Replaces []string `json:"-"`
	// EndpointURLTemplate overrides the persisted endpoint URL path
	EndpointURLTemplate string `json:"endpointUrlTemplate,omitempty"`
	// StaticRoute is a fixed URL path registered instead of a per-installation endpoint
	StaticRoute string `json:"staticRoute,omitempty"`
	// SecretSource returns an operator-supplied webhook secret instead of auto-generating one
	SecretSource func() string `json:"-"`
	// ResolveIntegration locates the integration record from the inbound request
	ResolveIntegration func(ctx context.Context, db *generated.Client, req WebhookInboundRequest) (*generated.Integration, error) `json:"-"`
	// Verify authenticates the inbound webhook request
	Verify WebhookVerifyFunc `json:"-"`
	// Event resolves the inbound request into a supported event
	Event WebhookEventFunc `json:"-"`
	// Events lists the supported event handlers for the webhook
	Events []WebhookEventRegistration `json:"events,omitempty"`
}
