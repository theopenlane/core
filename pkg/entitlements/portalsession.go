package entitlements

import (
	"context"

	"github.com/stripe/stripe-go/v82"
)

// CreateBillingPortalUpdateSession generates an update session in stripe's billing portal which displays the customers current subscription tier and allows them to upgrade or downgrade
func (sc *StripeClient) CreateBillingPortalUpdateSession(ctx context.Context, subsID, custID string) (*BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionCreateParams{
		Customer:  &custID,
		ReturnURL: &sc.Config.StripeBillingPortalSuccessURL,
		FlowData: &stripe.BillingPortalSessionCreateFlowDataParams{
			Type: stripe.String("subscription_update"),
			SubscriptionUpdate: &stripe.BillingPortalSessionCreateFlowDataSubscriptionUpdateParams{
				Subscription: &subsID,
			},
		},
	}

	billingPortalSession, err := sc.Client.V1BillingPortalSessions.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &BillingPortalSession{
		ManageSubscription: billingPortalSession.URL,
	}, nil
}

// CreateBillingPortalPaymentMethods generates a session in stripe's billing portal which allows the customer to add / update payment methods
func (sc *StripeClient) CreateBillingPortalPaymentMethods(ctx context.Context, custID string) (*BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionCreateParams{
		Customer:  &custID,
		ReturnURL: &sc.Config.StripeBillingPortalSuccessURL,
		FlowData: &stripe.BillingPortalSessionCreateFlowDataParams{
			Type: stripe.String("payment_method_update"),
		},
	}

	billingPortalSession, err := sc.Client.V1BillingPortalSessions.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &BillingPortalSession{
		PaymentMethods: billingPortalSession.URL,
	}, nil
}

// CancellationBillingPortalSession generates a session in stripe's billing portal which allows the customer to cancel their subscription
func (sc *StripeClient) CancellationBillingPortalSession(ctx context.Context, subsID, custID string) (*BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionCreateParams{
		Customer:  &custID,
		ReturnURL: &sc.Config.StripeBillingPortalSuccessURL, // this is the "return back to website" URL, not a cancellation / update specific one
		FlowData: &stripe.BillingPortalSessionCreateFlowDataParams{
			Type: stripe.String("subscription_cancel"),
			SubscriptionCancel: &stripe.BillingPortalSessionCreateFlowDataSubscriptionCancelParams{
				Subscription: &subsID,
			},
			AfterCompletion: &stripe.BillingPortalSessionCreateFlowDataAfterCompletionParams{
				Type: stripe.String("redirect"),
				Redirect: &stripe.BillingPortalSessionCreateFlowDataAfterCompletionRedirectParams{
					ReturnURL: &sc.Config.StripeCancellationReturnURL,
				},
			},
		},
	}

	billingPortalSession, err := sc.Client.V1BillingPortalSessions.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &BillingPortalSession{
		Cancellation: billingPortalSession.URL,
	}, nil
}

// CreateBillingPortalAddModuleSession generates a billing portal session that allows
// a customer to add a module to their subscription. The module is identified by the provided price ID
func (sc *StripeClient) CreateBillingPortalAddModuleSession(ctx context.Context, subsID, custID, priceID string) (*BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionCreateParams{
		Customer:  &custID,
		ReturnURL: &sc.Config.StripeBillingPortalSuccessURL,
		FlowData: &stripe.BillingPortalSessionCreateFlowDataParams{
			Type: stripe.String(string(stripe.BillingPortalSessionFlowTypeSubscriptionUpdateConfirm)),
			SubscriptionUpdateConfirm: &stripe.BillingPortalSessionCreateFlowDataSubscriptionUpdateConfirmParams{
				Subscription: &subsID,
				Items: []*stripe.BillingPortalSessionCreateFlowDataSubscriptionUpdateConfirmItemParams{
					{
						Price: &priceID,
					},
				},
			},
		},
	}

	sess, err := sc.Client.V1BillingPortalSessions.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &BillingPortalSession{ManageSubscription: sess.URL}, nil
}
