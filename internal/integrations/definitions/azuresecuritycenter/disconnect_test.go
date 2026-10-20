//go:build test

package azuresecuritycenter

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"
)

func TestRemoveDelegations(t *testing.T) {
	mock := microsoft.NewMockMicrosoftProvider()
	t.Cleanup(mock.Close)

	managingTenantID := "00000000-0000-0000-0000-00000000aaaa"
	code := "ABC123"

	mock.SetSubscriptions(
		microsoft.MockSubscription{ID: "sub-mine", State: "Enabled"},
		microsoft.MockSubscription{ID: "sub-other-code", State: "Enabled"},
		microsoft.MockSubscription{ID: "sub-other-tenant", TenantID: otherTenantID, State: "Enabled"},
	)
	mock.SetDelegations("sub-mine", microsoft.MockDelegation{OfferName: lighthouseOffer.OfferName(code), ManagedByTenantID: managingTenantID})
	mock.SetDelegations("sub-other-code", microsoft.MockDelegation{OfferName: lighthouseOffer.OfferName("ZZZ999"), ManagedByTenantID: managingTenantID})
	mock.SetDelegations("sub-other-tenant", microsoft.MockDelegation{OfferName: lighthouseOffer.OfferName(code), ManagedByTenantID: managingTenantID})

	client := newTenantClient(t, mock, microsoft.MockTenantID)
	client.delegation = lighthouseDelegation{managingTenantID: managingTenantID, offerName: lighthouseOffer.OfferName(code)}

	removed, err := client.removeDelegations(context.Background())
	assert.NilError(t, err)
	assert.Check(t, len(removed) == 1 && removed[0] == "sub-mine")
	assert.Check(t, len(mock.DeletedDelegations()) == 1)

	removed, err = client.removeDelegations(context.Background())
	assert.NilError(t, err)
	assert.Check(t, len(removed) == 0)
}
