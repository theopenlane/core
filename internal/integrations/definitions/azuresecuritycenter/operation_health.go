package azuresecuritycenter

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// HealthCheck reports whether the Security Center assessments API is accessible
type HealthCheck struct {
	// Count is the number of assessments returned by the health probe
	Count int `json:"count"`
}

// checkHealth verifies access by fetching the first page of security assessments
func checkHealth(ctx context.Context, _ types.OperationRequest, client *SecurityClient) (json.RawMessage, error) {
	pager := client.assessments.NewListPager(client.scope(), nil)

	page, err := pager.NextPage(ctx)
	if err != nil {
		return nil, ErrAssessmentFetchFailed
	}

	return providerkit.EncodeResult(HealthCheck{Count: len(page.Value)}, ErrResultEncode)
}
