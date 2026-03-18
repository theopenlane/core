//go:build test

package hooks_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	controlgen "github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
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

			assert.Equal(t, tt.expectedWrite, shouldWrite, "shouldWrite mismatch")
			assert.Equal(t, tt.expectedDelete, shouldDelete, "shouldDelete mismatch")
		})
	}
}

func (suite *HookTestSuite) TestHookControlTrustCenterVisibility_UpdateOpsParity() {
	t := suite.T()

	systemAdmin := suite.seedSystemAdmin()
	if len(systemAdmin.Edges.OrgMemberships) == 0 {
		t.Fatal("user has no org memberships")
	}

	orgID := systemAdmin.Edges.OrgMemberships[0].OrganizationID

	ctx := auth.NewTestContextForSystemAdmin(systemAdmin.ID, orgID)
	ctx = generated.NewContext(ctx, suite.client)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	checkWildcardViewerAccess := func(tb testing.TB, controlID string) bool {
		tb.Helper()

		allowed, err := suite.client.Authz.CheckAccessHighConsistency(context.Background(), fgax.AccessCheck{
			SubjectID:   fgax.Wildcard,
			SubjectType: auth.UserSubjectType,
			Relation:    fgax.CanView,
			ObjectID:    controlID,
			ObjectType:  fgax.Kind(generated.TypeControl),
		})
		require.NoError(tb, err)

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
				return suite.client.Control.UpdateOneID(controlID).
					SetTrustCenterVisibility(enums.TrustCenterControlVisibilityPubliclyVisible).
					Exec(ctx)
			},
			setNotVisible: func(ctx context.Context, controlID string) error {
				return suite.client.Control.UpdateOneID(controlID).
					SetTrustCenterVisibility(enums.TrustCenterControlVisibilityNotVisible).
					Exec(ctx)
			},
		},
		{
			name: "bulk update",
			setPublic: func(ctx context.Context, controlID string) error {
				return suite.client.Control.Update().
					Where(controlgen.ID(controlID)).
					SetTrustCenterVisibility(enums.TrustCenterControlVisibilityPubliclyVisible).
					Exec(ctx)
			},
			setNotVisible: func(ctx context.Context, controlID string) error {
				return suite.client.Control.Update().
					Where(controlgen.ID(controlID)).
					SetTrustCenterVisibility(enums.TrustCenterControlVisibilityNotVisible).
					Exec(ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl, err := suite.client.Control.Create().
				SetRefCode(gofakeit.UUID()).
				SetOwnerID(orgID).
				SetIsTrustCenterControl(true).
				SetTrustCenterVisibility(enums.TrustCenterControlVisibilityNotVisible).
				Save(allowCtx)
			require.NoError(t, err)

			assert.False(t, checkWildcardViewerAccess(t, ctrl.ID))

			err = tt.setPublic(allowCtx, ctrl.ID)
			require.NoError(t, err)
			assert.True(t, checkWildcardViewerAccess(t, ctrl.ID))

			err = tt.setNotVisible(allowCtx, ctrl.ID)
			require.NoError(t, err)
			assert.False(t, checkWildcardViewerAccess(t, ctrl.ID))
		})
	}
}
