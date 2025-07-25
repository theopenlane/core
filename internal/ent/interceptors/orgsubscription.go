package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/gqlgen-plugins/graphutils"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	gc "github.com/theopenlane/core/pkg/catalog/gencatalog"
)

// InterceptorSubscriptionURL is an ent interceptor to fetch data from an external source (in this case stripe) and populate the URLs in the graph return response
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
	if orgSub == nil || q.EntitlementManager == nil {
		log.Debug().Ctx(ctx).Msg("organization does not have a subscription or entitlement manager is nil, skipping URL setting")

		return nil
	}

	// if the subscription doesn't have a stripe ID or customer ID, skip
	if orgSub.StripeSubscriptionID == "" || orgSub.StripeCustomerID == "" {
		return nil
	}

	// create a billing portal session
	updateSubscription, err := q.EntitlementManager.CreateBillingPortalUpdateSession(ctx, orgSub.StripeSubscriptionID, orgSub.StripeCustomerID)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msg("failed to create update subscription billing portal session type")

		return err
	}

	cancelSubscription, err := q.EntitlementManager.CancellationBillingPortalSession(ctx, orgSub.StripeSubscriptionID, orgSub.StripeCustomerID)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msg("failed to create cancel subscription billing portal session type")

		return err
	}

	updatePaymentMethod, err := q.EntitlementManager.CreateBillingPortalPaymentMethods(ctx, orgSub.StripeCustomerID)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msg("failed to create update payment method billing portal session type")

		return err
	}

	moduleURLs := map[string]string{}
	visible := gc.DefaultCatalog.Visible("")
	for name, feat := range visible.Modules {
		if len(feat.Billing.Prices) == 0 {
			continue
		}

		priceID := feat.Billing.Prices[0].PriceID
		if priceID == "" {
			continue
		}

		sess, err := q.EntitlementManager.CreateBillingPortalAddModuleSession(ctx, orgSub.StripeSubscriptionID, orgSub.StripeCustomerID, priceID)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Str("module", name).Msg("failed to create module billing portal session")
			continue
		}

		moduleURLs[name] = sess.ManageSubscription
	}

	// add the subscription URL to the result
	orgSub.SubscriptionURL = updateSubscription.ManageSubscription
	orgSub.Cancellation = cancelSubscription.Cancellation
	orgSub.ManagePaymentMethods = updatePaymentMethod.PaymentMethods
	orgSub.ModuleBillingURLs = moduleURLs

	return nil
}
