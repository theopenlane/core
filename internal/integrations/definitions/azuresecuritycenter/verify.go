package azuresecuritycenter

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// verify probes the assessments API and derives the Azure subscription metadata
func verify(ctx context.Context, req types.ConnectionRequest[CredentialSchema], c *SecurityClient) (InstallationMetadata, error) {
	pager := c.assessments.NewListPager(c.scope(), nil)

	if _, err := pager.NextPage(ctx); err != nil {
		return InstallationMetadata{}, ErrAssessmentFetchFailed
	}

	return InstallationMetadata{
		TenantID:       req.Credential.TenantID,
		ClientID:       req.Credential.ClientID,
		SubscriptionID: req.Credential.SubscriptionID,
	}, nil
}
