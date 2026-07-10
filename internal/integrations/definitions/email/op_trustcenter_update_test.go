package email

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTrustCenterUpdateRegistered verifies the trust center update entry registers under its own
// unique identifier, distinct from the customer-selectable branded message key, so templates keyed
// by it resolve a dispatcher
func TestTrustCenterUpdateRegistered(t *testing.T) {
	assert.NotEqual(t, BrandedMessageOp.Name(), TrustCenterUpdateTemplate)

	_, ok := DispatcherByKey(TrustCenterUpdateTemplate)
	assert.True(t, ok)
}

// TestTrustCenterUpdateNotCustomerSelectable verifies the entry is excluded from the customer-facing
// catalog: the per-trust-center template is seeded and customized, never authored from the picker
func TestTrustCenterUpdateNotCustomerSelectable(t *testing.T) {
	d, ok := DispatcherByKey(TrustCenterUpdateTemplate)
	assert.True(t, ok)

	cs := d.Registration().CustomerSelectable
	assert.NotNil(t, cs)
	assert.False(t, *cs)

	for _, sel := range CustomerSelectableDispatchers() {
		assert.NotEqual(t, TrustCenterUpdateTemplate, sel.Name())
	}
}

// TestTrustCenterUpdateRendersAsBrandedMessage verifies the entry shares the branded message
// renderer: the same payload produces identical output through either dispatcher
func TestTrustCenterUpdateRendersAsBrandedMessage(t *testing.T) {
	branded, ok := DispatcherByKey(brandedMessageKey)
	assert.True(t, ok)

	update, ok := DispatcherByKey(TrustCenterUpdateTemplate)
	assert.True(t, ok)

	client := &Client{Config: *MockRuntimeConfig()}

	payload, err := json.Marshal(BrandedMessageRequest{
		RecipientInfo: TestRecipient("dolores@example.com"),
		Subject:       "SecureCorp update",
		Title:         "An update from SecureCorp",
		Intros:        []string{"We have news to share."},
		ButtonText:    "View trust center",
		ButtonLink:    "https://securecorp.example.com/trust",
	})
	assert.NoError(t, err)

	fromBranded, err := branded.RenderMessage(context.Background(), client, payload)
	assert.NoError(t, err)

	fromUpdate, err := update.RenderMessage(context.Background(), client, payload)
	assert.NoError(t, err)

	assert.NotEmpty(t, fromUpdate.GetHTML())
	assert.Equal(t, fromBranded.GetHTML(), fromUpdate.GetHTML())
}
