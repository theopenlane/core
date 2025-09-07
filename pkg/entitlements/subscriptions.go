package entitlements

import (
	"context"
	"maps"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v82"
)

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

// CreateSubscription creates a new subscription
func (sc *StripeClient) CreateSubscription(ctx context.Context, params *stripe.SubscriptionCreateParams) (*stripe.Subscription, error) {
	start := time.Now()
	subscription, err := sc.Client.V1Subscriptions.Create(ctx, params)

	duration := time.Since(start).Seconds()

	status := StatusSuccess
	if err != nil {
		status = StatusError
	}

	stripeRequestCounter.WithLabelValues("subscriptions", status).Inc()
	stripeRequestDuration.WithLabelValues("subscriptions", status).Observe(duration)

	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// GetSubscriptionByID gets a subscription by ID
func (sc *StripeClient) GetSubscriptionByID(ctx context.Context, id string) (*stripe.Subscription, error) {
	start := time.Now()

	subscription, err := sc.Client.V1Subscriptions.Retrieve(ctx, id, &stripe.SubscriptionRetrieveParams{
		Params: stripe.Params{
			Expand: []*string{stripe.String("customer")},
		},
	})

	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}

	stripeRequestCounter.WithLabelValues("subscriptions", status).Inc()
	stripeRequestDuration.WithLabelValues("subscriptions", status).Observe(duration)

	if err != nil {
		return nil, err
	}

	return subscription, nil
}

var trialdays int64 = 30

// CreateSubscriptionWithPrices creates a subscription using the provided prices
func (sc *StripeClient) CreateSubscriptionWithPrices(ctx context.Context, cust *stripe.Customer, o *OrganizationCustomer) (*Subscription, error) {
	subsMetadata := make(map[string]string)
	if cust.Metadata != nil {
		maps.Copy(subsMetadata, cust.Metadata)
	} else {
		subsMetadata["organization_id"] = cust.ID
	}

	// we want 1 subscription, many prices, so we create items for each price ID
	items := []*stripe.SubscriptionCreateItemParams{}
	for _, price := range o.Prices {
		if price.ID != "" {
			items = append(items, &stripe.SubscriptionCreateItemParams{Price: stripe.String(price.ID)})
		}
	}

	params := &stripe.SubscriptionCreateParams{
		Customer:        stripe.String(cust.ID),
		Items:           items,
		PaymentBehavior: stripe.String(("default_incomplete")),
		PaymentSettings: &stripe.SubscriptionCreatePaymentSettingsParams{
			SaveDefaultPaymentMethod: stripe.String(stripe.SubscriptionPaymentSettingsSaveDefaultPaymentMethodOnSubscription),
		},
		Metadata:         subsMetadata,
		CollectionMethod: stripe.String(string(stripe.SubscriptionCollectionMethodChargeAutomatically)),
	}

	params.TrialPeriodDays = stripe.Int64(trialdays)
	params.TrialSettings = &stripe.SubscriptionCreateTrialSettingsParams{
		EndBehavior: &stripe.SubscriptionCreateTrialSettingsEndBehaviorParams{
			MissingPaymentMethod: stripe.String(stripe.SubscriptionTrialSettingsEndBehaviorMissingPaymentMethodPause),
		},
	}

	subs, err := sc.CreateSubscriptionWithOptions(ctx, params)
	if err != nil {
		log.Err(err).Msg("Failed to create subscription")

		return nil, err
	}

	log.Debug().Msgf("Created subscription with ID: %s", subs.ID)

	return sc.MapStripeSubscription(ctx, subs), nil
}

// retrieveActiveEntitlements retrieves active entitlements for a customer
func (sc *StripeClient) retrieveActiveEntitlements(ctx context.Context, customerID string) (feat []string, featNames []string, err error) {
	params := &stripe.EntitlementsActiveEntitlementListParams{
		Customer: stripe.String(customerID),
		Expand:   []*string{stripe.String("data.feature")},
	}

	result := sc.Client.V1EntitlementsActiveEntitlements.List(ctx, params)

	for feature, err := range result {
		if err != nil {
			log.Err(err).Msg("failed to list active entitlements")

			return nil, nil, err
		}

		feat = append(feat, feature.LookupKey)
		featNames = append(featNames, feature.Feature.Name)
	}

	return feat, featNames, nil
}

