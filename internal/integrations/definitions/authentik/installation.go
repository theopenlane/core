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

	// build a temporary client to call the domains endpoint
	client := &Client{
		BaseURL: cred.BaseURL,
		Token:   cred.Token,
		HTTPClient: &http.Client{
			Timeout: authentikRequestTimeout,
		},
	}

	domain, tenantID, err := resolvePrimaryDomain(ctx, client)
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	return InstallationMetadata{
		Domain:   domain,
		TenantID: tenantID,
		BaseURL:  cred.BaseURL,
	}, true, nil
}

// resolvePrimaryDomain fetches the primary domain and tenant UUID from the Authentik instance
func resolvePrimaryDomain(ctx context.Context, client *Client) (string, string, error) {
	url := fmt.Sprintf("%s%s", client.BaseURL, authentikDomainsEndpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", ErrRequestBuildFailed
	}

	resp, err := client.do(ctx, req)
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()

	var result PaginatedResponse[DomainResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", ErrCredentialDecode
	}

	for _, d := range result.Results {
		if d.IsPrimary {
			return d.Domain, d.Tenant, nil
		}
	}

	// fallback to BaseURL if no primary domain found
	return client.BaseURL, "", nil
}
