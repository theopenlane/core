package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"

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

			// get the fields that were queried and check for the SubscriptionURL field
			fields := graphql.CollectFieldsCtx(ctx, []string{"SubscriptionURL"})

			// if the SubscriptionURL field wasn't queried, return the result as is
			if len(fields) == 0 {
				return v, nil
			}

			// cast to the expected output format
			orgSubResult, ok := v.([]*generated.OrgSubscription)
			if ok {
				for _, orgSub := range orgSubResult {
					setSubscriptionURL(orgSub, q) // nolint:errcheck
				}
			}

			// if its not a list, check the single entry
			orgSub, ok := v.(*generated.OrgSubscription)
			if !ok {
				setSubscriptionURL(orgSub, q) // nolint:errcheck
			}

			return v, nil
		})
	})
}

func setSubscriptionURL(orgSub *generated.OrgSubscription, q *generated.OrgSubscriptionQuery) error {
	if orgSub == nil {
		return nil
	}

	// if the subscription doesn't have a stripe subscription or customer ID, skip
	if orgSub.StripeSubscriptionID == "" || orgSub.StripeCustomerID == "" {
		return nil
	}

	// create a billing portal session
	portalSession, err := q.EntitlementManager.CreateBillingPortalUpdateSession(orgSub.StripeSubscriptionID, orgSub.StripeCustomerID)
	if err != nil {
		log.Err(err).Msg("failed to create billing portal session")

		return err
	}

	// add the subscription URL to the result
	orgSub.SubscriptionURL = portalSession.URL

	return nil
}
