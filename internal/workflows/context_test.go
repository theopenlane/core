package workflows

import (
	"context"
	"testing"

	"entgo.io/ent/privacy"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

func TestWorkflowContexts(t *testing.T) {
	base := context.Background()

	bypass := WithContext(base)
	assert.True(t, IsWorkflowBypass(bypass))

	decision, ok := privacy.DecisionFromContext(AllowContext(base))
	assert.True(t, ok)
	assert.NoError(t, decision)

	orgID := ulids.New().String()
	orgCtx := auth.NewTestContextWithOrgID(ulids.New().String(), orgID)

	allowCtx, resolvedOrg, err := AllowContextWithOrg(orgCtx)
	assert.NoError(t, err)
	assert.Equal(t, orgID, resolvedOrg)
	decision, ok = privacy.DecisionFromContext(allowCtx)
	assert.True(t, ok)
	assert.NoError(t, decision)

	bypassCtx, resolvedOrg, err := AllowBypassContextWithOrg(orgCtx)
	assert.NoError(t, err)
	assert.Equal(t, orgID, resolvedOrg)
	assert.True(t, IsWorkflowBypass(bypassCtx))
	decision, ok = privacy.DecisionFromContext(bypassCtx)
	assert.True(t, ok)
	assert.NoError(t, decision)

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

	decision, ok := privacy.DecisionFromContext(allowCtx)
	assert.True(t, ok)
	assert.NoError(t, decision)
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
