package store

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

func TestGetOrgOwnerIDUsesSingleAuthorizedOrg(t *testing.T) {
	t.Parallel()

	orgID := ulids.New().String()

	ctx := auth.WithCaller(context.Background(), &auth.Caller{
		OrganizationIDs: []string{orgID},
	})

	id, err := getOrgOwnerID(ctx, objects.File{})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(orgID, id))
}
