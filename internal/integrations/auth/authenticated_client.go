package auth

import (
	"context"
	"maps"

	"github.com/theopenlane/core/internal/integrations/types"
)

// AuthenticatedClient wraps a bearer token and headers for simple HTTP JSON calls
type AuthenticatedClient struct {
	// BearerToken is the optional bearer token for Authorization headers
	BearerToken string
	// Headers contains additional static headers for each request
	Headers map[string]string
}

// NewAuthenticatedClient builds an AuthenticatedClient with a cloned header map
func NewAuthenticatedClient(bearerToken string, headers map[string]string) *AuthenticatedClient {
	return &AuthenticatedClient{
		BearerToken: bearerToken,
		Headers:     maps.Clone(headers),
	}
}

// GetJSON issues a GET request using the stored credentials and decodes the JSON response
func (c *AuthenticatedClient) GetJSON(ctx context.Context, endpoint string, out any) error {
	return HTTPGetJSON(ctx, nil, endpoint, c.BearerToken, c.Headers, out)
}

// GetJSONWithClient uses the authenticated client when available, otherwise falls back to HTTPGetJSON
func GetJSONWithClient(ctx context.Context, client *AuthenticatedClient, endpoint string, bearer string, headers map[string]string, out any) error {
	if client != nil {
		return client.GetJSON(ctx, endpoint, out)
	}

	return HTTPGetJSON(ctx, nil, endpoint, bearer, headers, out)
}

// AuthenticatedClientFromClient attempts to unwrap an AuthenticatedClient from a wrapped client value
func AuthenticatedClientFromClient(value types.ClientInstance) *AuthenticatedClient {
	client, ok := types.ClientInstanceAs[*AuthenticatedClient](value)
	if !ok {
		return nil
	}

	return client
}
