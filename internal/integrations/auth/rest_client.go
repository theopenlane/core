package auth

import (
	"context"
	"net/url"
	"strings"
)

// RESTClient provides base-URL-relative JSON HTTP helpers for provider operations.
type RESTClient struct {
	// BaseURL is the base URL prepended to every path.
	BaseURL string
	// DefaultHeaders are static headers applied to every token-authenticated request.
	// When a pooled AuthenticatedClient is supplied these headers are ignored because
	// the client's own transport handles authentication.
	DefaultHeaders map[string]string
}

// GetJSON assembles an endpoint URL from the base URL, path, and optional query parameters,
// then performs an authenticated JSON GET request.
func (c RESTClient) GetJSON(ctx context.Context, client *AuthenticatedClient, token, path string, params url.Values, out any) error {
	endpoint := buildEndpointURL(c.BaseURL, path, params)
	return GetJSONWithClient(ctx, client, endpoint, token, c.DefaultHeaders, out)
}

// PostJSON assembles an endpoint URL from the base URL and path, then performs an
// authenticated JSON POST request. When client is non-nil it is used for the request;
// otherwise the bearer token and default headers are used.
func (c RESTClient) PostJSON(ctx context.Context, client *AuthenticatedClient, token, path string, body, out any) error {
	endpoint := buildEndpointURL(c.BaseURL, path, nil)
	return PostJSONWithClient(ctx, client, endpoint, token, c.DefaultHeaders, body, out)
}

// buildEndpointURL constructs a full URL by joining baseURL and path with a single slash
// and appending URL-encoded query parameters when present.
func buildEndpointURL(baseURL, path string, params url.Values) string {
	endpoint := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
	if params != nil {
		if encoded := params.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}
	}

	return endpoint
}
