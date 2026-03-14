package types

import (
	"context"
	"net/http"

	generated "github.com/theopenlane/core/internal/ent/generated"
)

// WebhookVerifyFunc verifies authenticity of one inbound webhook request
type WebhookVerifyFunc func(ctx context.Context, request *http.Request) error

// WebhookResolveFunc resolves one inbound webhook request to an integration record
type WebhookResolveFunc func(ctx context.Context, request *http.Request) (*generated.Integration, error)

// WebhookHandleFunc processes one inbound webhook for one integration record
type WebhookHandleFunc func(ctx context.Context, request *http.Request, integration *generated.Integration) error

// WebhookRegistration declares one inbound webhook contract for a definition
type WebhookRegistration struct {
	// Name is the stable webhook identifier within the definition
	Name string `json:"name"`
	// Verify authenticates the inbound webhook request
	Verify WebhookVerifyFunc `json:"-"`
	// Resolve finds the integration record targeted by the webhook
	Resolve WebhookResolveFunc `json:"-"`
	// Handle processes the verified webhook request
	Handle WebhookHandleFunc `json:"-"`
}
