//go:build test

package testharness

import (
	"context"
	"testing"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

// CleanupOrganizationDataWithContext soft deletes the caller's organization without the cascade
// listener, the rest of the data is discarded with the test database
func CleanupOrganizationDataWithContext(ctx context.Context, t *testing.T) {
	t.Helper()

	caller, _ := auth.CallerFromContext(ctx)
	if caller == nil || caller.OrganizationID == "" {
		FailNow(t)
	}

	ctx = entityops.WithEmissionVetoed(SetContext(ctx, Suite.Client.DB))

	err := Suite.Client.DB.Organization.DeleteOneID(caller.OrganizationID).Exec(ctx)
	RequireNoError(t, err)
}
