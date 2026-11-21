package aikido

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/logx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const (
	// pathToken is the OAuth client credentials token endpoint on the app host
	pathToken = "/api/oauth/token"
	// pathAPI is the public REST API prefix on the app host
	pathAPI = "/api/public/v1"
	// requestTimeout is the per-request timeout for Aikido API calls
	requestTimeout = 60 * time.Second
	// maxRateLimitRetries bounds how often one request is retried after a 429
	maxRateLimitRetries = 3
	// maxRetryAfter caps the wait between rate limited retries, Aikido uses a one minute sliding window
	maxRetryAfter = 60 * time.Second
)

// APIClient is a thin Aikido REST API client with the base URL and OAuth token source pre-applied
type APIClient struct {
	// requester is the underlying httpsling requester carrying the base URL and authenticated transport
	requester *httpsling.Requester
}

// buildClient constructs the Aikido API client for one installation
func buildClient(ctx context.Context, req types.ConnectionRequest[CredentialSchema]) (*APIClient, error) {
	if req.Credential.ClientID == "" || req.Credential.ClientSecret == "" {
		return nil, ErrClientCredentialsMissing
	}

	appURL, err := appURLForRegion(req.Credential.Region)
	if err != nil {
		return nil, err
	}

	return NewAPIClient(ctx, appURL, req.Credential.ClientID, req.Credential.ClientSecret)
}

// oauthClientBuilder returns the client constructor for installations connected through the Openlane Aikido OAuth app
func oauthClientBuilder(cfg Config) func(context.Context, types.ConnectionRequest[aikidoOAuthCred]) (*APIClient, error) {
	return func(ctx context.Context, req types.ConnectionRequest[aikidoOAuthCred]) (*APIClient, error) {
		return buildOAuthClient(ctx, cfg, oauthAppURL, req.Credential)
	}
}

// buildOAuthClient constructs the Aikido API client from stored OAuth token material, refreshing it against the app host
func buildOAuthClient(ctx context.Context, cfg Config, appURL string, cred aikidoOAuthCred) (*APIClient, error) {
	if cred.AccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	source := providerkit.OAuthTokenSource(tokenContext(ctx), cfg.ClientID, cfg.ClientSecret, oauthEndpoint(appURL), providerkit.OAuthToken(cred.AccessToken, cred.RefreshToken, cred.Expiry))

	return newAPIClient(ctx, appURL, source)
}

// NewAPIClient constructs an Aikido API client for the supplied app host and client credentials
func NewAPIClient(ctx context.Context, appURL, clientID, clientSecret string) (*APIClient, error) {
	config := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     appURL + pathToken,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	return newAPIClient(ctx, appURL, config.TokenSource(tokenContext(ctx)))
}

// oauthEndpoint returns the Aikido OAuth endpoints on the app host
func oauthEndpoint(appURL string) oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:   appURL + oauthAuthorizePath,
		TokenURL:  appURL + pathToken,
		AuthStyle: oauth2.AuthStyleInHeader,
	}
}

// tokenContext carries the token HTTP client and detaches cancellation, since the token source outlives the build context
func tokenContext(ctx context.Context) context.Context {
	return context.WithValue(context.WithoutCancel(ctx), oauth2.HTTPClient, &http.Client{Timeout: requestTimeout})
}

// newAPIClient constructs an Aikido API client authenticating every request with the supplied token source
func newAPIClient(ctx context.Context, appURL string, source oauth2.TokenSource) (*APIClient, error) {
	httpClient := oauth2.NewClient(tokenContext(ctx), source)
	httpClient.Timeout = requestTimeout

	requester, err := httpsling.New(
		httpsling.WithHTTPClient(httpClient),
		httpsling.URL(appURL),
	)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("aikido: failed to create requester")

		return nil, ErrClientBuild
	}

	return &APIClient{requester: requester}, nil
}

// get issues a GET request and decodes a successful response into out, retrying while Aikido rate limits the call
func (c *APIClient) get(ctx context.Context, path string, params map[string]string, out any) (http.Header, error) {
	opts := make([]httpsling.Option, 0, len(params)+1)
	opts = append(opts, httpsling.Get(path))

	for key, value := range params {
		opts = append(opts, httpsling.QueryParam(key, value))
	}

	for attempt := 0; ; attempt++ {
		status, header, err := c.receive(ctx, path, out, opts)
		if status != http.StatusTooManyRequests || attempt >= maxRateLimitRetries {
			return header, err
		}

		retryAfter := retryAfterDelay(header.Get("Retry-After"))

		logx.FromContext(ctx).Warn().Str("path", path).Dur("retry_after", retryAfter).Int("attempt", attempt+1).Msg("aikido: rate limited, waiting before retry")

		if err := waitFor(ctx, retryAfter); err != nil {
			return nil, err
		}
	}
}

// receive performs one request and returns the response status, the response headers and the mapped error,
// the status is checked first because httpsling decodes non-2xx bodies too
func (c *APIClient) receive(ctx context.Context, path string, out any, opts []httpsling.Option) (int, http.Header, error) {
	resp, err := c.requester.ReceiveWithContext(ctx, out, opts...)
	if resp != nil {
		defer resp.Body.Close() //nolint:errcheck
	}

	if resp == nil {
		if _, ok := errors.AsType[*oauth2.RetrieveError](err); ok {
			logx.FromContext(ctx).Error().Err(err).Str("path", path).Msg("aikido: token request was rejected")

			return 0, nil, ErrUnauthorized
		}

		if ctxErr := ctx.Err(); ctxErr != nil {
			return 0, nil, ctxErr
		}

		logx.FromContext(ctx).Error().Err(err).Str("path", path).Msg("aikido: request failed with no response")

		return 0, nil, ErrAPIRequest
	}

	if !httpsling.IsSuccess(resp) {
		logx.FromContext(ctx).Error().Err(err).Str("path", path).Int("status", resp.StatusCode).Msg("aikido: unexpected response status")

		return resp.StatusCode, resp.Header, statusError(resp.StatusCode)
	}

	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("path", path).Int("status", resp.StatusCode).Msg("aikido: unable to decode response")

		return resp.StatusCode, resp.Header, ErrAPIRequest
	}

	return resp.StatusCode, resp.Header, nil
}

// waitFor blocks for d or until ctx is done
func waitFor(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// retryAfterDelay parses a Retry-After seconds header, falling back to the full rate limit window
func retryAfterDelay(header string) time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(header))
	if err != nil || seconds < 0 {
		return maxRetryAfter
	}

	return min(time.Duration(seconds)*time.Second, maxRetryAfter)
}

// statusError maps a non-success Aikido response status to a sentinel error
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

// appURLForRegion resolves the Aikido app host for a region, defaulting to the EU region
func appURLForRegion(region string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(region))
	if normalized == "" {
		normalized = regionEU
	}

	appURL, ok := regionAppURLs[normalized]
	if !ok {
		return "", ErrRegionInvalid
	}

	return appURL, nil
}
