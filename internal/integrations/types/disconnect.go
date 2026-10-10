package types //nolint:revive

import (
	"context"
	"encoding/json"
)

// DisconnectResult captures the output of a disconnect flow
type DisconnectResult struct {
	// RedirectURL is a provider URL the user should visit to complete external teardown
	RedirectURL string `json:"redirectUrl,omitempty"`
	// Message is a user-facing summary of what was cleaned up or what action is required
	Message string `json:"message,omitempty"`
	// Details is an opaque provider-specific payload describing teardown actions taken or required
	Details json.RawMessage `json:"details,omitempty"`
	// SkipLocalCleanup indicates the runtime should not delete credentials and the installation
	// record, set when teardown is deferred to an external event such as a provider webhook
	SkipLocalCleanup bool `json:"-"`
}

// DisconnectFunc executes provider-specific disconnect logic for one installation
type DisconnectFunc func(ctx context.Context, input ConnectionInput) (DisconnectResult, error)

// DisconnectRegistration describes how one definition handles installation teardown
type DisconnectRegistration struct {
	// Description is the user-facing explanation of what disconnect does and any recommended provider-side cleanup
	Description string `json:"description,omitempty"`
	// Schema is the JSON schema describing the disconnect result details payload
	Schema json.RawMessage `json:"schema,omitempty"`
	// Disconnect executes provider-specific teardown logic
	Disconnect DisconnectFunc `json:"-"`
}
