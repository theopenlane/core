package types

import (
	"context"
	"encoding/json"
	"time"
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
	// InstallationInput is optional installation-scoped input captured during auth completion
	InstallationInput json.RawMessage `json:"installationInput,omitempty"`
}

// TokenView is a read-only view of a credential's active token and expiry
type TokenView struct {
	// AccessToken is the active access token
	AccessToken string `json:"accessToken,omitempty"`
	// ExpiresAt is the token expiry timestamp when available
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
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

// Values returns all query parameter values for the supplied name
func (i AuthCallbackInput) Values(name string) []string {
	for _, value := range i.Query {
		if value.Name == name {
			out := make([]string, len(value.Values))
			copy(out, value.Values)
			return out
		}
	}

	return nil
}

// AuthStartFunc initiates an auth flow and returns the redirect URL and opaque state
type AuthStartFunc func(ctx context.Context, input json.RawMessage) (AuthStartResult, error)

// AuthCompleteFunc finalizes an auth flow and returns the resulting credential
type AuthCompleteFunc func(ctx context.Context, state json.RawMessage, input AuthCallbackInput) (AuthCompleteResult, error)

// AuthTokenViewFunc produces a read-only token view from a persisted credential
type AuthTokenViewFunc func(ctx context.Context, credential CredentialSet) (*TokenView, error)

// AuthRegistration describes how one connection mode starts and completes auth
type AuthRegistration struct {
	// CredentialRef identifies which credential slot receives the auth result
	CredentialRef CredentialSlotID `json:"credentialRef"`
	// Start initiates the auth flow
	Start AuthStartFunc `json:"-"`
	// Complete finalizes the auth flow and returns the resulting credential
	Complete AuthCompleteFunc `json:"-"`
	// TokenView produces a read-only token view from a persisted credential
	TokenView AuthTokenViewFunc `json:"-"`
}
