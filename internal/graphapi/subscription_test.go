package graphapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionResolver_NotificationCreated(t *testing.T) {
	// Create a resolver with subscription manager
	r := &Resolver{}
	r = r.WithSubscriptions(true)

	sr := &subscriptionResolver{r}

	// Create a context with a user ID
	ctx := context.Background()
	// Note: This test requires auth.GetSubjectIDFromContext to work
	// In a real test, you would set up proper auth context
	// For now, we just verify the function doesn't panic and returns an error for missing auth
	ch, err := sr.NotificationCreated(ctx)

	// Without proper auth context, we expect an error
	assert.Error(t, err)
	assert.Nil(t, ch)
	assert.Contains(t, err.Error(), "failed to get user ID from context")
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
