//go:build test

package testharness

import (
	"context"
	"testing"

	"github.com/theopenlane/iam/auth"
)

// CleanupOrganizationDataWithContext removes the caller's organization
func CleanupOrganizationDataWithContext(ctx context.Context, t *testing.T) {
	t.Helper()

	caller, _ := auth.CallerFromContext(ctx)
	if caller == nil || caller.OrganizationID == "" {
		FailNow(t)
	}

	_, err := Suite.Client.API.DeleteOrganization(ctx, caller.OrganizationID)
	RequireNoError(t, err)
}
