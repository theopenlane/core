package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
)

func TestResolveOwnerID(t *testing.T) {
	ownerID, err := ResolveOwnerID(context.Background(), "owner-1")
	assert.NilError(t, err)
	assert.Check(t, ownerID == "owner-1")

	orgID := ulids.New().String()
	ctx := auth.NewTestContextWithOrgID(ulids.New().String(), orgID)
	ownerID, err = ResolveOwnerID(ctx, "")
	assert.NilError(t, err)
	assert.Check(t, ownerID == orgID)

	_, err = ResolveOwnerID(context.Background(), "")
	assert.Check(t, errors.Is(err, auth.ErrNoAuthUser))
}
