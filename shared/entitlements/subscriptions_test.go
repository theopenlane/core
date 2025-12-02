package entitlements_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v83"

	"github.com/theopenlane/shared/entitlements"
)

func TestIsSubscriptionActive(t *testing.T) {
	activeStatuses := []stripe.SubscriptionStatus{
		stripe.SubscriptionStatusActive,
		stripe.SubscriptionStatusTrialing,
		stripe.SubscriptionStatusPastDue,
		stripe.SubscriptionStatusIncomplete,
	}

	for _, s := range activeStatuses {
		assert.True(t, entitlements.IsSubscriptionActive(s))
	}

	inactiveStatuses := []stripe.SubscriptionStatus{
		stripe.SubscriptionStatusCanceled,
		stripe.SubscriptionStatusIncompleteExpired,
		stripe.SubscriptionStatusUnpaid,
		stripe.SubscriptionStatusPaused,
	}

	for _, s := range inactiveStatuses {
		assert.False(t, entitlements.IsSubscriptionActive(s))
	}

	assert.False(t, entitlements.IsSubscriptionActive("unknown"))
}
