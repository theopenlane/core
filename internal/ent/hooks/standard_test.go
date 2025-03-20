package hooks_test

import (
	"context"
	"testing"

	"entgo.io/ent"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

func (suite *HookTestSuite) TestAddOrDeleteStandardTuple() {
	t := suite.T()

	user := suite.seedSystemAdmin()

	if len(user.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := user.Edges.OrgMemberships[0].ID

	ctx := auth.NewTestContextWithOrgID(user.ID, orgID)
	ctx = generated.NewContext(ctx, suite.client)

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
			name: "UpdateOne with systemOwned cleared",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).SetSystemOwned(true).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.UpdateOne(std).ClearSystemOwned()

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
			name: "UpdateOne with systemOwned cleared, will try to re-delete",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.UpdateOne(std).ClearSystemOwned()

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
			name: "UpdateOne with systemOwned and isPublic both cleared",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.UpdateOne(std).ClearSystemOwned().ClearIsPublic()

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
			name: "UpdateOne with clear is public, should delete",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).SetSystemOwned(true).SetIsPublic(true).Save(ctx)
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
			name: "UpdateOne with systemOwned true, no-op because not public",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.UpdateOne(std).SetSystemOwned(true)

				err = update.Exec(ctx)
				require.NoError(t, err)

				return update.Mutation()
			}(),
			ctx:            ctx,
			expectedAdd:    false,
			expectedDelete: false,
			expectedErr:    nil,
		},
		{
			name: "UpdateOne with is public true, no-op because not system owned",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.UpdateOne(std).SetIsPublic(true)

				err = update.Exec(ctx)
				require.NoError(t, err)

				return update.Mutation()
			}(),
			ctx:            ctx,
			expectedAdd:    false,
			expectedDelete: false,
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
			name: "UpdateOne with systemOwned true, public true",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.UpdateOne(std).SetSystemOwned(true).SetIsPublic(true)

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
			name: "Update with systemOwned true, public true",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.Update().SetSystemOwned(true).SetIsPublic(true).Where(standard.ID(std.ID))

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
			name: "Update with systemOwned false",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).SetIsPublic(true).SetSystemOwned(true).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.Update().SetSystemOwned(false).Where(standard.ID(std.ID))

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
			name: "Update with system owned true, already public",
			mutation: func() *generated.StandardMutation {
				std, err := suite.client.Standard.Create().SetName(gofakeit.Name()).SetIsPublic(true).Save(ctx)
				require.NoError(t, err)

				update := suite.client.Standard.Update().SetSystemOwned(true).Where(standard.ID(std.ID))

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
			name: "UpdateOne with soft delete",
			mutation: func() *generated.StandardMutation {
				m := generated.StandardMutation{}
				m.SetOp(ent.OpUpdateOne)
				m.SetID(ulids.New().String())
				return &m
			}(),
			ctx:            entx.IsSoftDelete(ctx),
			expectedAdd:    false,
			expectedDelete: true,
			expectedErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			add, delete, err := hooks.AddOrDeleteStandardTuple(tt.ctx, tt.mutation)
			assert.Equal(t, tt.expectedAdd, add)
			assert.Equal(t, tt.expectedDelete, delete)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
