package hooks_test

import (
	"context"
	"testing"

	"entgo.io/ent"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/core/v2/internal/ent/hooks"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestAddOrDeleteStandardTuple(t *testing.T) {
	s := suite.setupTest(t)

	user := s.seedSystemAdmin()
	if len(user.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := user.Edges.OrgMemberships[0].ID

	ctx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	ctx = generated.NewContext(ctx, s.client)

	// soft delete context
	softDeleteCtx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	softDeleteCtx = generated.NewContext(softDeleteCtx, s.client)
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
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).SetSystemOwned(true).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.UpdateOne(std).ClearIsPublic()

				err = update.Exec(ctx)
				assert.NilError(t, err)

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
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.UpdateOne(std).ClearIsPublic()

				err = update.Exec(ctx)
				assert.NilError(t, err)

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
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).SetSystemOwned(true).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.UpdateOne(std).SetIsPublic(true)

				err = update.Exec(ctx)
				assert.NilError(t, err)

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
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.Update().SetIsPublic(true).Where(standard.ID(std.ID))

				err = update.Exec(ctx)
				assert.NilError(t, err)

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
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).SetIsPublic(true).SetSystemOwned(true).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.Update().SetIsPublic(false).Where(standard.ID(std.ID))

				err = update.Exec(ctx)
				assert.NilError(t, err)

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
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).SetSystemOwned(true).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.Update().SetIsPublic(true).Where(standard.ID(std.ID))

				err = update.Exec(ctx)
				assert.NilError(t, err)

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
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).SetSystemOwned(true).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.UpdateOne(std)
				err = update.Exec(softDeleteCtx)
				assert.NilError(t, err)

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
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.UpdateOne(std)
				err = update.Exec(softDeleteCtx)
				assert.NilError(t, err)

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
			assert.Check(t, is.Equal(tt.expectedAdd, add))
			assert.Check(t, is.Equal(tt.expectedDelete, delete))
			assert.Check(t, is.Equal(tt.expectedErr, err))
		})
	}
}

