package helpers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"

	"github.com/theopenlane/core/internal/integrations/types"
)

const defaultHTTPTimeout = 10 * time.Second

var defaultHTTPRequester = httpsling.MustNew(
	httpsling.Client(httpclient.Timeout(defaultHTTPTimeout)),
)

// OAuthTokenFromPayload extracts a usable access token from the credential payload.
func OAuthTokenFromPayload(payload types.CredentialPayload, provider string) (string, error) {
	tokenOpt := payload.OAuthTokenOption()
	if !tokenOpt.IsPresent() {
		return "", fmt.Errorf("%w (provider %s)", ErrOAuthTokenMissing, provider)
	}

	token := tokenOpt.MustGet()
	if token == nil || token.AccessToken == "" {
		return "", fmt.Errorf("%w (provider %s)", ErrAccessTokenEmpty, provider)
	}

	return token.AccessToken, nil
}

// APITokenFromPayload extracts a raw API token from the credential payload.
func APITokenFromPayload(payload types.CredentialPayload, provider string) (string, error) {
	token := strings.TrimSpace(payload.Data.APIToken)
	if token == "" {
		return "", fmt.Errorf("%w (provider %s)", ErrAPITokenMissing, provider)
	}

	return token, nil
}

// HTTPGetJSON issues a GET request with the provided bearer token and decodes JSON responses.
func HTTPGetJSON(ctx context.Context, client *http.Client, url string, bearer string, headers map[string]string, out any) error {
	requester := defaultHTTPRequester
	if client != nil {
		var err error
		requester, err = httpsling.New(httpsling.WithHTTPClient(client))
		if err != nil {
			return err
		}
	}

	options := []httpsling.Option{
		httpsling.Get(url),
		httpsling.Header(httpsling.HeaderAccept, "application/json"),
	}
	if bearer != "" {
		options = append(options, httpsling.BearerAuth(bearer))
	}
	if len(headers) > 0 {
		options = append(options, httpsling.HeadersFromMap(headers))
	}

	resp, err := requester.ReceiveWithContext(ctx, out, options...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("%w (url %s): %s", ErrHTTPRequestFailed, url, resp.Status)
	}

	return nil
}
