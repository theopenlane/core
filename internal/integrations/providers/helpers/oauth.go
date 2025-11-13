package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/theopenlane/core/internal/integrations/types"
)

var defaultHTTPClient = &http.Client{Timeout: 10 * time.Second}

// OAuthTokenFromPayload extracts a usable access token from the credential payload.
func OAuthTokenFromPayload(payload types.CredentialPayload, provider string) (string, error) {
	tokenOpt := payload.OAuthTokenOption()
	if !tokenOpt.IsPresent() {
		return "", fmt.Errorf("%s: oauth token missing", provider)
	}

	token := tokenOpt.MustGet()
	if token == nil || token.AccessToken == "" {
		return "", fmt.Errorf("%s: access token empty", provider)
	}

	return token.AccessToken, nil
}

// APITokenFromPayload extracts a raw API token from the credential payload.
func APITokenFromPayload(payload types.CredentialPayload, provider string) (string, error) {
	token := strings.TrimSpace(payload.Data.APIToken)
	if token == "" {
		return "", fmt.Errorf("%s: api token missing", provider)
	}

	return token, nil
}

// HTTPGetJSON issues a GET request with the provided bearer token and decodes JSON responses.
func HTTPGetJSON(ctx context.Context, client *http.Client, url string, bearer string, headers map[string]string, out any) error {
	if client == nil {
		client = defaultHTTPClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	req.Header.Set("Accept", "application/json")
	for key, value := range headers {
		if value == "" {
			continue
		}
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request to %s failed: %s", url, resp.Status)
	}

	if out == nil {
		return nil
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(out); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
