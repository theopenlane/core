package system

import (
	"context"

	"github.com/stripe/stripe-go/v84"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// activeOrTrialingSubscriptionPredicates matches organizations with an active or trialing subscription
func activeOrTrialingSubscriptionPredicates() []predicate.OrgSubscription {
	return []predicate.OrgSubscription{
		orgsubscription.Or(
			orgsubscription.ActiveEQ(true),
			orgsubscription.StripeSubscriptionStatusEQ(string(stripe.SubscriptionStatusTrialing)),
		),
	}
}

// systemSweepContext builds a cross-organization system caller context bypassing org filtering and FGA
func systemSweepContext(ctx context.Context) context.Context {
	return auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), &auth.Caller{
		Capabilities: auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation,
	})
}
