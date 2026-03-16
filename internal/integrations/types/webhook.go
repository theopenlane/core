package types

import (
	"context"
	"net/http"
)

// WebhookVerifyFunc verifies authenticity of one inbound webhook request
type WebhookVerifyFunc func(ctx context.Context, request *http.Request) error

// WebhookRegistration declares one inbound webhook contract for a definition
type WebhookRegistration struct {
	// Name is the stable webhook identifier within the definition
	Name string `json:"name"`
	// Verify authenticates the inbound webhook request
	Verify WebhookVerifyFunc `json:"-"`
}
