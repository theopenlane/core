package auth

import (
	"context"
	"maps"
	"net/url"
	"strings"

	"github.com/theopenlane/core/internal/integrations/types"
)

// AuthenticatedClient wraps a bearer token and headers for simple HTTP JSON calls
type AuthenticatedClient struct {
	// BaseURL is an optional base URL prepended to relative paths.
	BaseURL string
	// BearerToken is the optional bearer token for Authorization headers
	BearerToken string
	// Headers contains additional static headers for each request
	Headers map[string]string
}

// NewAuthenticatedClient builds an AuthenticatedClient with cloned headers.
func NewAuthenticatedClient(baseURL, bearerToken string, headers map[string]string) *AuthenticatedClient {
	return &AuthenticatedClient{
		BaseURL:     strings.TrimSpace(baseURL),
		BearerToken: bearerToken,
		Headers:     maps.Clone(headers),
	}
}

// GetJSON issues a GET request using the stored credentials and decodes the JSON response.
func (c *AuthenticatedClient) GetJSON(ctx context.Context, path string, out any) error {
	endpoint := buildEndpointURL(c.BaseURL, path, nil)
	return HTTPGetJSON(ctx, nil, endpoint, c.BearerToken, c.Headers, out)
}

// GetJSONWithParams issues a GET request with query parameters.
func (c *AuthenticatedClient) GetJSONWithParams(ctx context.Context, path string, params url.Values, out any) error {
	endpoint := buildEndpointURL(c.BaseURL, path, params)
	return HTTPGetJSON(ctx, nil, endpoint, c.BearerToken, c.Headers, out)
}

// PostJSON issues a POST request using the stored credentials and decodes the JSON response.
func (c *AuthenticatedClient) PostJSON(ctx context.Context, path string, body, out any) error {
	endpoint := buildEndpointURL(c.BaseURL, path, nil)
	return HTTPPostJSON(ctx, nil, endpoint, c.BearerToken, c.Headers, body, out)
}

// clone returns a deep copy of the client fields for safe mutation.
func (c *AuthenticatedClient) clone() *AuthenticatedClient {
	if c == nil {
		return nil
	}

	return &AuthenticatedClient{
		BaseURL:     c.BaseURL,
		BearerToken: c.BearerToken,
		Headers:     maps.Clone(c.Headers),
	}
}

// AuthenticatedClientFromClient attempts to unwrap an AuthenticatedClient from a wrapped client value
func AuthenticatedClientFromClient(value types.ClientInstance) *AuthenticatedClient {
	client, ok := types.ClientInstanceAs[*AuthenticatedClient](value)
	if !ok {
		return nil
	}

	return client
}

func buildEndpointURL(baseURL, path string, params url.Values) string {
	trimmedPath := strings.TrimSpace(path)
	switch {
	case strings.HasPrefix(trimmedPath, "http://"), strings.HasPrefix(trimmedPath, "https://"):
		if len(params) == 0 {
			return trimmedPath
		}

		encoded := params.Encode()
		if encoded == "" {
			return trimmedPath
		}
		if strings.Contains(trimmedPath, "?") {
			return trimmedPath + "&" + encoded
		}

		return trimmedPath + "?" + encoded
	case strings.TrimSpace(baseURL) == "":
		if len(params) == 0 {
			return trimmedPath
		}

		encoded := params.Encode()
		if encoded == "" {
			return trimmedPath
		}

		return trimmedPath + "?" + encoded
	default:
		endpoint := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(trimmedPath, "/")
		if len(params) == 0 {
			return endpoint
		}

		encoded := params.Encode()
		if encoded == "" {
			return endpoint
		}

		return endpoint + "?" + encoded
	}
}
