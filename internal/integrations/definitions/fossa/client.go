package fossa

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"

	"github.com/theopenlane/core/internal/integrations/types"
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

// newAPIClient constructs a FOSSA API client for the supplied base URL and token
func newAPIClient(baseURL, token string) (*APIClient, error) {
	requester, err := httpsling.New(
		httpsling.Client(httpclient.Timeout(fossaRequestTimeout)),
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

// get issues a GET request against the FOSSA API and decodes a successful response into out.
//
// httpsling does not treat a non-2xx status as an error; it unmarshals whatever body came back
// into out and returns any unmarshal error. The status is therefore checked before the decode
// result is trusted, so an auth failure surfaces as ErrUnauthorized rather than a decode error.
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
		return ErrAPIRequest
	}

	if !httpsling.IsSuccess(resp) {
		return statusError(resp.StatusCode)
	}

	if err != nil {
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

// ClientBuilder builds FOSSA API clients for one installation
type ClientBuilder struct{}

// Build constructs the FOSSA API client for one installation
func (ClientBuilder) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return nil, err
	}

	if cred.APIToken == "" {
		return nil, ErrAPITokenMissing
	}

	return newAPIClient(baseURLOrDefault(cred.BaseURL), cred.APIToken)
}

// resolveCredential extracts the CredentialSchema from the provided credential bindings
func resolveCredential(bindings types.CredentialBindings) (CredentialSchema, error) {
	cred, ok, err := fossaCredential.Resolve(bindings)
	if err != nil || !ok {
		return CredentialSchema{}, ErrCredentialDecode
	}

	return cred, nil
}

// baseURLOrDefault normalizes the configured base URL, falling back to the FOSSA SaaS endpoint
func baseURLOrDefault(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmed == "" {
		return defaultBaseURL
	}

	return trimmed
}
