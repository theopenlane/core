package graphapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionResolver_NotificationCreated(t *testing.T) {
	// Create a minimal test resolver - we just need the struct, not the full setup
	r := &Resolver{}

	sr := &subscriptionResolver{r}

	// The NotificationCreated method should panic since it's not implemented
	assert.Panics(t, func() {
		sr.NotificationCreated(context.Background())
	}, "Expected NotificationCreated to panic as it's not implemented")
}

func TestResolver_Subscription(t *testing.T) {
	// Create a minimal test resolver
	r := &Resolver{}

	// Test that Subscription() returns a subscriptionResolver
	subResolver := r.Subscription()
	assert.NotNil(t, subResolver)
	assert.IsType(t, &subscriptionResolver{}, subResolver)
}

func TestSubscriptionResolver_Type(t *testing.T) {
	// Verify the subscription resolver contains a Resolver pointer
	r := &Resolver{}

	sr := &subscriptionResolver{r}

	// Verify the embedded Resolver is accessible
	require.NotNil(t, sr.Resolver)
	assert.Equal(t, r, sr.Resolver)
}
