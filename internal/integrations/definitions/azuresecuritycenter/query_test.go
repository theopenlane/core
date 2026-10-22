//go:build test

package azuresecuritycenter

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"
)

const otherTenantID = "00000000-0000-0000-0000-000000000002"

func newTenantClient(t *testing.T, mock *microsoft.MockMicrosoftProvider, tenantID string) *Client {
	t.Helper()

	cfg := mock.AppConfig()

	cred, err := microsoft.NewTokenCredential(cfg, microsoft.MockTenantID)
	assert.NilError(t, err)

	armClient, err := microsoft.NewARMClient(cfg, cred)
	assert.NilError(t, err)

	return &Client{arm: armClient, tenantID: tenantID}
}

func TestSubscriptionsScopedToTenant(t *testing.T) {
	mock := microsoft.NewMockMicrosoftProvider()
	t.Cleanup(mock.Close)

	mock.SetSubscriptions(
		microsoft.MockSubscription{ID: "sub-own", State: "Enabled"},
		microsoft.MockSubscription{ID: "sub-other", TenantID: otherTenantID, State: "Enabled"},
	)

	client := newTenantClient(t, mock, microsoft.MockTenantID)

	tests := []struct {
		name      string
		requested []string
		expected  []string
		err       error
	}{
		{
			name:     "no request returns tenant subscriptions only",
			expected: []string{"sub-own"},
		},
		{
			name:      "requested subscription in tenant",
			requested: []string{"SUB-OWN"},
			expected:  []string{"SUB-OWN"},
		},
		{
			name:      "requested subscription in another tenant",
			requested: []string{"sub-own", "sub-other"},
			err:       ErrSubscriptionNotAccessible,
		},
		{
			name:      "requested subscription not visible",
			requested: []string{"sub-unknown"},
			err:       ErrSubscriptionNotAccessible,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ids, err := client.subscriptions(context.Background(), tc.requested)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				return
			}

			assert.NilError(t, err)
			assert.Check(t, len(ids) == len(tc.expected))

			for i := range tc.expected {
				assert.Check(t, ids[i] == tc.expected[i])
			}
		})
	}
}

func TestSubscriptionsLighthouseDelegation(t *testing.T) {
	mock := microsoft.NewMockMicrosoftProvider()
	t.Cleanup(mock.Close)

	managingTenantID := "00000000-0000-0000-0000-00000000aaaa"
	code := "ABC123"

	mock.SetSubscriptions(
		microsoft.MockSubscription{ID: "sub-mine", State: "Enabled"},
		microsoft.MockSubscription{ID: "sub-other-code", State: "Enabled"},
		microsoft.MockSubscription{ID: "sub-other-provider", State: "Enabled"},
	)
	mock.SetDelegations("sub-mine", microsoft.MockDelegation{OfferName: lighthouseOffer.OfferName(code), ManagedByTenantID: managingTenantID})
	mock.SetDelegations("sub-other-code", microsoft.MockDelegation{OfferName: lighthouseOffer.OfferName("ZZZ999"), ManagedByTenantID: managingTenantID})
	mock.SetDelegations("sub-other-provider", microsoft.MockDelegation{OfferName: lighthouseOffer.OfferName(code), ManagedByTenantID: otherTenantID})

	client := newTenantClient(t, mock, microsoft.MockTenantID)
	client.delegation = lighthouseDelegation{managingTenantID: managingTenantID, offerName: lighthouseOffer.OfferName(code)}

	ids, err := client.subscriptions(context.Background(), nil)
	assert.NilError(t, err)
	assert.Check(t, len(ids) == 1 && ids[0] == "sub-mine")

	_, err = client.subscriptions(context.Background(), []string{"sub-other-code"})
	assert.ErrorIs(t, err, ErrSubscriptionNotAccessible)

	client.delegation.offerName = lighthouseOffer.OfferName("NOPE")

	_, err = client.subscriptions(context.Background(), nil)
	assert.ErrorIs(t, err, ErrNoDelegatedSubscriptions)
}

func TestSubscriptionsLighthouseNoneInTenant(t *testing.T) {
	mock := microsoft.NewMockMicrosoftProvider()
	t.Cleanup(mock.Close)

	mock.SetSubscriptions(microsoft.MockSubscription{ID: "sub-other", TenantID: otherTenantID, State: "Enabled"})

	client := newTenantClient(t, mock, microsoft.MockTenantID)
	client.delegation = lighthouseDelegation{managingTenantID: "00000000-0000-0000-0000-00000000aaaa", offerName: lighthouseOffer.OfferName("ABC123")}

	_, err := client.subscriptions(context.Background(), nil)
	assert.ErrorIs(t, err, ErrNoTenantSubscriptionsDelegated)
}

func TestSubscriptionsNoneInTenant(t *testing.T) {
	mock := microsoft.NewMockMicrosoftProvider()
	t.Cleanup(mock.Close)

	mock.SetSubscriptions(microsoft.MockSubscription{ID: "sub-other", TenantID: otherTenantID, State: "Enabled"})

	client := newTenantClient(t, mock, microsoft.MockTenantID)

	_, err := client.subscriptions(context.Background(), []string{"sub-other"})
	assert.ErrorIs(t, err, ErrNoSubscriptions)
}
