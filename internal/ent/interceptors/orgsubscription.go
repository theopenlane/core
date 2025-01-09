package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
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
			if !ok {
				// if its not a list, check the single entry
				orgSub, ok := v.(*generated.OrgSubscription)
				if !ok {
					return v, nil
				}

				// add to an array to keep the logic the same
				orgSubResult = []*generated.OrgSubscription{orgSub}
			}

			// pull all the queried IDs
			queriedIDs, err := q.Clone().IDs(ctx)
			if err != nil {
				return nil, err
			}

			// grab the org subscriptions, we can't use the query result because it's not guaranteed to contain the fields we need
			orgSubs, err := generated.FromContext(ctx).OrgSubscription.Query().Where(orgsubscription.IDIn(queriedIDs...)).All(ctx)
			if err != nil {
				return nil, err
			}

			// loop through the org subscriptions and create a billing portal session for each
			// realistically, this should only be one org subscription but we'll handle multiple just in case
			for _, orgSub := range orgSubs {

				// if the subscription doesn't have a stripe subscription or customer ID, skip
				if orgSub.StripeSubscriptionID == "" || orgSub.StripeCustomerID == "" {
					continue
				}

				// create a billing portal session
				portalSession, err := q.EntitlementManager.CreateBillingPortalUpdateSession(orgSub.StripeSubscriptionID, orgSub.StripeCustomerID)
				if err != nil {
					return nil, err
				}

				// add the subscription URL to the result
				for _, r := range orgSubResult {
					if r.ID == orgSub.ID {
						r.SubscriptionURL = portalSession.URL
					}
				}

			}

			// if the query is for a single org subscription, return the single result
			if len(orgSubResult) == 1 {
				return orgSubResult[0], nil
			}

			// if the query is for multiple org subscriptions, return the list
			return orgSubResult, nil
		})
	})
}
