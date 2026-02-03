package helpers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"

	"github.com/theopenlane/core/common/integrations/types"
)

const (
	defaultHTTPTimeout = 10 * time.Second
	maxHTTPErrorBody   = 2048
)

var defaultHTTPRequester = httpsling.MustNew(
	httpsling.Client(httpclient.Timeout(defaultHTTPTimeout)),
)

// OAuthTokenFromPayload extracts a usable access token from the credential payload.
func OAuthTokenFromPayload(payload types.CredentialPayload, provider string) (string, error) {
	tokenOpt := payload.OAuthTokenOption()
	if !tokenOpt.IsPresent() {
		return "", ErrOAuthTokenMissing
	}

	token := tokenOpt.MustGet()
	if token == nil || token.AccessToken == "" {
		return "", ErrAccessTokenEmpty
	}

	return token.AccessToken, nil
}

// APITokenFromPayload extracts a raw API token from the credential payload.
func APITokenFromPayload(payload types.CredentialPayload, provider string) (string, error) {
	token := strings.TrimSpace(payload.Data.APIToken)
	if token == "" {
		return "", ErrAPITokenMissing
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
		return httpRequestError(resp, url)
	}

	return nil
}

// HTTPPostJSON issues a POST request with the provided bearer token and JSON body, then decodes JSON responses.
func HTTPPostJSON(ctx context.Context, client *http.Client, url string, bearer string, headers map[string]string, body any, out any) error {
	requester := defaultHTTPRequester
	if client != nil {
		var err error
		requester, err = httpsling.New(httpsling.WithHTTPClient(client))
		if err != nil {
			return err
		}
	}

	options := []httpsling.Option{
		httpsling.Post(url),
		httpsling.Header(httpsling.HeaderAccept, "application/json"),
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

// RandomState generates a URL-safe random string using crypto/rand
func RandomState(bytes int) (string, error) {
	if bytes <= 0 {
		bytes = 32
	}

	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("%w: %w", ErrRandomStateGeneration, err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func httpRequestError(resp *http.Response, url string) error {
	if resp == nil {
		return ErrHTTPRequestFailed
	}
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
