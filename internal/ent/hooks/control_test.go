//go:build test

package hooks_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	controlgen "github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/hooks"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestControlVisibilityTupleAction(t *testing.T) {
	tests := []struct {
		name              string
		newVisibility     enums.TrustCenterControlVisibility
		oldVisibility     enums.TrustCenterControlVisibility
		visibilityChanged bool
		expectedWrite     bool
		expectedDelete    bool
	}{
		{
			name:              "visibility not changed, no action",
			newVisibility:     enums.TrustCenterControlVisibilityPubliclyVisible,
			oldVisibility:     enums.TrustCenterControlVisibilityPubliclyVisible,
			visibilityChanged: false,
			expectedWrite:     false,
			expectedDelete:    false,
		},
		{
			name:              "same visibility values, no action",
			newVisibility:     enums.TrustCenterControlVisibilityNotVisible,
			oldVisibility:     enums.TrustCenterControlVisibilityNotVisible,
			visibilityChanged: true,
			expectedWrite:     false,
			expectedDelete:    false,
		},
		{
			name:              "not visible to publicly visible, should write",
			newVisibility:     enums.TrustCenterControlVisibilityPubliclyVisible,
			oldVisibility:     enums.TrustCenterControlVisibilityNotVisible,
			visibilityChanged: true,
			expectedWrite:     true,
			expectedDelete:    false,
		},
		{
			name:              "publicly visible to not visible, should delete",
			newVisibility:     enums.TrustCenterControlVisibilityNotVisible,
			oldVisibility:     enums.TrustCenterControlVisibilityPubliclyVisible,
			visibilityChanged: true,
			expectedWrite:     false,
			expectedDelete:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldWrite, shouldDelete := hooks.ControlVisibilityTupleAction(
				tt.newVisibility,
				tt.oldVisibility,
				tt.visibilityChanged,
			)

			assert.Check(t, is.Equal(tt.expectedWrite, shouldWrite), "shouldWrite mismatch")
			assert.Check(t, is.Equal(tt.expectedDelete, shouldDelete), "shouldDelete mismatch")
		})
	}
}

func TestHookControlTrustCenterVisibility_UpdateOpsParity(t *testing.T) {
	s := suite.setupTest(t)

	systemAdmin := s.seedSystemAdmin()
	if len(systemAdmin.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := systemAdmin.Edges.OrgMemberships[0].OrganizationID

	ctx := auth.NewTestContextForSystemAdmin(systemAdmin.ID, orgID)
	ctx = generated.NewContext(ctx, s.client)
	internalCtx := auth.WithInternalOperationContext(ctx)

	checkWildcardViewerAccess := func(tb testing.TB, controlID string) bool {
		tb.Helper()

		allowed, err := s.client.Authz.CheckAccessHighConsistency(context.Background(), fgax.AccessCheck{
			SubjectID:   fgax.Wildcard,
			SubjectType: auth.UserSubjectType,
			Relation:    fgax.CanView,
			ObjectID:    controlID,
			ObjectType:  fgax.Kind(generated.TypeControl),
		})
		assert.NilError(tb, err)

		return allowed
	}

	tests := []struct {
		name          string
		setPublic     func(context.Context, string) error
		setNotVisible func(context.Context, string) error
	}{
		{
			name: "update one",
			setPublic: func(ctx context.Context, controlID string) error {
				return s.client.Control.UpdateOneID(controlID).
					SetTrustCenterVisibility(enums.TrustCenterControlVisibilityPubliclyVisible).
					Exec(ctx)
			},
			setNotVisible: func(ctx context.Context, controlID string) error {
				return s.client.Control.UpdateOneID(controlID).
					SetTrustCenterVisibility(enums.TrustCenterControlVisibilityNotVisible).
					Exec(ctx)
			},
		},
		{
			name: "bulk update",
			setPublic: func(ctx context.Context, controlID string) error {
				return s.client.Control.Update().
					Where(controlgen.ID(controlID)).
					SetTrustCenterVisibility(enums.TrustCenterControlVisibilityPubliclyVisible).
					Exec(ctx)
			},
			setNotVisible: func(ctx context.Context, controlID string) error {
				return s.client.Control.Update().
					Where(controlgen.ID(controlID)).
					SetTrustCenterVisibility(enums.TrustCenterControlVisibilityNotVisible).
					Exec(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl, err := s.client.Control.Create().
				SetRefCode(gofakeit.UUID()).
				SetOwnerID(orgID).
				SetIsTrustCenterControl(true).
				SetTrustCenterVisibility(enums.TrustCenterControlVisibilityNotVisible).
				Save(internalCtx)
			assert.NilError(t, err)

			assert.Check(t, !checkWildcardViewerAccess(t, ctrl.ID))

			err = tt.setPublic(internalCtx, ctrl.ID)
			assert.NilError(t, err)
			assert.Check(t, checkWildcardViewerAccess(t, ctrl.ID))

			err = tt.setNotVisible(internalCtx, ctrl.ID)
			assert.NilError(t, err)
			assert.Check(t, !checkWildcardViewerAccess(t, ctrl.ID))
		})
	}
}