func TestHookStandardDelete_InUseByTrustcenter(t *testing.T) {
	s := suite.setupTest(t)

	user := s.seedSystemAdmin()
	if len(user.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := user.Edges.OrgMemberships[0].OrganizationID

	ctx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	ctx = generated.NewContext(ctx, s.client)
	internalCtx := auth.WithInternalOperationContext(ctx)

	std, err := s.client.Standard.Create().
		SetName(gofakeit.Name()).
		Save(internalCtx)
	assert.NilError(t, err)

	trustCenter, err := s.client.TrustCenter.Create().
		SetOwnerID(orgID).
		Save(internalCtx)
	assert.NilError(t, err)

	_, err = s.client.TrustCenterCompliance.Create().
		SetStandardID(std.ID).
		SetTrustCenterID(trustCenter.ID).
		Save(internalCtx)
	assert.NilError(t, err)

	softDeleteCtx := entx.IsSoftDelete(ctx, generated.TypeStandard)

	err = s.client.Standard.UpdateOneID(std.ID).Exec(softDeleteCtx)
	assert.Assert(t, err != nil)
	assert.Check(t, is.ErrorIs(err, hooks.ErrStandardInUseByTrustCenter))
}

func TestHookStandardDelete_InUseByControls(t *testing.T) {
	s := suite.setupTest(t)

	regularUser := s.seedUser()
	if len(regularUser.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := regularUser.Edges.OrgMemberships[0].OrganizationID

	regularUserCtx := auth.NewTestContextWithOrgID(regularUser.ID, orgID)
	regularUserCtx = generated.NewContext(regularUserCtx, s.client)
	internalCtx := auth.WithInternalOperationContext(regularUserCtx)

	std, err := s.client.Standard.Create().
		SetName(gofakeit.Name()).
		Save(internalCtx)
	assert.NilError(t, err)

	_, err = s.client.Control.Create().
		SetRefCode(gofakeit.UUID()).
		SetTitle(gofakeit.HipsterSentence()).
		SetStandardID(std.ID).
		SetSource(enums.ControlSourceFramework).
		SetOwnerID(orgID).
		Save(internalCtx)
	assert.NilError(t, err)

	systemAdmin := s.seedSystemAdmin()
	adminCtx := auth.NewTestContextForSystemAdmin(systemAdmin.ID, orgID)
	adminCtx = generated.NewContext(adminCtx, s.client)
	softDeleteCtx := entx.IsSoftDelete(adminCtx, generated.TypeStandard)

	err = s.client.Standard.UpdateOneID(std.ID).Exec(softDeleteCtx)
	assert.Assert(t, err != nil)
	assert.Check(t, is.ErrorIs(err, hooks.ErrStandardInUseByControls))
}

func TestHookStandardDelete_AllowsWhenNotInUse(t *testing.T) {
	s := suite.setupTest(t)

	user := s.seedSystemAdmin()
	if len(user.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := user.Edges.OrgMemberships[0].OrganizationID

	ctx := auth.NewTestContextForSystemAdmin(user.ID, orgID)
	ctx = generated.NewContext(ctx, s.client)
	internalCtx := auth.WithInternalOperationContext(ctx)

	std, err := s.client.Standard.Create().
		SetName(gofakeit.Name()).
		Save(internalCtx)
	assert.NilError(t, err)

	err = s.client.Standard.DeleteOneID(std.ID).Exec(internalCtx)
	assert.NilError(t, err)

	deletedStd, err := s.client.Standard.Query().
		Where(standard.ID(std.ID)).
		Only(entx.SkipSoftDelete(internalCtx))
	assert.NilError(t, err)
	assert.Check(t, !deletedStd.DeletedAt.IsZero())
}

func TestHookStandardDelete_RegularUserCanDeleteNonSystemOwned(t *testing.T) {
	s := suite.setupTest(t)

	regularUser := s.seedUser()
	if len(regularUser.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := regularUser.Edges.OrgMemberships[0].OrganizationID

	regularUserCtx := auth.NewTestContextWithOrgID(regularUser.ID, orgID)
	regularUserCtx = generated.NewContext(regularUserCtx, s.client)
	internalCtx := auth.WithInternalOperationContext(regularUserCtx)

	std, err := s.client.Standard.Create().
		SetName(gofakeit.Name()).
		SetSystemOwned(false).
		Save(internalCtx)
	assert.NilError(t, err)

	err = s.client.Standard.DeleteOneID(std.ID).Exec(internalCtx)
	assert.NilError(t, err)

	deletedStd, err := s.client.Standard.Query().
		Where(standard.ID(std.ID)).
		Only(entx.SkipSoftDelete(internalCtx))
	assert.NilError(t, err)
	assert.Check(t, !deletedStd.DeletedAt.IsZero())
}

func TestHookStandardDelete_BlocksPublicSystemOwned(t *testing.T) {
	s := suite.setupTest(t)

	systemAdmin := s.seedSystemAdmin()
	if len(systemAdmin.Edges.OrgMemberships) == 0 {
		t.Fatal("system admin has no org memberships")
	}

	orgID := systemAdmin.Edges.OrgMemberships[0].OrganizationID

	adminCtx := auth.NewTestContextForSystemAdmin(systemAdmin.ID, orgID)
	adminCtx = generated.NewContext(adminCtx, s.client)
	internalCtx := auth.WithInternalOperationContext(adminCtx)

	std, err := s.client.Standard.Create().
		SetName(gofakeit.Name()).
		SetSystemOwned(true).
		SetIsPublic(true).
		Save(internalCtx)
	assert.NilError(t, err)

	retrievedStd, err := s.client.Standard.Get(internalCtx, std.ID)
	assert.NilError(t, err)
	assert.Assert(t, retrievedStd.SystemOwned, "standard should be system-owned")
	assert.Assert(t, retrievedStd.IsPublic, "standard should be public")

	softDeleteCtx := entx.IsSoftDelete(adminCtx, generated.TypeStandard)

	err = s.client.Standard.UpdateOneID(std.ID).Exec(softDeleteCtx)
	assert.Assert(t, err != nil)
	assert.Check(t, is.ErrorIs(err, hooks.ErrPublicStandardCannotBeDeleted))
}

func TestHookStandardDelete_SystemAdminCascadesSystemOwnedControls(t *testing.T) {
	s := suite.setupTest(t)

	systemAdmin := s.seedSystemAdmin()
	if len(systemAdmin.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := systemAdmin.Edges.OrgMemberships[0].OrganizationID

	adminCtx := auth.NewTestContextForSystemAdmin(systemAdmin.ID, orgID)
	adminCtx = generated.NewContext(adminCtx, s.client)
	internalCtx := auth.WithInternalOperationContext(adminCtx)

	std, err := s.client.Standard.Create().
		SetName(gofakeit.Name()).
		SetSystemOwned(true).
		SetIsPublic(false).
		Save(internalCtx)
	assert.NilError(t, err)

	systemControl, err := s.client.Control.Create().
		SetRefCode(gofakeit.UUID()).
		SetTitle(gofakeit.HipsterSentence()).
		SetStandardID(std.ID).
		SetSource(enums.ControlSourceFramework).
		Save(internalCtx)
	assert.NilError(t, err)

	orgControl, err := s.client.Control.Create().
		SetRefCode(gofakeit.UUID()).
		SetTitle(gofakeit.HipsterSentence()).
		SetStandardID(std.ID).
		SetSource(enums.ControlSourceFramework).
		SetOwnerID(orgID).
		Save(internalCtx)
	assert.NilError(t, err)

	err = s.client.Standard.DeleteOneID(std.ID).Exec(internalCtx)
	assert.NilError(t, err)

	deletedStd, err := s.client.Standard.Query().
		Where(standard.ID(std.ID)).
		Only(entx.SkipSoftDelete(internalCtx))
	assert.NilError(t, err)
	assert.Check(t, !deletedStd.DeletedAt.IsZero())

	_, err = s.client.Control.Get(internalCtx, systemControl.ID)
	assert.Assert(t, err != nil)
	assert.Check(t, generated.IsNotFound(err))

	updatedOrgControl, err := s.client.Control.Get(internalCtx, orgControl.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(updatedOrgControl.StandardID, ""))
}
