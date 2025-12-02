//go:build ignore

package main

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v83"
	"github.com/theopenlane/shared/entitlements"
)

// mockClient is a mock implementation of the stripeClient interface
type mockClient struct {
	taggedFrom string
	taggedTo   string
	subs       map[string][]*stripe.Subscription
	migrated   []string
}

// TagPriceMigration simulates tagging a price migration in Stripe
func (m *mockClient) TagPriceMigration(ctx context.Context, from, to string) error {
	m.taggedFrom = from
	m.taggedTo = to
	return nil
}

// ListSubscriptions simulates listing subscriptions for a customer in Stripe
func (m *mockClient) ListSubscriptions(ctx context.Context, cid string) ([]*stripe.Subscription, error) {
	return m.subs[cid], nil
}

// MigrateSubscriptionPrice simulates migrating a subscription's price in Stripe
func (m *mockClient) MigrateSubscriptionPrice(ctx context.Context, sub *stripe.Subscription, oldID, newID string) (*stripe.Subscription, error) {
	m.migrated = append(m.migrated, sub.ID)
	return sub, nil
}

// TestPriceMigrateDryRun tests the dry run functionality of the price migration command
// it should list the customers and subscriptions that would be migrated without making any changes
func TestPriceMigrateDryRun(t *testing.T) {
	client := &mockClient{subs: map[string][]*stripe.Subscription{
		"cus_1": {
			{ID: "sub_old", Items: &stripe.SubscriptionItemList{Data: []*stripe.SubscriptionItem{
				{ID: "it_old", Price: &stripe.Price{ID: "old"}},
			}}},
		},
	}}

	newClient = func(opts ...entitlements.StripeOptions) (stripeClient, error) { return client, nil }
	defer func() {
		newClient = func(opts ...entitlements.StripeOptions) (stripeClient, error) {
			return entitlements.NewStripeClient(opts...)
		}
	}()

	buf := &bytes.Buffer{}
	outWriter = buf
	defer func() { outWriter = os.Stdout }()

	app := migrationApp()
	err := app.Run(context.Background(), []string{"pricemigrate", "--old-price", "old", "--new-price", "new", "--customers", "cus_1", "--dry-run"})
	require.NoError(t, err)

	require.Empty(t, client.taggedFrom)
	require.Contains(t, buf.String(), "cus_1")
	require.Contains(t, buf.String(), "sub_old")
}

// TestPriceMigrateApply tests the price migration command when it applies changes
func TestPriceMigrateApply(t *testing.T) {
	client := &mockClient{subs: map[string][]*stripe.Subscription{
		"cus_1": {
			{ID: "sub_old", Items: &stripe.SubscriptionItemList{Data: []*stripe.SubscriptionItem{
				{ID: "it_old", Price: &stripe.Price{ID: "old"}},
			}}},
		},
	}}

	newClient = func(opts ...entitlements.StripeOptions) (stripeClient, error) { return client, nil }
	defer func() {
		newClient = func(opts ...entitlements.StripeOptions) (stripeClient, error) {
			return entitlements.NewStripeClient(opts...)
		}
	}()

	buf := &bytes.Buffer{}
	outWriter = buf
	defer func() { outWriter = os.Stdout }()

	app := migrationApp()
	err := app.Run(context.Background(), []string{"pricemigrate", "--old-price", "old", "--new-price", "new", "--customers", "cus_1", "--dry-run=false"})
	require.NoError(t, err)

	require.Equal(t, "old", client.taggedFrom)
	require.Equal(t, "new", client.taggedTo)
	require.Len(t, client.migrated, 1)
	require.Equal(t, "sub_old", client.migrated[0])
}
