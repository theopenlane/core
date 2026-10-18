package microsoft

import (
	"context"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"

	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// subscriptionsListPath is the ARM path for listing subscriptions visible to the caller
	subscriptionsListPath = "/subscriptions?api-version=2022-12-01"
)

// subscriptionStatesExcluded are subscription states that cannot be queried for resources
var subscriptionStatesExcluded = map[string]bool{
	"Disabled": true,
	"Deleted":  true,
}

// subscription is one entry from the ARM subscriptions listing
type subscription struct {
	// SubscriptionID is the Azure subscription identifier
	SubscriptionID string `json:"subscriptionId"`
	// TenantID is the tenant that owns the subscription, which differs from the caller's tenant for Lighthouse delegations
	TenantID string `json:"tenantId"`
	// State is the subscription lifecycle state (Enabled, Warned, PastDue, Disabled, Deleted)
	State string `json:"state"`
}

// subscriptionListPage is one page of the ARM subscriptions listing
type subscriptionListPage struct {
	// Value is the subscriptions on this page
	Value []subscription `json:"value"`
	// NextLink is the absolute URL of the next page, empty on the last page
	NextLink string `json:"nextLink"`
}

// ListSubscriptionIDs returns every usable subscription owned by tenantID that the credential has an RBAC role assignment on
func (c *ARMClient) ListSubscriptionIDs(ctx context.Context, tenantID string) ([]string, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	var ids []string

	next := c.client.Endpoint() + subscriptionsListPath
	for next != "" {
		page, err := getARM[subscriptionListPage](ctx, c.client.Pipeline(), next, ErrSubscriptionListFailed)
		if err != nil {
			return nil, err
		}

		for _, sub := range page.Value {
			if !subscriptionStatesExcluded[sub.State] && strings.EqualFold(sub.TenantID, tenantID) {
				ids = append(ids, sub.SubscriptionID)
			}
		}

		next = page.NextLink
	}

	return ids, nil
}

// getARM requests and decodes one ARM GET response, returning failure on any request, status, or decode error
func getARM[T any](ctx context.Context, pipeline runtime.Pipeline, requestURL string, failure error) (T, error) {
	var out T

	req, err := runtime.NewRequest(ctx, http.MethodGet, requestURL)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("url", requestURL).Msg("failed to build arm request")
		return out, failure
	}

	resp, err := pipeline.Do(req)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("url", requestURL).Msg("arm request failed")
		return out, failure
	}

	defer resp.Body.Close()

	if !runtime.HasStatusCode(resp, http.StatusOK) {
		logx.FromContext(ctx).Error().Err(runtime.NewResponseError(resp)).Str("url", requestURL).Msg("arm request returned unexpected status")
		return out, failure
	}

	if err := runtime.UnmarshalAsJSON(resp, &out); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("url", requestURL).Msg("failed to decode arm response")
		return out, failure
	}

	return out, nil
}
