package hooks_test

import (
	"context"
	"testing"

	"entgo.io/ent"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/hooks"
)

func (suite *HookTestSuite) TestAddOrDeleteStandardTuple() {
	t := suite.T()

	user := suite.seedSystemAdmin()
	if len(user.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := user.Edges.OrgMemberships[0].ID

	ctx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	ctx = generated.NewContext(ctx, suite.client)

	// soft delete context
	softDeleteCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	softDeleteCtx = generated.NewContext(softDeleteCtx, suite.client)
	softDeleteCtx = entx.IsSoftDelete(softDeleteCtx, generated.TypeStandard)

	tests := []struct {
		name           string
		mutation       *generated.StandardMutation
		ctx            context.Context
		expectedAdd    bool
		expectedDelete bool
		expectedErr    error
	}{
		{
			name: "Create with systemOwned and isPublic true",
			mutation: func() *generated.StandardMutation {
				m := generated.StandardMutation{}
				m.SetOp(ent.OpCreate)
				m.SetSystemOwned(true)
				m.SetIsPublic(true)
				return &m
			}(),
			ctx:            ctx,
			expectedAdd:    true,
			expectedDelete: false,
			expectedErr:    nil,
		},
		{
			name: "Create with systemOwned false and isPublic true",
			mutation: func() *generated.StandardMutation {
				m := generated.StandardMutation{}
				m.SetOp(ent.OpCreate)
				m.SetSystemOwned(false)
				m.SetIsPublic(true)
				return &m
			}(),
			ctx:            ctx,
			expectedAdd:    false,
			expectedDelete: false,
			expectedErr:    nil,
		},
		{
			name: "Create with systemOwned false and isPublic false",
			mutation: func() *generated.StandardMutation {
				m := generated.StandardMutation{}
				m.SetOp(ent.OpCreate)
				m.SetSystemOwned(false)
				m.SetIsPublic(true)
				return &m
			}(),
			ctx:            ctx,
			expectedAdd:    false,
			expectedDelete: false,
			expectedErr:    nil,
		},
		{
			name: "Delete operation",
			mutation: func() *generated.StandardMutation {
				m := generated.StandardMutation{}
				m.SetOp(ent.OpDelete)
				return &m
			}(),
			expectedAdd:    false,
			expectedDelete: true,
			expectedErr:    nil,
		},
		{
			name: "UpdateOne with isPublic cleared",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).SetSystemOwned(true).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.UpdateOne(std).ClearIsPublic()

				err = update.Exec(ctx)
				require.NoError(t, err)

				return update.Mutation()
			}(),
			ctx:            ctx,
			expectedAdd:    false,
			expectedDelete: true,
			expectedErr:    nil,
		},
		{
			name: "UpdateOne with isPublic cleared, will try to re-delete",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.UpdateOne(std).ClearIsPublic()

				err = update.Exec(ctx)
				require.NoError(t, err)

				return update.Mutation()
			}(),
			ctx:            ctx,
			expectedAdd:    false,
			expectedDelete: true,
			expectedErr:    nil,
		},
		{
			name: "UpdateOne with is public true, already system owned",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).SetSystemOwned(true).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.UpdateOne(std).SetIsPublic(true)

				err = update.Exec(ctx)
				require.NoError(t, err)

				return update.Mutation()
			}(),
			ctx:            ctx,
			expectedAdd:    true,
			expectedDelete: false,
			expectedErr:    nil,
		},
		{
			name: "Update with public true",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.Update().SetIsPublic(true).Where(standard.ID(std.ID))

				err = update.Exec(ctx)
				require.NoError(t, err)

				return update.Mutation()
			}(),
			ctx:            ctx,
			expectedAdd:    true,
			expectedDelete: false,
			expectedErr:    nil,
		},
		{
			name: "Update with is public false",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).SetIsPublic(true).SetSystemOwned(true).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.Update().SetIsPublic(false).Where(standard.ID(std.ID))

				err = update.Exec(ctx)
				require.NoError(t, err)

				return update.Mutation()
			}(),
			ctx:            ctx,
			expectedAdd:    false,
			expectedDelete: true,
			expectedErr:    nil,
		},
		{
			name: "Update with public true, already system owned",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).SetSystemOwned(true).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.Update().SetIsPublic(true).Where(standard.ID(std.ID))

				err = update.Exec(ctx)
				require.NoError(t, err)

				return update.Mutation()
			}(),
			ctx:            ctx,
			expectedAdd:    true,
			expectedDelete: false,
			expectedErr:    nil,
		},
		{
			name: "UpdateOne with system owned, with soft delete",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).SetSystemOwned(true).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.UpdateOne(std)
				err = update.Exec(softDeleteCtx)
				require.NoError(t, err)

				m := update.Mutation()
				return m
			}(),
			ctx:            softDeleteCtx,
			expectedAdd:    false,
			expectedDelete: true,
			expectedErr:    nil,
		},
		{
			name: "UpdateOne with soft delete",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.UpdateOne(std)
				err = update.Exec(softDeleteCtx)
				require.NoError(t, err)

				m := update.Mutation()
				return m
			}(),
			ctx:            softDeleteCtx,
			expectedAdd:    false,
			expectedDelete: true,
			expectedErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			add, delete, err := hooks.AddOrDeletePublicStandardTuple(tt.ctx, tt.mutation)
			assert.Equal(t, tt.expectedAdd, add)
			assert.Equal(t, tt.expectedDelete, delete)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func (suite *HookTestSuite) TestHookStandardDelete_InUseByTrustcenter() {
	t := suite.T()

	user := suite.seedSystemAdmin()
	if len(user.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := user.Edges.OrgMemberships[0].OrganizationID

	ctx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	ctx = generated.NewContext(ctx, suite.client)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	std, err := suite.client.Standard.Create().
		SetName(gofakeit.Name()).
		Save(allowCtx)
	require.NoError(t, err)

	trustCenter, err := suite.client.TrustCenter.Create().
		SetOwnerID(orgID).
		Save(allowCtx)
	require.NoError(t, err)

	_, err = suite.client.TrustCenterCompliance.Create().
		SetStandardID(std.ID).
		SetTrustCenterID(trustCenter.ID).
		Save(allowCtx)
	require.NoError(t, err)

	softDeleteCtx := entx.IsSoftDelete(ctx, generated.TypeStandard)

	err = suite.client.Standard.UpdateOneID(std.ID).Exec(softDeleteCtx)
	require.Error(t, err)
	assert.ErrorIs(t, err, hooks.ErrStandardInUseByTrustCenter)
}

func (suite *HookTestSuite) TestHookStandardDelete_InUseByControls() {
	t := suite.T()

	regularUser := suite.seedUser()
	if len(regularUser.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := regularUser.Edges.OrgMemberships[0].OrganizationID

	regularUserCtx := auth.NewTestContextWithOrgID(regularUser.ID, orgID)
	regularUserCtx = generated.NewContext(regularUserCtx, suite.client)
	allowCtx := privacy.DecisionContext(regularUserCtx, privacy.Allow)

	std, err := suite.client.Standard.Create().
		SetName(gofakeit.Name()).
		Save(allowCtx)
	require.NoError(t, err)

	_, err = suite.client.Control.Create().
		SetRefCode(gofakeit.UUID()).
		SetTitle(gofakeit.HipsterSentence()).
		SetStandardID(std.ID).
		SetSource(enums.ControlSourceFramework).
		SetOwnerID(orgID).
		Save(allowCtx)
	require.NoError(t, err)

	systemAdmin := suite.seedSystemAdmin()
	adminCtx := auth.NewTestContextForSystemAdmin(systemAdmin.ID, orgID)
	adminCtx = generated.NewContext(adminCtx, suite.client)
	softDeleteCtx := entx.IsSoftDelete(adminCtx, generated.TypeStandard)

	err = suite.client.Standard.UpdateOneID(std.ID).Exec(softDeleteCtx)
	require.Error(t, err)
	assert.ErrorIs(t, err, hooks.ErrStandardInUseByControls)
}

func (suite *HookTestSuite) TestHookStandardDelete_AllowsWhenNotInUse() {
	t := suite.T()

	user := suite.seedSystemAdmin()
	if len(user.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := user.Edges.OrgMemberships[0].OrganizationID

	ctx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	ctx = generated.NewContext(ctx, suite.client)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	std, err := suite.client.Standard.Create().
		SetName(gofakeit.Name()).
		Save(allowCtx)
	require.NoError(t, err)

	softDeleteCtx := entx.IsSoftDelete(allowCtx, generated.TypeStandard)

	err = suite.client.Standard.UpdateOneID(std.ID).Exec(softDeleteCtx)
	require.NoError(t, err)

	deletedStd, err := suite.client.Standard.Query().
		Where(standard.ID(std.ID)).
		Only(allowCtx)
	require.NoError(t, err)
	assert.NotNil(t, deletedStd.DeletedAt)
}

func (suite *HookTestSuite) TestHookStandardDelete_RegularUserCanDeleteNonSystemOwned() {
	t := suite.T()

	regularUser := suite.seedUser()
	if len(regularUser.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := regularUser.Edges.OrgMemberships[0].OrganizationID

	regularUserCtx := auth.NewTestContextWithOrgID(regularUser.ID, orgID)
	regularUserCtx = generated.NewContext(regularUserCtx, suite.client)
	allowCtx := privacy.DecisionContext(regularUserCtx, privacy.Allow)

	std, err := suite.client.Standard.Create().
		SetName(gofakeit.Name()).
		SetSystemOwned(false).
		Save(allowCtx)
	require.NoError(t, err)

	softDeleteCtx := entx.IsSoftDelete(allowCtx, generated.TypeStandard)

	err = suite.client.Standard.UpdateOneID(std.ID).Exec(softDeleteCtx)
	require.NoError(t, err)

	deletedStd, err := suite.client.Standard.Query().
		Where(standard.ID(std.ID)).
		Only(allowCtx)
	require.NoError(t, err)
	assert.NotNil(t, deletedStd.DeletedAt)
}

func (suite *HookTestSuite) TestHookStandardDelete_BlocksPublicSystemOwned() {
	t := suite.T()

	systemAdmin := suite.seedSystemAdmin()
	if len(systemAdmin.Edges.OrgMemberships) == 0 {
		t.Fatal("system admin has no org memberships")
	}

	orgID := systemAdmin.Edges.OrgMemberships[0].OrganizationID

	adminCtx := auth.NewTestContextForSystemAdmin(systemAdmin.ID, orgID)
	adminCtx = generated.NewContext(adminCtx, suite.client)
	allowCtx := privacy.DecisionContext(adminCtx, privacy.Allow)

	std, err := suite.client.Standard.Create().
		SetName(gofakeit.Name()).
		SetSystemOwned(true).
		SetIsPublic(true).
		Save(allowCtx)
	require.NoError(t, err)

	retrievedStd, err := suite.client.Standard.Get(allowCtx, std.ID)
	require.NoError(t, err)
	require.True(t, retrievedStd.SystemOwned, "standard should be system-owned")
	require.True(t, retrievedStd.IsPublic, "standard should be public")

	softDeleteCtx := entx.IsSoftDelete(adminCtx, generated.TypeStandard)

	err = suite.client.Standard.UpdateOneID(std.ID).Exec(softDeleteCtx)
	require.Error(t, err)
	assert.ErrorIs(t, err, hooks.ErrPublicStandardCannotBeDeleted)
}

func (suite *HookTestSuite) TestHookStandardDelete_SystemAdminCascadesSystemOwnedControls() {
	t := suite.T()

	systemAdmin := suite.seedSystemAdmin()
	if len(systemAdmin.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := systemAdmin.Edges.OrgMemberships[0].OrganizationID

	adminCtx := auth.NewTestContextForSystemAdmin(systemAdmin.ID, orgID)
	adminCtx = generated.NewContext(adminCtx, suite.client)
	allowCtx := privacy.DecisionContext(adminCtx, privacy.Allow)

	std, err := suite.client.Standard.Create().
		SetName(gofakeit.Name()).
		SetSystemOwned(true).
		SetIsPublic(false).
		Save(allowCtx)
	require.NoError(t, err)

	systemControl, err := suite.client.Control.Create().
		SetRefCode(gofakeit.UUID()).
		SetTitle(gofakeit.HipsterSentence()).
		SetStandardID(std.ID).
		SetSource(enums.ControlSourceFramework).
		Save(allowCtx)
	require.NoError(t, err)

	orgControl, err := suite.client.Control.Create().
		SetRefCode(gofakeit.UUID()).
		SetTitle(gofakeit.HipsterSentence()).
		SetStandardID(std.ID).
		SetSource(enums.ControlSourceFramework).
		SetOwnerID(orgID).
		Save(allowCtx)
	require.NoError(t, err)

	softDeleteCtx := entx.IsSoftDelete(allowCtx, generated.TypeStandard)

	err = suite.client.Standard.UpdateOneID(std.ID).Exec(softDeleteCtx)
	require.NoError(t, err)

	deletedStd, err := suite.client.Standard.Query().
		Where(standard.ID(std.ID)).
		Only(allowCtx)
	require.NoError(t, err)
	assert.NotNil(t, deletedStd.DeletedAt)

	_, err = suite.client.Control.Get(allowCtx, systemControl.ID)
	require.Error(t, err)
	assert.True(t, generated.IsNotFound(err))

	updatedOrgControl, err := suite.client.Control.Get(allowCtx, orgControl.ID)
	require.NoError(t, err)
	assert.Empty(t, updatedOrgControl.StandardID)
}
