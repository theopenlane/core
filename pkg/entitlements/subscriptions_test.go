package entitlements_test

import (
	"testing"

	"github.com/stripe/stripe-go/v87"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/pkg/entitlements"
)

func TestIsSubscriptionActive(t *testing.T) {
	activeStatuses := []stripe.SubscriptionStatus{
		stripe.SubscriptionStatusActive,
		stripe.SubscriptionStatusTrialing,
		stripe.SubscriptionStatusPastDue,
		stripe.SubscriptionStatusIncomplete,
	}

	for _, s := range activeStatuses {
		assert.Check(t, entitlements.IsSubscriptionActive(s))
	}

	inactiveStatuses := []stripe.SubscriptionStatus{
		stripe.SubscriptionStatusCanceled,
		stripe.SubscriptionStatusIncompleteExpired,
		stripe.SubscriptionStatusUnpaid,
		stripe.SubscriptionStatusPaused,
	}

	for _, s := range inactiveStatuses {
		assert.Check(t, !entitlements.IsSubscriptionActive(s))
	}

	assert.Check(t, !entitlements.IsSubscriptionActive("unknown"))
}
