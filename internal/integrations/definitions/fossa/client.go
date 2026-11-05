package fossa

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

const (
	// defaultBaseURL is the FOSSA SaaS base URL used when the credential does not override it
	defaultBaseURL = "https://app.fossa.com"
	// fossaRequestTimeout is the per-request timeout for FOSSA API calls
	fossaRequestTimeout = 30 * time.Second
)

// APIClient is a thin FOSSA REST API client with the base URL and bearer token pre-applied
type APIClient struct {
	// requester is the underlying httpsling requester carrying the base URL and auth header
	requester *httpsling.Requester
}

// clientBuilder returns the FOSSA client constructor, allowPrivate permits non-public base urls for local development
func clientBuilder(allowPrivate bool) func(context.Context, types.ConnectionRequest[CredentialSchema]) (*APIClient, error) {
	return func(_ context.Context, req types.ConnectionRequest[CredentialSchema]) (*APIClient, error) {
		return buildClient(req, allowPrivate)
	}
}

// buildClient constructs the FOSSA API client for one installation
func buildClient(req types.ConnectionRequest[CredentialSchema], allowPrivate bool) (*APIClient, error) {
	if req.Credential.APIToken == "" {
		return nil, ErrAPITokenMissing
	}

	return NewAPIClient(baseURLOrDefault(req.Credential.BaseURL), req.Credential.APIToken, allowPrivate)
}

// NewAPIClient constructs a FOSSA API client for the supplied base URL and token, the customer supplied
// base URL must be public unless allowPrivate is set
func NewAPIClient(baseURL, token string, allowPrivate bool) (*APIClient, error) {
	httpClient, err := urlx.NewCustomerHTTPClient(baseURL, allowPrivate, httpclient.Timeout(fossaRequestTimeout))
	if err != nil {
		return nil, err
	}

	requester, err := httpsling.New(
		httpsling.WithHTTPClient(httpClient),
		httpsling.URL(baseURL),
	)
	if err != nil {
		return nil, ErrClientBuild
	}

	if err := requester.Apply(httpsling.BearerAuth(token)); err != nil {
		return nil, ErrClientBuild
	}

	return &APIClient{requester: requester}, nil
}

// get issues a GET request and decodes a successful response into out, checking the status first
// because httpsling decodes non-2xx bodies too and an auth failure must not surface as a decode error
func (c *APIClient) get(ctx context.Context, path string, params map[string]string, out any) error {
	opts := make([]httpsling.Option, 0, len(params)+1)
	opts = append(opts, httpsling.Get(path))

	for key, value := range params {
		opts = append(opts, httpsling.QueryParam(key, value))
	}

	resp, err := c.requester.ReceiveWithContext(ctx, out, opts...)
	if resp != nil {
		defer resp.Body.Close() //nolint:errcheck
	}

	if resp == nil {
		logx.FromContext(ctx).Error().Err(err).Str("path", path).Msg("fossa: request failed with no response")

		return ErrAPIRequest
	}

	if !httpsling.IsSuccess(resp) {
		return statusError(resp.StatusCode)
	}

	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("path", path).Int("status", resp.StatusCode).Msg("fossa: unable to decode response")

		return ErrAPIRequest
	}

	return nil
}

// statusError maps a non-success FOSSA response status to a sentinel error
func statusError(status int) error {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrUnauthorized
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		return ErrAPIRequest
	}
}

// baseURLOrDefault normalizes the configured base URL, falling back to the FOSSA SaaS endpoint
func baseURLOrDefault(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmed == "" {
		return defaultBaseURL
	}

	return trimmed
}
