package entitlements

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v86"
)

func TestBuildOnlyPricesItems(t *testing.T) {
	sub := func(items ...*stripe.SubscriptionItem) *stripe.Subscription {
		return &stripe.Subscription{Items: &stripe.SubscriptionItemList{Data: items}}
	}
	item := func(id, priceID string) *stripe.SubscriptionItem {
		return &stripe.SubscriptionItem{ID: id, Price: &stripe.Price{ID: priceID}}
	}

	testCases := []struct {
		name        string
		sub         *stripe.Subscription
		priceIDs    []string
		wantDeleted []string
		wantAdded   []string
	}{
		{
			name:        "removes the paid item and keeps the free one",
			sub:         sub(item("si_base", "price_free"), item("si_paid", "price_paid")),
			priceIDs:    []string{"price_free"},
			wantDeleted: []string{"si_paid"},
		},
		{
			// nothing is added back, every subscription already carries the free base module
			name:        "removes the paid item without adding anything",
			sub:         sub(item("si_paid", "price_paid")),
			priceIDs:    []string{"price_free"},
			wantDeleted: []string{"si_paid"},
		},
		{
			name:     "no changes when already on exactly the free prices",
			sub:      sub(item("si_base", "price_free")),
			priceIDs: []string{"price_free"},
		},
		{
			name:        "handles an item with no price",
			sub:         sub(&stripe.SubscriptionItem{ID: "si_odd"}),
			priceIDs:    []string{"price_free"},
			wantDeleted: []string{"si_odd"},
		},
		{
			name:     "nil subscription yields nothing",
			sub:      nil,
			priceIDs: []string{"price_free"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildOnlyPricesItems(tc.sub, tc.priceIDs)

			var deleted, added []string

			for _, i := range got {
				switch {
				case i.Deleted != nil && *i.Deleted:
					deleted = append(deleted, *i.ID)
				case i.Price != nil:
					added = append(added, *i.Price)
				}
			}

			assert.ElementsMatch(t, tc.wantDeleted, deleted)
			assert.ElementsMatch(t, tc.wantAdded, added)
		})
	}
}