// MapStripeSubscription maps a stripe.Subscription to a "internal" subscription struct
func (sc *StripeClient) MapStripeSubscription(ctx context.Context, subs *stripe.Subscription) *Subscription {
	prices := []Price{}
	productID := ""

	if subs == nil || subs.Items == nil {
		log.Warn().Msg("subscription or subscription items is nil, unable to map data")
		return nil
	}

	if len(subs.Items.Data) > 1 {
		log.Warn().Msg("customer has more than one subscription")
	}

	for _, item := range subs.Items.Data {
		if item.Price == nil || item.Price.Product == nil {
			log.Warn().Msg("failed to map subscription item")

			continue
		}

		productID = item.Price.Product.ID

		product, err := sc.GetProductByID(ctx, productID)
		if err != nil {
			log.Warn().Err(err).Msg("failed to get product by ID")
		}

		interval := "month"
		if item.Price.Recurring != nil {
			interval = string(item.Price.Recurring.Interval)
		}

		prices = append(prices, Price{
			ID:          item.Price.ID,
			Price:       float64(item.Price.UnitAmount) / 100, // nolint:mnd
			ProductID:   productID,
			ProductName: product.Name,
			Interval:    interval,
			Currency:    string(item.Price.Currency),
		})
	}

	return &Subscription{
		ID:               subs.ID,
		Prices:           prices,
		TrialEnd:         subs.TrialEnd,
		ProductID:        productID,
		Status:           string(subs.Status),
		StripeCustomerID: subs.Customer.ID,
		OrganizationID:   subs.Metadata["organization_id"],
		DaysUntilDue:     subs.DaysUntilDue,
	}
}

// IsSubscriptionActive checks if a subscription is active based on its status
func IsSubscriptionActive(status stripe.SubscriptionStatus) bool {
	switch status {
	case stripe.SubscriptionStatusActive,
		stripe.SubscriptionStatusTrialing:
		return true
	case stripe.SubscriptionStatusPastDue,
		stripe.SubscriptionStatusIncomplete:
		return true
	case stripe.SubscriptionStatusCanceled,
		stripe.SubscriptionStatusIncompleteExpired,
		stripe.SubscriptionStatusUnpaid,
		stripe.SubscriptionStatusPaused:
		return false
	default:
		return false
	}
}

// ListSubscriptions retrieves all subscriptions for the given customer without creating new ones
func (sc *StripeClient) ListSubscriptions(ctx context.Context, customerID string) ([]*stripe.Subscription, error) {
	params := &stripe.SubscriptionListParams{Customer: stripe.String(customerID)}

	var subs []*stripe.Subscription

	it := sc.Client.V1Subscriptions.List(ctx, params)
	for s, err := range it {
		if err != nil {
			return nil, err
		}

		subs = append(subs, s)
	}

	return subs, nil
}

// MigrateSubscriptionPrice replaces occurrences of oldPriceID with newPriceID on the subscription
// Don't run unless you know what you're doing!
func (sc *StripeClient) MigrateSubscriptionPrice(ctx context.Context, sub *stripe.Subscription, oldPriceID, newPriceID string) (*stripe.Subscription, error) {
	if sub == nil {
		return nil, nil
	}

	var updateItems []*stripe.SubscriptionUpdateItemParams

	for _, item := range sub.Items.Data {
		if item.Price != nil && item.Price.ID == oldPriceID {
			updateItems = append(updateItems, &stripe.SubscriptionUpdateItemParams{
				ID:    stripe.String(item.ID),
				Price: stripe.String(newPriceID),
			})
		}
	}

	if len(updateItems) == 0 {
		return sub, nil
	}

	params := &stripe.SubscriptionUpdateParams{Items: updateItems}

	return sc.UpdateSubscription(ctx, sub.ID, params)
}

// UpdateSubscription updates a subscription
func (sc *StripeClient) UpdateSubscription(ctx context.Context, id string, params *stripe.SubscriptionUpdateParams) (*stripe.Subscription, error) {
	subscription, err := sc.Client.V1Subscriptions.Update(ctx, id, params)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// CancelSubscription cancels a subscription
func (sc *StripeClient) CancelSubscription(ctx context.Context, id string, params *stripe.SubscriptionCancelParams) (*stripe.Subscription, error) {
	subscription, err := sc.Client.V1Subscriptions.Cancel(ctx, id, params)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}
