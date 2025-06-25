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

// ListStripeSubscriptions lists stripe subscriptions by customer
func (sc *StripeClient) ListOrCreateSubscriptions(ctx context.Context, customerID string) (*Subscription, error) {
	result := sc.Client.V1Subscriptions.List(ctx, &stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
	})

	if seq2IsEmpty(result) {
		sub, err := sc.CreateTrialSubscription(ctx, &stripe.Customer{ID: customerID})
		if err != nil {
			log.Error().Err(err).Msg("failed to create trial subscription")
			return nil, err
		}

		return sub, nil
	}

	var customerSub *stripe.Subscription

	subCount := 0

	for sub, err := range result {
		if err != nil {
			log.Error().Err(err).Msg("failed to list subscriptions")

			return nil, err
		}

		customerSub = sub
		subCount++
	}

	if subCount > 1 {
		log.Warn().Msg("customer has more than one subscription")

		return nil, ErrMultipleSubscriptions
	}

	// assumes customer can only have 1 subscription if there are any
	subs := sc.MapStripeSubscription(ctx, customerSub)

	return subs, nil
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

var trialdays int64 = 30

// CreateTrialSubscription creates a trial subscription with the configured price
func (sc *StripeClient) CreateTrialSubscription(ctx context.Context, cust *stripe.Customer) (*Subscription, error) {
	subsMetadata := make(map[string]string)
	if cust.Metadata != nil {
		maps.Copy(subsMetadata, cust.Metadata)
	} else {
		subsMetadata["organization_id"] = cust.ID
	}

	baseParams := &stripe.SubscriptionCreateParams{
		Customer:        stripe.String(cust.ID),
		TrialPeriodDays: stripe.Int64(trialdays),
		PaymentSettings: &stripe.SubscriptionCreatePaymentSettingsParams{
			SaveDefaultPaymentMethod: stripe.String(string(stripe.SubscriptionPaymentSettingsSaveDefaultPaymentMethodOnSubscription)),
		},
		Metadata:         subsMetadata,
		CollectionMethod: stripe.String(string(stripe.SubscriptionCollectionMethodChargeAutomatically)),
		TrialSettings: &stripe.SubscriptionCreateTrialSettingsParams{
			EndBehavior: &stripe.SubscriptionCreateTrialSettingsEndBehaviorParams{
				MissingPaymentMethod: stripe.String(string(stripe.SubscriptionTrialSettingsEndBehaviorMissingPaymentMethodPause)),
			},
		},
	}

	item := &stripe.SubscriptionCreateItemParams{Price: &sc.Config.TrialSubscriptionPriceID}

	subs, err := sc.CreateSubscriptionWithOptions(ctx, baseParams, WithSubscriptionItems(item))
	if err != nil {
		log.Err(err).Msg("Failed to create trial subscription")
		return nil, err
	}

	log.Debug().Msgf("Created trial subscription with ID: %s", subs.ID)

	mappedsubscription := sc.MapStripeSubscription(ctx, subs)

	return mappedsubscription, nil
}

// CreatePersonalOrgFreeTierSubs creates a subscription with the configured $0 price used for personal organizations only
func (sc *StripeClient) CreatePersonalOrgFreeTierSubs(ctx context.Context, customerID string) (*Subscription, error) {
	baseParams := &stripe.SubscriptionCreateParams{
		Customer:         stripe.String(customerID),
		CollectionMethod: stripe.String(string(stripe.SubscriptionCollectionMethodChargeAutomatically)),
	}

	item := &stripe.SubscriptionCreateItemParams{Price: &sc.Config.PersonalOrgSubscriptionPriceID}

	subs, err := sc.CreateSubscriptionWithOptions(ctx, baseParams, WithSubscriptionItems(item))
	if err != nil {
		log.Err(err).Msg("Failed to create trial subscription")
		return nil, err
	}

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
	subscript := Subscription{}

	prices := []Price{}
	productID := ""

	if len(subs.Items.Data) > 1 {
		log.Warn().Msg("customer has more than one subscription")
	}

	for _, item := range subs.Items.Data {
		productID = item.Price.Product.ID

		product, err := sc.GetProductByID(ctx, productID)
		if err != nil {
			log.Warn().Err(err).Msg("failed to get product by ID")
		}

		prices = append(prices, Price{
			ID:          item.Price.ID,
			Price:       float64(item.Price.UnitAmount) / 100, // nolint:mnd
			ProductID:   productID,
			ProductName: product.Name,
			Interval:    string(item.Price.Recurring.Interval),
			Currency:    string(item.Price.Currency),
		})

		subscript.Prices = append(subscript.Prices, prices...)
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
		Features:         subscript.Features,
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
