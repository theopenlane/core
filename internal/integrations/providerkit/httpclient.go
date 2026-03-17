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
	// defaultHTTPTimeout is the default HTTP client timeout for all outbound requests
	defaultHTTPTimeout = 10 * time.Second
	// maxHTTPErrorBody is the maximum number of bytes read from an error response body
	maxHTTPErrorBody = 2048
)

// defaultHTTPRequester is the shared HTTP requester with the default timeout applied
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

// PostJSON issues a POST request using the stored credentials and decodes the JSON response
func (c *AuthenticatedClient) PostJSON(ctx context.Context, path string, body, out any) error {
	endpoint := buildEndpointURL(c.BaseURL, path, nil)
	return httpJSON(ctx, http.MethodPost, endpoint, c.BearerToken, c.Headers, body, out)
}

// httpJSON executes an HTTP request with optional bearer auth, additional headers, and JSON body, decoding the response into out
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

// buildHTTPRequestError constructs an HTTPRequestError from a non-2xx response, reading up to maxHTTPErrorBody bytes from the body
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

// buildEndpointURL resolves the final request URL by joining baseURL and path, then appending encoded query params
func buildEndpointURL(baseURL, path string, params url.Values) string {
	trimmedPath := strings.TrimSpace(path)

	var endpoint string
	switch {
	case strings.HasPrefix(trimmedPath, "http://"), strings.HasPrefix(trimmedPath, "https://"),
		strings.TrimSpace(baseURL) == "":
		endpoint = trimmedPath
	default:
		endpoint = strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(trimmedPath, "/")
	}

	return appendQueryParams(endpoint, params)
}

// appendQueryParams appends encoded query parameters to a URL, using & if the URL already contains a query string
func appendQueryParams(base string, params url.Values) string {
	if len(params) == 0 {
		return base
	}

	encoded := params.Encode()
	if encoded == "" {
		return base
	}

	if strings.Contains(base, "?") {
		return base + "&" + encoded
	}

	return base + "?" + encoded
}
