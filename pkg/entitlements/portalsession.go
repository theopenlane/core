package entitlements

import "github.com/stripe/stripe-go/v82"

// CreateBillingPortalUpdateSession generates an update session in stripe's billing portal which displays the customers current subscription tier and allows them to upgrade or downgrade
func (sc *StripeClient) CreateBillingPortalUpdateSession(subsID, custID string) (*BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  &custID,
		ReturnURL: &sc.Config.StripeBillingPortalSuccessURL,
		FlowData: &stripe.BillingPortalSessionFlowDataParams{
			Type: stripe.String("subscription_update"),
			SubscriptionUpdate: &stripe.BillingPortalSessionFlowDataSubscriptionUpdateParams{
				Subscription: &subsID,
			},
		},
	}

	billingPortalSession, err := sc.Client.BillingPortalSessions.New(params)
	if err != nil {
		return nil, err
	}

	return &BillingPortalSession{
		ManageSubscription: billingPortalSession.URL,
	}, nil
}

// CreateBillingPortalPaymentMethods generates a session in stripe's billing portal which allows the customer to add / update payment methods
func (sc *StripeClient) CreateBillingPortalPaymentMethods(custID string) (*BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  &custID,
		ReturnURL: &sc.Config.StripeBillingPortalSuccessURL,
		FlowData: &stripe.BillingPortalSessionFlowDataParams{
			Type: stripe.String("payment_method_update"),
		},
	}

	billingPortalSession, err := sc.Client.BillingPortalSessions.New(params)
	if err != nil {
		return nil, err
	}

	return &BillingPortalSession{
		PaymentMethods: billingPortalSession.URL,
	}, nil
}

// CancellationBillingPortalSession generates a session in stripe's billing portal which allows the customer to cancel their subscription
func (sc *StripeClient) CancellationBillingPortalSession(subsID, custID string) (*BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  &custID,
		ReturnURL: &sc.Config.StripeBillingPortalSuccessURL, // this is the "return back to website" URL, not a cancellation / update specific one
		FlowData: &stripe.BillingPortalSessionFlowDataParams{
			Type: stripe.String("subscription_cancel"),
			SubscriptionCancel: &stripe.BillingPortalSessionFlowDataSubscriptionCancelParams{
				Subscription: &subsID,
			},
			AfterCompletion: &stripe.BillingPortalSessionFlowDataAfterCompletionParams{
				Type: stripe.String("redirect"),
				Redirect: &stripe.BillingPortalSessionFlowDataAfterCompletionRedirectParams{
					ReturnURL: &sc.Config.StripeCancellationReturnURL,
				},
			},
		},
	}

	billingPortalSession, err := sc.Client.BillingPortalSessions.New(params)
	if err != nil {
		return nil, err
	}

	return &BillingPortalSession{
		Cancellation: billingPortalSession.URL,
	}, nil
}
