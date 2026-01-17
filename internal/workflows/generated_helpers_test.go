package workflows

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrganizationOwnerIDsEmpty(t *testing.T) {
	ids, err := OrganizationOwnerIDs(context.Background(), nil, "")
	assert.NoError(t, err)
	assert.Nil(t, ids)
}

func TestWorkflowMetadata(t *testing.T) {
	metadata := WorkflowMetadata()
	assert.NotNil(t, metadata)
	assert.NotEmpty(t, metadata)
}
