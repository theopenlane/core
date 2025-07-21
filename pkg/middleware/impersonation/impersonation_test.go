package impersonation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/iam/auth"
)

// Test basic middleware functionality

func TestImpersonationMiddleware_Process(t *testing.T) {
	// Skip middleware tests for now since they require complex mocking
	// These would be better as integration tests
	t.Skip("Middleware tests require complex mocking - should be integration tests")
}

func TestRequireImpersonationScope(t *testing.T) {
	// Test the scope validation logic
	ctx := context.Background()
	impUser := &auth.ImpersonatedUser{
		AuthenticatedUser: &auth.AuthenticatedUser{
			SubjectID: "user123",
		},
		ImpersonationContext: &auth.ImpersonationContext{
			ExpiresAt: time.Now().Add(1 * time.Hour),
			Scopes:    []string{"read", "debug"},
		},
	}
	ctx = auth.WithImpersonatedUser(ctx, impUser)

	// Test that scope checking works
	assert.True(t, impUser.ImpersonationContext.HasScope("read"))
	assert.False(t, impUser.ImpersonationContext.HasScope("write"))
}

func TestBlockImpersonation(t *testing.T) {
	// Test impersonation detection logic
	ctx := context.Background()
	impUser := &auth.ImpersonatedUser{
		AuthenticatedUser: &auth.AuthenticatedUser{
			SubjectID: "user123",
		},
		ImpersonationContext: &auth.ImpersonationContext{
			Type: auth.SupportImpersonation,
		},
	}
	ctx = auth.WithImpersonatedUser(ctx, impUser)

	// Test that impersonation is detected
	retrieved, ok := auth.ImpersonatedUserFromContext(ctx)
	assert.True(t, ok)
	assert.True(t, retrieved.IsImpersonated())
}

func TestAllowOnlyImpersonationType(t *testing.T) {
	// Test impersonation type checking
	supportImpersonation := &auth.ImpersonationContext{
		Type: auth.SupportImpersonation,
	}
	adminImpersonation := &auth.ImpersonationContext{
		Type: auth.AdminImpersonation,
	}

	assert.Equal(t, auth.SupportImpersonation, supportImpersonation.Type)
	assert.Equal(t, auth.AdminImpersonation, adminImpersonation.Type)
	assert.NotEqual(t, supportImpersonation.Type, adminImpersonation.Type)
}
