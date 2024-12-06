package entitlements

import (
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"
)

// CreateSubscription creates a new subscription
func (sc *StripeClient) CreateSubscription(params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	subscription, err := sc.Client.Subscriptions.New(params)
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
		sub, err := sc.CreateTrialSubscription(customerID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create trial subscription")
			return nil, err
		}

		return sub, nil
	}

	// assumes customer can only have 1 subscription if there are any
	subs := sc.mapStripeSubscription(i.Subscription())

	return subs, nil
}

// GetSubscriptionByID gets a subscription by ID
func (sc *StripeClient) GetSubscriptionByID(id string) (*stripe.Subscription, error) {
	subscription, err := sc.Client.Subscriptions.Get(id, &stripe.SubscriptionParams{
		Params: stripe.Params{
			Expand: []*string{stripe.String("customer")},
		},
	})
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
func (sc *StripeClient) CreateTrialSubscription(customerID string) (*Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: &sc.config.TrialSubscriptionPriceID,
			},
		},
		TrialPeriodDays: stripe.Int64(trialdays),
		PaymentSettings: &stripe.SubscriptionPaymentSettingsParams{
			SaveDefaultPaymentMethod: stripe.String(string(stripe.SubscriptionPaymentSettingsSaveDefaultPaymentMethodOnSubscription)),
		},
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

	mappedsubscription := sc.mapStripeSubscription(subs)

	return mappedsubscription, nil
}

// CreateBillingPortalUpdateSession generates an update session in stripe's billing portal which displays the customers current subscription tier and allows them to upgrade or downgrade
func (sc *StripeClient) CreateBillingPortalUpdateSession(subsID, custID string) (Checkout, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  &custID,
		ReturnURL: &sc.config.StripeBillingPortalSuccessURL,
		FlowData: &stripe.BillingPortalSessionFlowDataParams{
			Type: stripe.String("subscription_update"),
			SubscriptionUpdate: &stripe.BillingPortalSessionFlowDataSubscriptionUpdateParams{
				Subscription: &subsID,
			},
		},
	}

	billingPortalSession, err := sc.Client.BillingPortalSessions.New(params)
	if err != nil {
		return Checkout{}, err
	}

	return Checkout{
		ID:  billingPortalSession.ID,
		URL: billingPortalSession.URL,
	}, nil
}

// mapStripeSubscription maps a stripe.Subscription to a "internal" subscription struct
func (sc *StripeClient) mapStripeSubscription(subs *stripe.Subscription) *Subscription {
	subscript := Subscription{}

	var prices []Price

	for _, item := range subs.Items.Data {
		prices = append(prices, Price{
			ID:        item.Price.ID,
			Price:     float64(item.Price.UnitAmount),
			ProductID: item.Price.Product.ID,
			Interval:  string(item.Price.Recurring.Interval),
		})
	}

	for _, product := range prices {
		prodFeat := sc.GetProductFeatures(product.ProductID)
		for _, feature := range prodFeat {
			featureList := []Feature{
				{
					ID:               feature.FeatureID,
					ProductFeatureID: feature.ProductFeatureID,
					Name:             feature.Name,
					Lookupkey:        feature.Lookupkey,
				},
			}

			subscript.Features = append(subscript.Features, featureList...)
		}
	}

	subscription := &Subscription{
		ID:               subs.ID,
		Prices:           prices,
		StartDate:        subs.CurrentPeriodStart,
		EndDate:          subs.CurrentPeriodEnd,
		TrialEnd:         subs.TrialEnd,
		Status:           string(subs.Status),
		StripeCustomerID: subs.Customer.ID,
		OrganizationID:   subs.Metadata["organization_id"],
		DaysUntilDue:     subs.DaysUntilDue,
		Features:         subscript.Features,
	}

	return subscription
}

type Subs struct {
	SubsID string
	Prices []Price
}
