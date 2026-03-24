package azureentraid

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// graphOrganizationURL is the Microsoft Graph endpoint for tenant organization metadata
const graphOrganizationURL = "https://graph.microsoft.com/v1.0/organization"

// resolveInstallationMetadata derives Azure Entra installation metadata from the persisted credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	var cred entraIDCred
	if err := jsonx.UnmarshalIfPresent(req.Credential.Data, &cred); err != nil {
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.TenantID == "" {
		return InstallationMetadata{}, false, nil
	}

	meta := InstallationMetadata{
		TenantID: cred.TenantID,
	}

	if cred.AccessToken != "" {
		enrichOrganizationMetadata(ctx, cred.AccessToken, &meta)
	}

	return meta, true, nil
}

// graphOrganizationResponse represents the Microsoft Graph /organization response
type graphOrganizationResponse struct {
	// Value holds the organization entries
	Value []graphOrganization `json:"value"`
}

// graphOrganization represents a single organization entry from Microsoft Graph
type graphOrganization struct {
	// DisplayName is the organization display name
	DisplayName string `json:"displayName"`
	// VerifiedDomains holds the verified domain entries
	VerifiedDomains []graphVerifiedDomain `json:"verifiedDomains"`
}

// graphVerifiedDomain represents a verified domain from the Microsoft Graph organization response
type graphVerifiedDomain struct {
	// Name is the domain name
	Name string `json:"name"`
	// IsDefault indicates whether this is the default domain
	IsDefault bool `json:"isDefault"`
}

// enrichOrganizationMetadata fetches tenant display name and verified domains from Microsoft Graph
func enrichOrganizationMetadata(ctx context.Context, accessToken string, meta *InstallationMetadata) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, graphOrganizationURL, nil)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var orgResp graphOrganizationResponse
	if err := json.Unmarshal(body, &orgResp); err != nil {
		return
	}

	if len(orgResp.Value) == 0 {
		return
	}

	org := orgResp.Value[0]
	meta.DisplayName = org.DisplayName
	meta.VerifiedDomains = lo.Map(org.VerifiedDomains, func(d graphVerifiedDomain, _ int) VerifiedDomain {
		return VerifiedDomain(d)
	})
}
