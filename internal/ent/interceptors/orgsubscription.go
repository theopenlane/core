package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/gqlgen-plugins/graphutils"
)

func InterceptorSubscriptionURL() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.OrgSubscriptionFunc(func(ctx context.Context, q *generated.OrgSubscriptionQuery) (generated.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			hasField := false

			urlFields := []string{"subscriptionURL", "managePaymentMethods", "cancellation"}
			for _, field := range urlFields {
				if graphutils.CheckForRequestedField(ctx, field) {
					hasField = true
					break
				}
			}

			if !hasField {
				return v, nil
			}

			// cast to the expected output format
			orgSubResult, ok := v.([]*generated.OrgSubscription)
			if ok {
				for _, orgSub := range orgSubResult {
					if err := setSubscriptionURL(ctx, orgSub, q); err != nil {
						zerolog.Ctx(ctx).Warn().Err(err).Msg("failed to set subscription URL")
					}
				}

				return v, nil
			}

			// if its not a list, check the single entry
			orgSub, ok := v.(*generated.OrgSubscription)
			if ok {
				if err := setSubscriptionURL(ctx, orgSub, q); err != nil {
					zerolog.Ctx(ctx).Warn().Err(err).Msg("failed to set subscription URL")
				}

				return v, nil
			}

			return v, nil
		})
	})
}

// setSubscriptionURL sets the subscription URL for the org subscription response
func setSubscriptionURL(ctx context.Context, orgSub *generated.OrgSubscription, q *generated.OrgSubscriptionQuery) error {
	if orgSub == nil {
		return nil
	}

	// if the subscription doesn't have a stripe ID or customer ID, skip
	if orgSub.StripeSubscriptionID == "" || orgSub.StripeCustomerID == "" {
		return nil
	}

	// create a billing portal session
	updateSubscription, err := q.EntitlementManager.CreateBillingPortalUpdateSession(orgSub.StripeSubscriptionID, orgSub.StripeCustomerID)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msg("failed to create billing portal session")

		return err
	}

	cancelSubscription, err := q.EntitlementManager.CancellationBillingPortalSession(orgSub.StripeSubscriptionID, orgSub.StripeCustomerID)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msg("failed to create billing portal session")

		return err
	}

	updatePaymentMethod, err := q.EntitlementManager.CreateBillingPortalPaymentMethods(orgSub.StripeSubscriptionID, orgSub.StripeCustomerID)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msg("failed to create billing portal session")

		return err
	}

	// add the subscription URL to the result
	orgSub.SubscriptionURL = updateSubscription.ManageSubscription
	orgSub.Cancellation = cancelSubscription.Cancellation
	orgSub.ManagePaymentMethods = updatePaymentMethod.PaymentMethods

	return nil
}
