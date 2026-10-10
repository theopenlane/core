package types //nolint:revive

import (
	"context"
	"encoding/json"
)

// AuthStartFunc initiates an auth flow and returns the redirect URL and opaque state
type AuthStartFunc func(ctx context.Context, input json.RawMessage) (AuthStartResult, error)

// AuthCompleteFunc finalizes an auth flow and returns the resulting credential
type AuthCompleteFunc func(ctx context.Context, state json.RawMessage, input AuthCallbackInput) (AuthCompleteResult, error)

// AuthRegistration describes how one connection mode starts and completes auth
type AuthRegistration struct {
	// Start initiates the auth flow
	Start AuthStartFunc `json:"-"`
	// Complete finalizes the auth flow and returns the resulting credential
	Complete AuthCompleteFunc `json:"-"`
}

// AuthFlow is an auth registration whose completed credential is T
type AuthFlow[T any] struct {
	// registration is the erased auth registration
	registration AuthRegistration
}

// NewAuthFlow creates a typed auth flow from its start and complete functions
func NewAuthFlow[T any](start AuthStartFunc, complete AuthCompleteFunc) AuthFlow[T] {
	return AuthFlow[T]{registration: AuthRegistration{Start: start, Complete: complete}}
}

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
}

// AuthCallbackValue captures one callback parameter and its values
type AuthCallbackValue struct {
	// Name is the callback parameter name
	Name string `json:"name"`
	// Values are the values supplied for the callback parameter
	Values []string `json:"values,omitempty"`
}

// AuthCallbackInput captures the provider callback payload in a typed JSON-friendly shape
type AuthCallbackInput struct {
	// Query lists the query parameters supplied on the callback request
	Query []AuthCallbackValue `json:"query,omitempty"`
}

// First returns the first query parameter value for the supplied name
func (i AuthCallbackInput) First(name string) string {
	for _, value := range i.Query {
		if value.Name == name && len(value.Values) > 0 {
			return value.Values[0]
		}
	}

	return ""
}
