package entitlements

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"
)

// CreateSubscription creates a new subscription
func (sc *StripeClient) CreateSubscription(params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	start := time.Now()
	subscription, err := sc.Client.Subscriptions.New(params)

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

// ListStripeSubscriptions lists stripe subscriptions by customer
func (sc *StripeClient) ListOrCreateSubscriptions(customerID string) (*Subscription, error) {
	i := sc.Client.Subscriptions.List(&stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
	})

	if !i.Next() {
		sub, err := sc.CreateTrialSubscription(&stripe.Customer{ID: customerID})
		if err != nil {
			log.Error().Err(err).Msg("Failed to create trial subscription")
			return nil, err
		}

		return sub, nil
	}

	// assumes customer can only have 1 subscription if there are any
	subs := sc.MapStripeSubscription(i.Subscription())

	return subs, nil
}

// GetSubscriptionByID gets a subscription by ID
func (sc *StripeClient) GetSubscriptionByID(id string) (*stripe.Subscription, error) {
	start := time.Now()

	subscription, err := sc.Client.Subscriptions.Get(id, &stripe.SubscriptionParams{
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
func (sc *StripeClient) UpdateSubscription(id string, params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	subscription, err := sc.Client.Subscriptions.Update(id, params)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// CancelSubscription cancels a subscription
func (sc *StripeClient) CancelSubscription(id string, params *stripe.SubscriptionCancelParams) (*stripe.Subscription, error) {
	subscription, err := sc.Client.Subscriptions.Cancel(id, params)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

var trialdays int64 = 30

// CreateTrialSubscription creates a trial subscription with the configured price
func (sc *StripeClient) CreateTrialSubscription(cust *stripe.Customer) (*Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(cust.ID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: &sc.Config.TrialSubscriptionPriceID,
			},
		},
		TrialPeriodDays: stripe.Int64(trialdays),
		PaymentSettings: &stripe.SubscriptionPaymentSettingsParams{
			SaveDefaultPaymentMethod: stripe.String(string(stripe.SubscriptionPaymentSettingsSaveDefaultPaymentMethodOnSubscription)),
		},
		Metadata:         cust.Metadata,
		CollectionMethod: stripe.String(string(stripe.SubscriptionCollectionMethodChargeAutomatically)),
		TrialSettings: &stripe.SubscriptionTrialSettingsParams{
			EndBehavior: &stripe.SubscriptionTrialSettingsEndBehaviorParams{
				MissingPaymentMethod: stripe.String(string(stripe.SubscriptionTrialSettingsEndBehaviorMissingPaymentMethodPause)),
			},
		},
	}

	subs, err := sc.CreateSubscription(params)
	if err != nil {
		log.Err(err).Msg("Failed to create trial subscription")
		return nil, err
	}

	log.Debug().Msgf("Created trial subscription with ID: %s", subs.ID)

	mappedsubscription := sc.MapStripeSubscription(subs)

	return mappedsubscription, nil
}

// CreatePersonalOrgFreeTierSubs creates a subscription with the configured $0 price used for personal organizations only
func (sc *StripeClient) CreatePersonalOrgFreeTierSubs(customerID string) (*Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: &sc.Config.PersonalOrgSubscriptionPriceID,
			},
		},
		CollectionMethod: stripe.String(string(stripe.SubscriptionCollectionMethodChargeAutomatically)),
	}

	subs, err := sc.CreateSubscription(params)
	if err != nil {
		log.Err(err).Msg("Failed to create trial subscription")
		return nil, err
	}

	return sc.MapStripeSubscription(subs), nil
}

// retrieveActiveEntitlements retrieves active entitlements for a customer
func (sc *StripeClient) retrieveActiveEntitlements(customerID string) ([]string, []string, error) {
	params := &stripe.EntitlementsActiveEntitlementListParams{
		Customer: stripe.String(customerID),
		Expand:   []*string{stripe.String("data.feature")},
	}

	iter := sc.Client.EntitlementsActiveEntitlements.List(params)

	feat := []string{}
	featNames := []string{}

	for iter.Next() {
		feat = append(feat, iter.EntitlementsActiveEntitlement().LookupKey)
		featNames = append(featNames, iter.EntitlementsActiveEntitlement().Feature.Name)
	}

	if iter.Err() != nil {
		log.Err(iter.Err()).Msg("failed to find active entitlements")

		return nil, nil, iter.Err()
	}

	return feat, featNames, nil
}

// MapStripeSubscription maps a stripe.Subscription to a "internal" subscription struct
func (sc *StripeClient) MapStripeSubscription(subs *stripe.Subscription) *Subscription {
	subscript := Subscription{}

	prices := []Price{}
	productID := ""

	if len(subs.Items.Data) > 1 {
		log.Warn().Msg("customer has more than one subscription")
	}

	for _, item := range subs.Items.Data {
		productID = item.Price.Product.ID

		product, err := sc.GetProductByID(productID)
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
		StartDate:        subs.CurrentPeriodStart,
		EndDate:          subs.CurrentPeriodEnd,
		TrialEnd:         subs.TrialEnd,
		ProductID:        productID,
		Status:           string(subs.Status),
		StripeCustomerID: subs.Customer.ID,
		OrganizationID:   subs.Metadata["organization_id"],
		DaysUntilDue:     subs.DaysUntilDue,
		Features:         subscript.Features,
	}
}
