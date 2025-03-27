package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/rs/zerolog"
	"github.com/theopenlane/gqlgen-plugins/graphutils"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
)

func InterceptorSubscriptionURL() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.OrgSubscriptionFunc(func(ctx context.Context, q *generated.OrgSubscriptionQuery) (generated.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			// get the fields that were queried and check for the subscriptionURL field
			fields := graphutils.CheckForRequestedField(ctx, "subscriptionURL")

			// if the SubscriptionURL field wasn't queried, return the result as is
			if !fields {
				return v, nil
			}

			// cast to the expected output format
			orgSubResult, ok := v.([]*generated.OrgSubscription)
			if ok {
				for _, orgSub := range orgSubResult {
					if err := setSubscriptionURL(orgSub, q); err != nil {
						zerolog.Ctx(ctx).Warn().Err(err).Msg("failed to set subscription URL")
					}
				}

				return v, nil
			}

			// if its not a list, check the single entry
			orgSub, ok := v.(*generated.OrgSubscription)
			if ok {
				if err := setSubscriptionURL(orgSub, q); err != nil {
					zerolog.Ctx(ctx).Warn().Err(err).Msg("failed to set subscription URL")
				}

				return v, nil
			}

			return v, nil
		})
	})
}

// setSubscriptionURL sets the subscription URL for the org subscription response
func setSubscriptionURL(orgSub *generated.OrgSubscription, q *generated.OrgSubscriptionQuery) error {
	if orgSub == nil {
		return nil
	}

	// if the subscription doesn't have a stripe ID or customer ID, skip
	if orgSub.StripeSubscriptionID == "" || orgSub.StripeCustomerID == "" {
		return nil
	}

	// create a billing portal session
	portalSession, err := q.EntitlementManager.CreateBillingPortalUpdateSession(orgSub.StripeSubscriptionID, orgSub.StripeCustomerID)
	if err != nil {
		return err
	}

	// add the subscription URL to the result
	orgSub.SubscriptionURL = portalSession.URL

	return nil
}
