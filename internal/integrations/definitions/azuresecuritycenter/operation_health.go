package azuresecuritycenter

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck reports whether the Security Center assessments API is accessible
type HealthCheck struct {
	// Count is the number of assessments returned by the health probe
	Count int `json:"count"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClient(securityCenterClient, h.Run)
}

// Run verifies access by fetching the first page of security assessments
func (HealthCheck) Run(ctx context.Context, client *azureSecurityClient) (json.RawMessage, error) {
	pager := client.assessments.NewListPager(client.scope(), nil)

	page, err := pager.NextPage(ctx)
	if err != nil {
		return nil, ErrAssessmentFetchFailed
	}

	return providerkit.EncodeResult(HealthCheck{Count: len(page.Value)}, ErrResultEncode)
}
