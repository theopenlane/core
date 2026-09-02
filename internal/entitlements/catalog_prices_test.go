package entitlements

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/pkg/catalog/gencatalog"
)

func TestFreeMonthlyPriceIDs(t *testing.T) {
	for _, useSandbox := range []bool{false, true} {
		free := FreeMonthlyPriceIDs(useSandbox)
		assert.Assert(t, len(free) > 0, "expected at least one free monthly price")

		freeSet := map[string]bool{}
		for _, id := range free {
			freeSet[id] = true
		}

		// every returned price must actually cost nothing, an organization downgraded onto a
		// priced module would keep accruing a balance it is already failing to pay
		for _, f := range gencatalog.GetModules(useSandbox) {
			for _, p := range f.Billing.Prices {
				if p.Interval != "month" || !freeSet[p.PriceID] {
					continue
				}

				assert.Check(t, is.Equal(int64(0), p.UnitAmount), "price %s is not free", p.PriceID)
			}
		}

		// and a module with any priced monthly tier must not be offered as free
		for _, f := range gencatalog.GetModules(useSandbox) {
			for _, p := range f.Billing.Prices {
				if p.Interval == "month" && p.UnitAmount > 0 {
					assert.Check(t, !freeSet[p.PriceID], "priced module %s offered as free", p.PriceID)
				}
			}
		}
	}
}
