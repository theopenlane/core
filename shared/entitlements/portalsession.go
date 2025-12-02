package entitlements

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v83"
)

// CreateBillingPortalPaymentMethods generates a session in stripe's billing portal which allows the customer to add / update payment methods
func (sc *StripeClient) CreateBillingPortalPaymentMethods(ctx context.Context, custID string) (*BillingPortalSession, error) {
	returnURL := fmt.Sprintf("%s?%s", sc.Config.StripeBillingPortalSuccessURL, "paymentupdate=complete")
	params := &stripe.BillingPortalSessionCreateParams{
		Customer:  &custID,
		ReturnURL: &sc.Config.StripeBillingPortalSuccessURL,
		FlowData: &stripe.BillingPortalSessionCreateFlowDataParams{
			Type: stripe.String("payment_method_update"),
			AfterCompletion: &stripe.BillingPortalSessionCreateFlowDataAfterCompletionParams{
				Redirect: &stripe.BillingPortalSessionCreateFlowDataAfterCompletionRedirectParams{
					ReturnURL: &returnURL,
				},
				Type: stripe.String("redirect"),
			},
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
