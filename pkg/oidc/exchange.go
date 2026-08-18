package oidc

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zitadel/oidc/v3/pkg/client/tokenexchange"
	zoidc "github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"

	"github.com/theopenlane/iam/tokens"
)

const defaultAssertionLifetime = 10 * time.Minute

// FederationSource configures the assertion exchange against one RFC 8693 security token service
type FederationSource struct {
	// Manager signs the assertions with the platform signing keys
	Manager *tokens.TokenManager
	// OrganizationID is the organization bound into every assertion's identity claims
	OrganizationID string
	// Audience is the relying party identity the assertion is exchanged against
	Audience string
	// Scopes are the OAuth scopes requested during the exchange
	Scopes []string
	// Endpoint is the RFC 8693 token exchange endpoint
	Endpoint string
	// AssertionOptions customize the minted assertions; identity options always win
	AssertionOptions []tokens.ConfigOpt
}

// NewTokenSource returns a caching token source that mints assertions and exchanges them at the configured endpoint
func NewTokenSource(ctx context.Context, source FederationSource) (oauth2.TokenSource, error) {
	if source.OrganizationID == "" {
		return nil, ErrOrganizationIDRequired
	}

	if source.Audience == "" {
		return nil, ErrAudienceRequired
	}

	if source.Endpoint == "" {
		return nil, ErrEndpointRequired
	}

	metadata, err := source.Manager.GetKeyMetadata(source.Manager.CurrentKeyID())
	if err != nil {
		return nil, err
	}

	if metadata.Algorithm != jwt.SigningMethodRS256.Alg() {
		return nil, tokens.ErrSigningAlgorithmMismatch
	}

	return oauth2.ReuseTokenSource(nil, &federationTokenSource{
		// the token source outlives the request-scoped context in pooled clients
		ctx:    context.WithoutCancel(ctx),
		source: source,
	}), nil
}

// federationTokenSource mints assertions and exchanges them for provider access tokens
type federationTokenSource struct {
	// ctx carries caller values without cancellation since Token takes no context
	ctx context.Context
	// source holds the organization, audience, scopes, and endpoint for the exchange
	source FederationSource
}

// Token mints a fresh assertion and exchanges it for a provider access token
func (s *federationTokenSource) Token() (*oauth2.Token, error) {
	opts := append([]tokens.ConfigOpt{}, s.source.AssertionOptions...)
	opts = append(opts,
		tokens.WithSubject(s.source.OrganizationID),
		tokens.WithAudience(s.source.Audience),
		tokens.WithClaim("organization_id", s.source.OrganizationID),
		tokens.WithAccessDuration(defaultAssertionLifetime),
		tokens.WithRequiredAlgorithm(jwt.SigningMethodRS256.Alg()),
	)

	assertion, err := s.source.Manager.CreateSignedToken(opts...)
	if err != nil {
		return nil, err
	}

	exchanger, err := tokenexchange.NewTokenExchanger(s.ctx, "", tokenexchange.WithStaticTokenEndpoint("", s.source.Endpoint))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrExchangeFailed, err)
	}

	resp, err := tokenexchange.ExchangeToken(s.ctx, exchanger, assertion, zoidc.JWTTokenType, "", "", nil, []string{s.source.Audience}, s.source.Scopes, zoidc.AccessTokenType)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrExchangeFailed, err)
	}

	expiresIn := resp.ExpiresIn
	if expiresIn > math.MaxInt32 {
		expiresIn = math.MaxInt32
	}

	return &oauth2.Token{
		AccessToken: resp.AccessToken,
		TokenType:   resp.TokenType,
		Expiry:      time.Now().Add(time.Duration(expiresIn) * time.Second),
	}, nil
}
