package entitlements

import (
	"context"

	"github.com/stripe/stripe-go/v83"
)

// SubscriptionCreateOption allows customizing SubscriptionCreateParams
type SubscriptionCreateOption func(params *stripe.SubscriptionCreateParams)

// WithSubscriptionItems allows adding multiple items to the subscription creation params
func WithSubscriptionItems(items ...*stripe.SubscriptionCreateItemParams) SubscriptionCreateOption {
	return func(params *stripe.SubscriptionCreateParams) {
		params.Items = items
	}
}

// CreateSubscriptionWithOptions creates a subscription with functional options
func (sc *StripeClient) CreateSubscriptionWithOptions(ctx context.Context, baseParams *stripe.SubscriptionCreateParams, opts ...SubscriptionCreateOption) (*stripe.Subscription, error) {
	params := baseParams

	for _, opt := range opts {
		opt(params)
	}

	return sc.CreateSubscription(ctx, params)
}

// SubscriptionUpdateOption allows customizing SubscriptionUpdateParams
type SubscriptionUpdateOption func(params *stripe.SubscriptionUpdateParams)

// WithUpdateSubscriptionItems allows adding multiple items to the subscription update params
func WithUpdateSubscriptionItems(newItems ...*stripe.SubscriptionUpdateItemParams) SubscriptionUpdateOption {
	return func(params *stripe.SubscriptionUpdateParams) {
		params.Items = append(params.Items, newItems...)
	}
}

// UpdateSubscriptionWithOptions updates a subscription with functional options
func (sc *StripeClient) UpdateSubscriptionWithOptions(ctx context.Context, id string, baseParams *stripe.SubscriptionUpdateParams, opts ...SubscriptionUpdateOption) (*stripe.Subscription, error) {
	params := baseParams

	for _, opt := range opts {
		opt(params)
	}

	return sc.UpdateSubscription(ctx, id, params)
}

// AddNewItemsIfNotExist is a helper to add new items to update params if they don't already exist
func AddNewItemsIfNotExist(existing []*stripe.SubscriptionItem, params *stripe.SubscriptionUpdateParams, newItems ...*stripe.SubscriptionUpdateItemParams) {
	existingPrices := make(map[string]struct{})

	for _, item := range existing {
		if item.Price != nil {
			existingPrices[item.Price.ID] = struct{}{}
		}
	}

	for _, newItem := range newItems {
		if newItem.Price != nil {
			if _, found := existingPrices[*newItem.Price]; !found {
				params.Items = append(params.Items, newItem)
			}
		}
	}
}

// --- Example Usage ---
// Creating a subscription with multiple items:
// params := &stripe.SubscriptionCreateParams{
//     Customer: stripe.String("cus_123"),
// }
// items := []*stripe.SubscriptionCreateItemParams{
//     {Price: stripe.String("price_1")},
//     {Price: stripe.String("price_2")},
// }
// sub, err := sc.CreateSubscriptionWithOptions(ctx, params, WithSubscriptionItems(items...))

// Updating a subscription to add new items if they don't already exist:
// updateParams := &stripe.SubscriptionUpdateParams{}
// newItems := []*stripe.SubscriptionUpdateItemParams{
//     {Price: stripe.String("price_3")},
// }
// existingItems := sub.Items.Data // from the current subscription
// AddNewItemsIfNotExist(existingItems, updateParams, newItems...)
// updatedSub, err := sc.UpdateSubscriptionWithOptions(ctx, sub.ID, updateParams, WithUpdateSubscriptionItems(updateParams.Items...))
