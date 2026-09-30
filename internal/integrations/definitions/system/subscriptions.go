package system

import (
	"github.com/stripe/stripe-go/v86"

	"github.com/theopenlane/core/v2/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
)

// activeOrTrialingSubscriptionPredicates matches active or trialing subscriptions
func activeOrTrialingSubscriptionPredicates() []predicate.OrgSubscription {
	return []predicate.OrgSubscription{
		orgsubscription.DeletedAtIsNil(),
		orgsubscription.Or(
			orgsubscription.ActiveEQ(true),
			orgsubscription.StripeSubscriptionStatusEQ(string(stripe.SubscriptionStatusTrialing)),
		),
	}
}
