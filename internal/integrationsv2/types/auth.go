package types

import (
	"context"
	"encoding/json"
)

// AuthStartResult captures the output of an auth start function
type AuthStartResult struct {
	// URL is the third-party URL the user should be sent to
	URL string `json:"url,omitempty"`
	// State is the opaque callback state payload persisted for the auth flow
	State json.RawMessage `json:"state,omitempty"`
}

// AuthCompleteResult captures the output of an auth completion function
type AuthCompleteResult struct {
	// Credential is the credential material produced by the auth flow
	Credential CredentialSet `json:"credential"`
	// State is the installation state produced by the auth flow
	State json.RawMessage `json:"state,omitempty"`
}

// AuthStartFunc starts an auth flow for one definition
type AuthStartFunc func(ctx context.Context, input json.RawMessage) (AuthStartResult, error)

// AuthCompleteFunc completes an auth flow for one definition
type AuthCompleteFunc func(ctx context.Context, state json.RawMessage, input json.RawMessage) (AuthCompleteResult, error)
