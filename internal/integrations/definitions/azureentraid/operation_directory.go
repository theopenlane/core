package azureentraid

import (
	"context"
	"encoding/json"

	"github.com/samber/lo"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// DirectoryInspect collects Azure Entra ID tenant metadata
type DirectoryInspect struct {
	// ID is the organization identifier
	ID string `json:"id"`
	// DisplayName is the organization display name
	DisplayName string `json:"displayName"`
	// VerifiedDomains is the list of verified domains for the tenant
	VerifiedDomains []verifiedDomain `json:"verifiedDomains"`
}

// verifiedDomain holds one verified domain entry for an Entra ID tenant
type verifiedDomain struct {
	// Name is the domain name
	Name string `json:"name"`
	// IsDefault indicates whether this is the default domain for the tenant
	IsDefault bool `json:"isDefault"`
}

// Handle adapts directory inspect to the generic operation registration boundary
func (d DirectoryInspect) Handle() types.OperationHandler {
	return providerkit.OperationWithClient(EntraClient, d.Run)
}

// Run collects Azure Entra ID tenant metadata via Microsoft Graph
func (DirectoryInspect) Run(ctx context.Context, c *msgraphsdk.GraphServiceClient) (json.RawMessage, error) {
	result, err := c.Organization().Get(ctx, nil)
	if err != nil {
		return nil, ErrOrganizationLookupFailed
	}

	orgs := result.GetValue()
	if len(orgs) == 0 {
		return nil, ErrNoOrganizations
	}

	org := orgs[0]

	return providerkit.EncodeResult(DirectoryInspect{
		ID:              lo.FromPtr(org.GetId()),
		DisplayName:     lo.FromPtr(org.GetDisplayName()),
		VerifiedDomains: collectVerifiedDomains(org.GetVerifiedDomains()),
	}, ErrResultEncode)
}

// collectVerifiedDomains maps SDK verified domain models to the local type
func collectVerifiedDomains(raw []models.VerifiedDomainable) []verifiedDomain {
	out := make([]verifiedDomain, 0, len(raw))
	for _, d := range raw {
		if d == nil {
			continue
		}
		out = append(out, verifiedDomain{
			Name:      lo.FromPtr(d.GetName()),
			IsDefault: lo.FromPtr(d.GetIsDefault()),
		})
	}
	return out
}
