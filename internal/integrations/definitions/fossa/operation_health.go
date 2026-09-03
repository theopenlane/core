package fossa

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// HealthCheck holds the result of a FOSSA health check
type HealthCheck struct {
	// OrganizationID is the FOSSA organization the token authenticates against
	OrganizationID string `json:"organizationId,omitempty"`
	// Subscription is the FOSSA subscription tier for the organization
	Subscription string `json:"subscription,omitempty"`
	// IssueCounts is the number of open issues per FOSSA issue category
	IssueCounts map[string]int `json:"issueCounts,omitempty"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClientRequest(fossaClient, func(ctx context.Context, _ types.OperationRequest, client *APIClient) (json.RawMessage, error) {
		return h.Run(ctx, client)
	})
}

// Run validates FOSSA access by reading the organization details and issue category counts.
// Both calls read data, so a push-only token fails here rather than at collection time.
func (HealthCheck) Run(ctx context.Context, c *APIClient) (json.RawMessage, error) {
	org, err := c.organization(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("fossa: error fetching organization details")

		return nil, err
	}

	counts, err := c.issueCategories(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("fossa: error fetching issue categories")

		return nil, ErrCategoriesFetchFailed
	}

	details := HealthCheck{
		OrganizationID: org.identifier(),
		Subscription:   org.Subscription,
		IssueCounts:    counts,
	}

	return providerkit.EncodeResult(details, ErrResultEncode)
}
