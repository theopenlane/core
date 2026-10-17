package microsoft

import (
	"context"
	"net/http"

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

// ListSubscriptionIDs returns every usable subscription the credential has an RBAC role assignment on
func (c *ARMClient) ListSubscriptionIDs(ctx context.Context) ([]string, error) {
	var ids []string

	next := c.client.Endpoint() + subscriptionsListPath
	for next != "" {
		page, err := fetchSubscriptionPage(ctx, c.client.Pipeline(), next)
		if err != nil {
			return nil, err
		}

		for _, sub := range page.Value {
			if !subscriptionStatesExcluded[sub.State] {
				ids = append(ids, sub.SubscriptionID)
			}
		}

		next = page.NextLink
	}

	return ids, nil
}

// fetchSubscriptionPage requests and decodes one page of the ARM subscriptions listing
func fetchSubscriptionPage(ctx context.Context, pipeline runtime.Pipeline, pageURL string) (subscriptionListPage, error) {
	req, err := runtime.NewRequest(ctx, http.MethodGet, pageURL)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to build subscriptions request")
		return subscriptionListPage{}, ErrSubscriptionListFailed
	}

	resp, err := pipeline.Do(req)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("subscriptions request failed")
		return subscriptionListPage{}, ErrSubscriptionListFailed
	}

	defer resp.Body.Close()

	if !runtime.HasStatusCode(resp, http.StatusOK) {
		logx.FromContext(ctx).Error().Err(runtime.NewResponseError(resp)).Msg("subscriptions request returned unexpected status")
		return subscriptionListPage{}, ErrSubscriptionListFailed
	}

	var page subscriptionListPage
	if err := runtime.UnmarshalAsJSON(resp, &page); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to decode subscriptions response")
		return subscriptionListPage{}, ErrSubscriptionListFailed
	}

	return page, nil
}
