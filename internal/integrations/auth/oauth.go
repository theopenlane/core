package auth

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"

	"github.com/theopenlane/core/common/models"
)

const (
	defaultHTTPTimeout = 10 * time.Second
	maxHTTPErrorBody   = 2048
)

var defaultHTTPRequester = httpsling.MustNew(
	httpsling.Client(httpclient.Timeout(defaultHTTPTimeout)),
)

// OAuthTokenFromPayload extracts a usable access token from the credential set.
func OAuthTokenFromPayload(payload models.CredentialSet) (string, error) {
	if payload.OAuthAccessToken == "" &&
		payload.OAuthRefreshToken == "" &&
		payload.OAuthTokenType == "" &&
		payload.OAuthExpiry == nil {
		return "", ErrOAuthTokenMissing
	}

	if payload.OAuthAccessToken == "" {
		return "", ErrAccessTokenEmpty
	}

	return payload.OAuthAccessToken, nil
}

// APITokenFromPayload extracts a raw API token from the credential set.
func APITokenFromPayload(payload models.CredentialSet) (string, error) {
	token := payload.APIToken
	if token == "" {
		return "", ErrAPITokenMissing
	}

	return token, nil
}

// HTTPGetJSON issues a GET request with the provided bearer token and decodes JSON responses
func HTTPGetJSON(ctx context.Context, client *http.Client, url string, bearer string, headers map[string]string, out any) error {
	return httpJSON(ctx, client, http.MethodGet, url, bearer, headers, nil, out)
}

// HTTPPostJSON issues a POST request with the provided bearer token and JSON body, then decodes JSON responses
func HTTPPostJSON(ctx context.Context, client *http.Client, url string, bearer string, headers map[string]string, body any, out any) error {
	return httpJSON(ctx, client, http.MethodPost, url, bearer, headers, body, out)
}

func httpJSON(ctx context.Context, client *http.Client, method string, url string, bearer string, headers map[string]string, body any, out any) error {
	requester := defaultHTTPRequester
	if client != nil {
		var err error
		requester, err = httpsling.New(httpsling.WithHTTPClient(client))
		if err != nil {
			return err
		}
	}

	options := []httpsling.Option{
		httpsling.Method(method),
		httpsling.URL(url),
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

	resp, err := requester.ReceiveWithContext(ctx, out, options...)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return httpRequestError(resp, url)
	}

	return nil
}

// httpRequestError constructs an HTTPRequestError from a non-2xx HTTP response
func httpRequestError(resp *http.Response, url string) error {
	body := ""
	if resp.Body != nil {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, maxHTTPErrorBody))
		body = strings.TrimSpace(string(data))
	}

	return &HTTPRequestError{
		URL:        url,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Body:       body,
	}
}
