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

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		FailNow(t)
	}

	ctx = entityops.WithEmissionVetoed(SetInternalContext(ctx, Suite.Client.DB))

	err = Suite.Client.DB.Organization.DeleteOneID(orgID).Exec(ctx)
	RequireNoError(t, err)
}
