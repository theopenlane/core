package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/notification"
	"github.com/theopenlane/iam/auth"
)

func TestNotificationQueryFilter_Creation(t *testing.T) {
	// Test that the filter can be created without error
	filter := NotificationQueryFilter()
	assert.NotNil(t, filter)
}

func TestNotificationQueryFilter_WithValidUserContext(t *testing.T) {
	ctx := auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
		SubjectID:       "01HHAS67AM73778S0QEZ3CEAGE",
		OrganizationIDs: []string{"01HHAS67AM73778S0QEZ3CEAGF", "01HHAS67AM73778S0QEZ3CEAGG"},
	})

	// Test that we can get the subject ID
	subjectID, err := auth.GetSubjectIDFromContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "01HHAS67AM73778S0QEZ3CEAGE", subjectID)

	// Test that we can get org IDs
	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	assert.NoError(t, err)
	assert.Len(t, orgIDs, 2)
	assert.Contains(t, orgIDs, "01HHAS67AM73778S0QEZ3CEAGF")
	assert.Contains(t, orgIDs, "01HHAS67AM73778S0QEZ3CEAGG")
}

func TestNotificationQueryFilter_WithNoAuthContext(t *testing.T) {
	ctx := context.Background()

	// Test that GetSubjectIDFromContext returns error with no auth
	_, err := auth.GetSubjectIDFromContext(ctx)
	assert.Error(t, err)
}

func TestNotificationQueryFilter_WithUserButNoOrgIDs(t *testing.T) {
	ctx := auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
		SubjectID:       "01HHAS67AM73778S0QEZ3CEAGE",
		OrganizationIDs: []string{},
	})

	subjectID, err := auth.GetSubjectIDFromContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "01HHAS67AM73778S0QEZ3CEAGE", subjectID)

	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	assert.NoError(t, err)
	assert.Empty(t, orgIDs)
}

func TestNotificationQueryFilter_WithSingleOrgID(t *testing.T) {
	ctx := auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
		SubjectID:       "01HHAS67AM73778S0QEZ3CEAGE",
		OrganizationIDs: []string{"01HHAS67AM73778S0QEZ3CEAGF"},
	})

	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	assert.NoError(t, err)
	assert.Len(t, orgIDs, 1)
	assert.Equal(t, "01HHAS67AM73778S0QEZ3CEAGF", orgIDs[0])
}

func TestNotificationQueryFilter_QueryTypeCheck(t *testing.T) {
	client := generated.NewClient()

	// Test NotificationQuery type
	nq := client.Notification.Query()
	assert.IsType(t, &generated.NotificationQuery{}, nq)

	// Test that other query types are different
	uq := client.User.Query()
	assert.IsType(t, &generated.UserQuery{}, uq)
	assert.NotEqual(t, nq, uq)
}

func TestNotificationQueryFilter_Predicates(t *testing.T) {
	userID := "01HHAS67AM73778S0QEZ3CEAGE"
	orgID := "01HHAS67AM73778S0QEZ3CEAGF"

	// Test individual predicates
	userPredicate := notification.UserID(userID)
	assert.NotNil(t, userPredicate)

	nilPredicate := notification.UserIDIsNil()
	assert.NotNil(t, nilPredicate)

	ownerPredicate := notification.OwnerIDIn(orgID)
	assert.NotNil(t, ownerPredicate)

	// Test combined predicates
	combined := notification.Or(
		userPredicate,
		notification.And(
			nilPredicate,
			ownerPredicate,
		),
	)
	assert.NotNil(t, combined)
}

func TestNotificationQueryFilter_WithAnonymousQuestionnaireUser(t *testing.T) {
	ctx := auth.WithAnonymousQuestionnaireUser(context.Background(), &auth.AnonymousQuestionnaireUser{
		SubjectID:      "01HHAS67AM73778S0QEZ3CEAGE",
		OrganizationID: "01HHAS67AM73778S0QEZ3CEAGF",
	})

	// Anonymous questionnaire user should be retrievable as authenticated user
	user, ok := auth.AuthenticatedUserFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "01HHAS67AM73778S0QEZ3CEAGE", user.SubjectID)
	assert.Equal(t, "01HHAS67AM73778S0QEZ3CEAGF", user.OrganizationID)
}

func TestNotificationQueryFilter_MultipleOrgIDs(t *testing.T) {
	orgID1 := "01HHAS67AM73778S0QEZ3CEAGF"
	orgID2 := "01HHAS67AM73778S0QEZ3CEAGG"
	orgID3 := "01HHAS67AM73778S0QEZ3CEAGH"

	ctx := auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
		SubjectID:       "01HHAS67AM73778S0QEZ3CEAGE",
		OrganizationIDs: []string{orgID1, orgID2, orgID3},
	})

	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	assert.NoError(t, err)
	assert.Len(t, orgIDs, 3)

	// Test that the owner predicate can handle multiple org IDs
	ownerPredicate := notification.OwnerIDIn(orgID1, orgID2, orgID3)
	assert.NotNil(t, ownerPredicate)
}

func TestNotificationQueryFilter_ErrorOnNoAuth(t *testing.T) {
	// Test auth error handling
	ctx := context.Background()

	_, err := auth.GetSubjectIDFromContext(ctx)
	assert.Error(t, err)
	assert.Equal(t, auth.ErrNoAuthUser, err)

	_, err = auth.GetOrganizationIDsFromContext(ctx)
	assert.Error(t, err)
	assert.Equal(t, auth.ErrNoAuthUser, err)
}

func TestNotificationQueryFilter_PredicateCombinations(t *testing.T) {
	// Test various predicate combinations that the filter uses
	userID := "01HHAS67AM73778S0QEZ3CEAGE"
	orgID1 := "01HHAS67AM73778S0QEZ3CEAGF"
	orgID2 := "01HHAS67AM73778S0QEZ3CEAGG"

	// Test single org predicate
	p1 := notification.OwnerIDIn(orgID1)
	assert.NotNil(t, p1)

	// Test multiple org predicate
	p2 := notification.OwnerIDIn(orgID1, orgID2)
	assert.NotNil(t, p2)

	// Test user ID predicate
	p3 := notification.UserID(userID)
	assert.NotNil(t, p3)

	// Test combined with And
	p4 := notification.And(
		notification.UserIDIsNil(),
		notification.OwnerIDIn(orgID1),
	)
	assert.NotNil(t, p4)

	// Test combined with Or
	p5 := notification.Or(
		notification.UserID(userID),
		notification.And(
			notification.UserIDIsNil(),
			notification.OwnerIDIn(orgID1, orgID2),
		),
	)
	assert.NotNil(t, p5)
}

func TestNotificationQueryFilter_ContextPaths(t *testing.T) {
	// Test different context configurations
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name: "valid user with orgs",
			ctx: auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
				SubjectID:       "01HHAS67AM73778S0QEZ3CEAGE",
				OrganizationIDs: []string{"01HHAS67AM73778S0QEZ3CEAGF"},
			}),
			wantErr: false,
		},
		{
			name: "valid user without orgs",
			ctx: auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
				SubjectID:       "01HHAS67AM73778S0QEZ3CEAGE",
				OrganizationIDs: []string{},
			}),
			wantErr: false,
		},
		{
			name:    "no auth context",
			ctx:     context.Background(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := auth.GetSubjectIDFromContext(tt.ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
