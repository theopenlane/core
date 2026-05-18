package authentik

import (
	"context"
	"net/http"
	"net/url"
	"time"

	authentikSDK "goauthentik.io/api/v3"

	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// authentikRequestTimeout is the per-request timeout for Authentik API calls
	authentikRequestTimeout = 30 * time.Second
)

// Client builds Authentik API clients for one installation
type Client struct{}

// Build constructs the Authentik API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return nil, err
	}

	if cred.Token == "" {
		return nil, ErrAPITokenMissing
	}

	if cred.BaseURL == "" {
		return nil, ErrBaseURLMissing
	}

	host, err := extractHost(cred.BaseURL)
	if err != nil {
		return nil, err
	}

	scheme, err := extractScheme(cred.BaseURL)
	if err != nil {
		return nil, err
	}

	cfg := authentikSDK.NewConfiguration()
	cfg.Host = host
	cfg.Scheme = scheme
	cfg.HTTPClient = &http.Client{Timeout: authentikRequestTimeout}
	cfg.AddDefaultHeader("Authorization", "Bearer "+cred.Token)

	return authentikSDK.NewAPIClient(cfg), nil
}

// resolveCredential extracts the CredentialSchema from the provided credential bindings
func resolveCredential(bindings types.CredentialBindings) (CredentialSchema, error) {
	cred, ok, err := authentikCredential.Resolve(bindings)
	if err != nil {
		return CredentialSchema{}, ErrCredentialDecode
	}

	if !ok {
		return CredentialSchema{}, ErrCredentialDecode
	}

	return cred, nil
}

// extractHost extracts the host from a base URL
func extractHost(baseURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	return u.Host, nil
}

// extractScheme extracts the scheme from a base URL
func extractScheme(baseURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	return u.Scheme, nil
}
