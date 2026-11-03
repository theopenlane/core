package workflows

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

func TestWorkflowContexts(t *testing.T) {
	base := context.Background()

	bypass := WithContext(base)
	assert.Check(t, IsWorkflowBypass(bypass))

	assert.Check(t, auth.IsInternalRequest(auth.WithInternalCrossOrgContext(base)))

	orgID := ulids.New().String()
	orgCtx := auth.NewTestContextWithOrgID(ulids.New().String(), orgID)

	allowCtx, resolvedOrg, err := AllowContextWithOrg(orgCtx)
	assert.Check(t, err)
	assert.Check(t, is.Equal(orgID, resolvedOrg))
	assert.Check(t, auth.IsInternalRequest(allowCtx))

	bypassCtx, resolvedOrg, err := AllowBypassContextWithOrg(orgCtx)
	assert.Check(t, err)
	assert.Check(t, is.Equal(orgID, resolvedOrg))
	assert.Check(t, IsWorkflowBypass(bypassCtx))
	assert.Check(t, auth.IsInternalRequest(bypassCtx))

	_, _, err = AllowContextWithOrg(base)
	assert.Check(t, err != nil)
}

func TestAllowContextWithOrg_SingleAuthorizedOrgFallback(t *testing.T) {
	orgID := ulids.New().String()
	ctx := auth.WithCaller(context.Background(), &auth.Caller{
		SubjectID:       ulids.New().String(),
		OrganizationIDs: []string{orgID},
	})

	allowCtx, resolvedOrg, err := AllowContextWithOrg(ctx)
	assert.Check(t, err)
	assert.Check(t, is.Equal(orgID, resolvedOrg))
	assert.Check(t, auth.IsInternalRequest(allowCtx))
}

func TestAllowContextWithOrg_MultipleAuthorizedOrgsWithoutSelection(t *testing.T) {
	ctx := auth.WithCaller(context.Background(), &auth.Caller{
		SubjectID:       ulids.New().String(),
		OrganizationIDs: []string{ulids.New().String(), ulids.New().String()},
	})

	_, _, err := AllowContextWithOrg(ctx)
	assert.Check(t, err != nil)
}

func TestAllowContextWithOrg_EmptyAuthorizedOrgs(t *testing.T) {
	ctx := auth.WithCaller(context.Background(), &auth.Caller{
		SubjectID:       ulids.New().String(),
		OrganizationIDs: []string{},
	})

	_, _, err := AllowContextWithOrg(ctx)
	assert.Check(t, err != nil)
}

func TestAllowContextForOrgSeedsCaller(t *testing.T) {
	orgID := ulids.New().String()

	allowCtx := AllowContextForOrg(context.Background(), orgID)

	caller, ok := auth.CallerFromContext(allowCtx)
	assert.Check(t, ok)
	assert.Check(t, caller != nil)
	assert.Check(t, is.Equal(orgID, caller.OrganizationID))
	assert.Check(t, is.Contains(caller.OrganizationIDs, orgID))
	assert.Check(t, caller.Has(auth.CapInternalOperation))
}
