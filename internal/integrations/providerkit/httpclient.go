package providerkit

import (
	"context"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
)

const (
	defaultHTTPTimeout = 10 * time.Second
	maxHTTPErrorBody   = 2048
)

var defaultHTTPRequester = httpsling.MustNew(
	httpsling.Client(httpclient.Timeout(defaultHTTPTimeout)),
)

// AuthenticatedClient wraps a bearer token and headers for simple HTTP JSON calls
type AuthenticatedClient struct {
	// BaseURL is an optional base URL prepended to relative paths
	BaseURL string
	// BearerToken is the optional bearer token for Authorization headers
	BearerToken string
	// Headers contains additional static headers for each request
	Headers map[string]string
}

// NewAuthenticatedClient builds an AuthenticatedClient with cloned headers
func NewAuthenticatedClient(baseURL, token string, headers map[string]string) *AuthenticatedClient {
	return &AuthenticatedClient{
		BaseURL:     strings.TrimSpace(baseURL),
		BearerToken: token,
		Headers:     maps.Clone(headers),
	}
}

// GetJSON issues a GET request using the stored credentials and decodes the JSON response
func (c *AuthenticatedClient) GetJSON(ctx context.Context, path string, out any) error {
	endpoint := buildEndpointURL(c.BaseURL, path, nil)
	return httpJSON(ctx, http.MethodGet, endpoint, c.BearerToken, c.Headers, nil, out)
}

// GetJSONWithParams issues a GET request with query parameters and decodes the JSON response
func (c *AuthenticatedClient) GetJSONWithParams(ctx context.Context, path string, params url.Values, out any) error {
	endpoint := buildEndpointURL(c.BaseURL, path, params)
	return httpJSON(ctx, http.MethodGet, endpoint, c.BearerToken, c.Headers, nil, out)
}

// PostJSON issues a POST request using the stored credentials and decodes the JSON response
func (c *AuthenticatedClient) PostJSON(ctx context.Context, path string, body, out any) error {
	endpoint := buildEndpointURL(c.BaseURL, path, nil)
	return httpJSON(ctx, http.MethodPost, endpoint, c.BearerToken, c.Headers, body, out)
}

func httpJSON(ctx context.Context, method string, endpoint string, bearer string, headers map[string]string, body any, out any) error {
	options := []httpsling.Option{
		httpsling.Method(method),
		httpsling.URL(endpoint),
		httpsling.Header(httpsling.HeaderAccept, httpsling.ContentTypeJSON),
	}

	if body != nil {
		options = append(options, httpsling.JSONBody(body))
	}

	if bearer != "" {
		options = append(options, httpsling.BearerAuth(bearer))
	}

	if len(headers) > 0 {
		options = append(options, httpsling.HeadersFromMap(headers))
	}

	resp, err := defaultHTTPRequester.SendWithContext(ctx, options...)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return buildHTTPRequestError(resp, endpoint)
	}

	if out == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func buildHTTPRequestError(resp *http.Response, endpoint string) error {
	body := ""
	if resp.Body != nil {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, maxHTTPErrorBody))
		body = strings.TrimSpace(string(data))
	}

	return &HTTPRequestError{
		URL:        endpoint,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Body:       body,
	}
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
