package interceptors

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/gqlgen-plugins/graphutils"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/intercept"
)

// InterceptorBillingPortalURLs is an ent interceptor to fetch data from an external source (in this case stripe) and populate the URLs in the graph return response
func InterceptorBillingPortalURLs() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.OrgSubscriptionFunc(func(ctx context.Context, q *generated.OrgSubscriptionQuery) (generated.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			hasField := false

			urlFields := []string{"managePaymentMethods"}
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
					if err := setPortalURLs(ctx, orgSub, q); err != nil {
						logx.FromContext(ctx).Warn().Err(err).Msg("failed to set subscription URL")
					}
				}

				return v, nil
			}

			// if its not a list, check the single entry
			orgSub, ok := v.(*generated.OrgSubscription)
			if ok {
				if err := setPortalURLs(ctx, orgSub, q); err != nil {
					logx.FromContext(ctx).Warn().Err(err).Msg("failed to set subscription URL")
				}

				return v, nil
			}

			return v, nil
		})
	})
}

// setPortalURLs sets the subscription URL for the org subscription response
func setPortalURLs(ctx context.Context, orgSub *generated.OrgSubscription, q *generated.OrgSubscriptionQuery) error {
	if orgSub == nil || !q.EntitlementManager.Config.IsEnabled() {
		logx.FromContext(ctx).Debug().Msg("organization does not have a subscription or entitlement manager is nil, skipping URL setting")

		return nil
	}

	// if the subscription doesn't have a stripe ID
	if orgSub.StripeSubscriptionID == "" {
		return nil
	}

	client := generated.FromContext(ctx)
	if client == nil {
		logx.FromContext(ctx).Error().Msg("ent client not found in context")
		return nil
	}

	org, err := client.Organization.Get(ctx, orgSub.OwnerID)
	if err != nil {
		logx.FromContext(ctx).Err(err).Str("owner_id", orgSub.OwnerID).Msg("failed to fetch organization")
		return err
	}

	if org.StripeCustomerID == nil || *org.StripeCustomerID == "" {
		logx.FromContext(ctx).Warn().Str("owner_id", orgSub.OwnerID).Msg("organization does not have a stripe customer ID")
		return nil
	}

	customerID := *org.StripeCustomerID

	updatePaymentMethod, err := q.EntitlementManager.CreateBillingPortalPaymentMethods(ctx, customerID)
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("failed to create update payment method billing portal session type")

		return err
	}

	orgSub.ManagePaymentMethods = updatePaymentMethod.PaymentMethods

	return nil
}
