// Package oauth defines provider specifications and token exchange helpers.
package oauth

import (
	"context"
	"time"
)

// Spec describes an OAuth capable integration provider.
type Spec struct {
	Name         string
	DisplayName  string
	AuthURL      string
	TokenURL     string
	Scopes       []string
	UsePKCE      bool
	RedirectURIs []string
	Metadata     map[string]any
}

// ExchangeOptions conveys runtime overrides for token exchange flows.
type ExchangeOptions struct {
	CodeVerifier  string
	CodeChallenge string
	RedirectURI   string
	Scopes        []string
}

// Token captures the result of an OAuth exchange.
type Token struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
	TokenType    string
	Raw          map[string]any
}

// Exchanger handles the code -> token exchange for a provider spec.
type Exchanger interface {
	Exchange(ctx context.Context, spec Spec, code string, opts ExchangeOptions) (Token, error)
	Refresh(ctx context.Context, spec Spec, refreshToken string) (Token, error)
}

// Validator runs metadata validation before activation persists.
type Validator interface {
	ValidateActivation(ctx context.Context, spec Spec, metadata map[string]any) error
}
