package workflows

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
)

// TestOrganizationOwnerIDsEmpty verifies error behavior when client is missing
func TestOrganizationOwnerIDsEmpty(t *testing.T) {
	ids, err := OrganizationOwnerIDs(context.Background(), nil, "")
	assert.Check(t, is.ErrorIs(err, ErrNilClient))
	assert.Check(t, is.Nil(ids))
}

// TestOrganizationOwnerIDsMissingOrgID verifies error behavior when org ID is missing
func TestOrganizationOwnerIDsMissingOrgID(t *testing.T) {
	ids, err := OrganizationOwnerIDs(context.Background(), &generated.Client{}, "")
	assert.Check(t, is.ErrorIs(err, ErrMissingOrganizationID))
	assert.Check(t, is.Nil(ids))
}

// TestWorkflowMetadata verifies metadata entries are populated
func TestWorkflowMetadata(t *testing.T) {
	metadata := WorkflowMetadata()
	assert.Check(t, metadata != nil)
	assert.Check(t, len(metadata) != 0)
}
