package auth

import (
	"context"
	"maps"
	"strings"
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
	cloned := cloneHeaders(headers)
	return &AuthenticatedClient{
		BearerToken: strings.TrimSpace(bearerToken),
		Headers:     cloned,
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

// AuthenticatedClientFromAny attempts to unwrap an AuthenticatedClient from an arbitrary value
func AuthenticatedClientFromAny(value any) *AuthenticatedClient {
	client, ok := value.(*AuthenticatedClient)
	if !ok {
		return nil
	}

	return client
}

// cloneHeaders creates a shallow copy of the header map
func cloneHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	return maps.Clone(headers)
}
