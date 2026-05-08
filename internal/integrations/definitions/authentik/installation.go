package authentik

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives Authentik instance metadata from the persisted credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.Token == "" {
		return InstallationMetadata{}, false, nil
	}

	if cred.BaseURL == "" {
		return InstallationMetadata{}, false, nil
	}

	// build a temporary client to call the system endpoint
	client := &Client{
		BaseURL: cred.BaseURL,
		Token:   cred.Token,
		HTTPClient: &http.Client{
			Timeout: authentikRequestTimeout,
		},
	}

	brand, host, err := resolveSystemInfo(ctx, client)
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	return InstallationMetadata{
		Brand:   brand,
		Host:    host,
		BaseURL: cred.BaseURL,
	}, true, nil
}

// resolveSystemInfo fetches the brand name and HTTP host from the Authentik admin system endpoint
func resolveSystemInfo(ctx context.Context, client *Client) (string, string, error) {
	url := fmt.Sprintf("%s%s", client.BaseURL, authentikSystemEndpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", ErrRequestBuildFailed
	}

	resp, err := client.do(ctx, req)
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()

	var result SystemResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", ErrCredentialDecode
	}

	return result.Brand, result.HTTPHost, nil
}
