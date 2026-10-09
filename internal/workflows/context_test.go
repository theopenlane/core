package workflows

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

func TestWorkflowContexts(t *testing.T) {
	base := context.Background()

	bypass := WithContext(base)
	assert.True(t, IsWorkflowBypass(bypass))

	assert.True(t, auth.IsInternalRequest(auth.WithInternalCrossOrgContext(base)))

	orgID := ulids.New().String()
	orgCtx := auth.NewTestContextWithOrgID(ulids.New().String(), orgID)

	allowCtx, resolvedOrg, err := AllowContextWithOrg(orgCtx)
	assert.NoError(t, err)
	assert.Equal(t, orgID, resolvedOrg)
	assert.True(t, auth.IsInternalRequest(allowCtx))

	bypassCtx, resolvedOrg, err := AllowBypassContextWithOrg(orgCtx)
	assert.NoError(t, err)
	assert.Equal(t, orgID, resolvedOrg)
	assert.True(t, IsWorkflowBypass(bypassCtx))
	assert.True(t, auth.IsInternalRequest(bypassCtx))

	_, _, err = AllowContextWithOrg(base)
	assert.Error(t, err)
}

func TestAllowContextWithOrg_SingleAuthorizedOrgFallback(t *testing.T) {
	orgID := ulids.New().String()
	ctx := auth.WithCaller(context.Background(), &auth.Caller{
		SubjectID:       ulids.New().String(),
		OrganizationIDs: []string{orgID},
	})

	allowCtx, resolvedOrg, err := AllowContextWithOrg(ctx)
	assert.NoError(t, err)
	assert.Equal(t, orgID, resolvedOrg)
	assert.True(t, auth.IsInternalRequest(allowCtx))
}

func TestAllowContextWithOrg_MultipleAuthorizedOrgsWithoutSelection(t *testing.T) {
	ctx := auth.WithCaller(context.Background(), &auth.Caller{
		SubjectID:       ulids.New().String(),
		OrganizationIDs: []string{ulids.New().String(), ulids.New().String()},
	})

	_, _, err := AllowContextWithOrg(ctx)
	assert.Error(t, err)
}

func TestAllowContextWithOrg_EmptyAuthorizedOrgs(t *testing.T) {
	ctx := auth.WithCaller(context.Background(), &auth.Caller{
		SubjectID:       ulids.New().String(),
		OrganizationIDs: []string{},
	})

	_, _, err := AllowContextWithOrg(ctx)
	assert.Error(t, err)
}

func TestAllowContextForOrgSeedsCaller(t *testing.T) {
	orgID := ulids.New().String()

	allowCtx := AllowContextForOrg(context.Background(), orgID)

	caller, ok := auth.CallerFromContext(allowCtx)
	assert.True(t, ok)
	assert.NotNil(t, caller)
	assert.Equal(t, orgID, caller.OrganizationID)
	assert.Contains(t, caller.OrganizationIDs, orgID)
	assert.True(t, caller.Has(auth.CapInternalOperation))
}
