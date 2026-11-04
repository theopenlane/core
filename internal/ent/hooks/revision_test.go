package hooks_test

import (
	"context"
	"testing"

	"entgo.io/ent"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/iam/auth"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestSetNewRevision(t *testing.T) {
	s := suite.setupTest(t)

	// Create a system admin user (needed right now for standards until we allow other users to create them and we aren't testing permissions here)
	testUser := s.seedSystemAdmin()
	orgID := testUser.Edges.OrgMemberships[0].ID
	ctx := auth.NewTestContextForSystemAdmin(testUser.ID, orgID)

	// add the client to the context for hooks
	ctx = generated.NewContext(ctx, s.client)

	patchCtx := auth.NewTestContextForSystemAdmin(testUser.ID, orgID)
	models.WithVersionBumpRequestContext(patchCtx, &models.Patch)

	minorCtx := auth.NewTestContextForSystemAdmin(testUser.ID, orgID)
	models.WithVersionBumpRequestContext(minorCtx, &models.Minor)

	majorCtx := auth.NewTestContextForSystemAdmin(testUser.ID, orgID)
	models.WithVersionBumpRequestContext(majorCtx, &models.Major)

	draftCtx := auth.NewTestContextForSystemAdmin(testUser.ID, orgID)
	models.WithVersionBumpRequestContext(draftCtx, &models.PreRelease)

	tests := []struct {
		name             string
		mutation         ent.Mutation
		ctx              context.Context
		expectedRevision string
		expectedErr      error
	}{
		{
			name: "mutation with revision set",
			mutation: func() *generated.StandardMutation {
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.UpdateOne(std).SetRevision("v0.4.3")

				err = update.Exec(ctx)
				assert.NilError(t, err)

				return update.Mutation()
			}(),
			ctx:              context.Background(),
			expectedRevision: "v0.4.3",
			expectedErr:      nil,
		},
		{
			name: "mutation with nothing set - defaults to patch",
			mutation: func() *generated.StandardMutation {
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.UpdateOne(std)

				err = update.Exec(ctx)
				assert.NilError(t, err)

				return update.Mutation()
			}(),
			ctx:              ctx,
			expectedRevision: "v0.0.2",
			expectedErr:      nil,
		},
		{
			name: "mutation with revision bump set to patch",
			mutation: func() *generated.StandardMutation {
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.UpdateOne(std)

				err = update.Exec(patchCtx)
				assert.NilError(t, err)

				return update.Mutation()
			}(),
			ctx:              patchCtx,
			expectedRevision: "v0.0.2",
			expectedErr:      nil,
		},
		{
			name: "mutation with revision bump set to major",
			mutation: func() *generated.StandardMutation {
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.UpdateOne(std)

				err = update.Exec(majorCtx)
				assert.NilError(t, err)

				return update.Mutation()
			}(),
			ctx:              majorCtx,
			expectedRevision: "v1.0.0",
			expectedErr:      nil,
		},
		{
			name: "mutation with revision bump set to minor",
			mutation: func() *generated.StandardMutation {
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.UpdateOne(std)

				err = update.Exec(minorCtx)
				assert.NilError(t, err)

				return update.Mutation()
			}(),
			ctx:              minorCtx,
			expectedRevision: "v0.1.0",
			expectedErr:      nil,
		},
		{
			name: "mutation with revision bump set to draft",
			mutation: func() *generated.StandardMutation {
				std, err := s.client.Standard.Create().SetName(gofakeit.Name()).Save(ctx)
				assert.NilError(t, err)

				update := s.client.Standard.UpdateOne(std)

				err = update.Exec(draftCtx)
				assert.NilError(t, err)

				return update.Mutation()
			}(),
			ctx:              draftCtx,
			expectedRevision: "v0.0.2-draft",
			expectedErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mut := tt.mutation.(hooks.MutationWithRevision)
			err := hooks.SetNewRevision(tt.ctx, mut)

			if tt.expectedErr != nil {
				assert.Check(t, err != nil)

				return
			}

			assert.NilError(t, err)

			out, ok := mut.Revision()
			assert.Check(t, ok)
			assert.Check(t, is.Equal(tt.expectedRevision, out))
		})
	}
}
