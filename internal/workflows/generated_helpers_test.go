package workflows

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/ent/generated"
)

// TestOrganizationOwnerIDsEmpty verifies error behavior when client is missing
func TestOrganizationOwnerIDsEmpty(t *testing.T) {
	ids, err := OrganizationOwnerIDs(context.Background(), nil, "")
	assert.ErrorIs(t, err, ErrNilClient)
	assert.Nil(t, ids)
}

// TestOrganizationOwnerIDsMissingOrgID verifies error behavior when org ID is missing
func TestOrganizationOwnerIDsMissingOrgID(t *testing.T) {
	ids, err := OrganizationOwnerIDs(context.Background(), &generated.Client{}, "")
	assert.ErrorIs(t, err, ErrMissingOrganizationID)
	assert.Nil(t, ids)
}

// TestWorkflowMetadata verifies metadata entries are populated
func TestWorkflowMetadata(t *testing.T) {
	metadata := WorkflowMetadata()
	assert.NotNil(t, metadata)
	assert.NotEmpty(t, metadata)
}
