package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	generated "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"
)

// TestAPITokenRevocation verifies that an inactive or revoked API token is rejected at authentication
// time, since expiry alone does not cover tokens that were revoked without backdating their expiration
func TestAPITokenRevocation(t *testing.T) {
	origFetch := fetchAPITokenFunc
	origSSO := isSSOEnforcedFunc
	origAdmin := isSystemAdminFunc

	defer func() {
		fetchAPITokenFunc = origFetch
		isSSOEnforcedFunc = origSSO
		isSystemAdminFunc = origAdmin
	}()

	isSSOEnforcedFunc = func(context.Context, *generated.Client, string) (bool, error) { return false, nil }
	isSystemAdminFunc = func(context.Context, *generated.Client, string, string) (bool, error) { return false, nil }

	t.Run("inactive token is rejected", func(t *testing.T) {
		fetchAPITokenFunc = func(context.Context, *generated.Client, string) (*generated.APIToken, error) {
			return &generated.APIToken{ID: "t1", OwnerID: "org", Token: "tola", IsActive: false}, nil
		}

		_, _, err := isValidAPIToken(context.Background(), (*generated.Client)(nil), "tola")
		assert.ErrorIs(t, err, ErrTokenRevoked)
	})

	t.Run("revoked token is rejected", func(t *testing.T) {
		revoked := time.Now()

		fetchAPITokenFunc = func(context.Context, *generated.Client, string) (*generated.APIToken, error) {
			return &generated.APIToken{ID: "t1", OwnerID: "org", Token: "tola", IsActive: true, RevokedAt: &revoked}, nil
		}

		_, _, err := isValidAPIToken(context.Background(), (*generated.Client)(nil), "tola")
		assert.ErrorIs(t, err, ErrTokenRevoked)
	})

	t.Run("active token passes the revocation check", func(t *testing.T) {
		fetchAPITokenFunc = func(context.Context, *generated.Client, string) (*generated.APIToken, error) {
			return &generated.APIToken{ID: "t1", OwnerID: "org", Name: "n", Token: "tola", IsActive: true}, nil
		}

		au, id, err := isValidAPIToken(context.Background(), (*generated.Client)(nil), "tola")
		assert.NoError(t, err)
		assert.Equal(t, "t1", id)
		assert.Equal(t, "org", au.OrganizationID)
	})
}

// TestPATRevocation verifies that an inactive or revoked personal access token is rejected at
// authentication time
func TestPATRevocation(t *testing.T) {
	origFetch := fetchPATFunc
	origMustSSO := userMustSSOFunc
	origAdmin := isSystemAdminFunc
	origGetOrgRole := getOrgRoleFunc

	defer func() {
		fetchPATFunc = origFetch
		userMustSSOFunc = origMustSSO
		isSystemAdminFunc = origAdmin
		getOrgRoleFunc = origGetOrgRole
	}()

	userMustSSOFunc = func(context.Context, *generated.Client, string, string) (bool, error) { return false, nil }
	isSystemAdminFunc = func(context.Context, *generated.Client, string, string) (bool, error) { return false, nil }

	user := &generated.User{ID: "u1", Email: "u@example.com", DisplayName: "u"}
	org := &generated.Organization{ID: "org", Name: "org"}
	otherOrg := &generated.Organization{ID: "other-org", Name: "other-org"}

	getOrgRoleFunc = func(_ context.Context, _ *generated.Client, _ string, orgID string) *auth.OrganizationRoleType {
		if orgID != org.ID {
			return nil
		}

		role := auth.MemberRole

		return &role
	}

	newPAT := func() *generated.PersonalAccessToken {
		return &generated.PersonalAccessToken{
			ID:      "p1",
			OwnerID: user.ID,
			Token:   "tolp",
			Edges: generated.PersonalAccessTokenEdges{
				Owner:         user,
				Organizations: []*generated.Organization{org},
			},
		}
	}

	t.Run("inactive token is rejected", func(t *testing.T) {
		pat := newPAT()
		pat.IsActive = false

		fetchPATFunc = func(context.Context, *generated.Client, string) (*generated.PersonalAccessToken, error) {
			return pat, nil
		}

		_, _, err := isValidPersonalAccessToken(context.Background(), (*generated.Client)(nil), pat.Token, "")
		assert.ErrorIs(t, err, ErrTokenRevoked)
	})

	t.Run("revoked token is rejected", func(t *testing.T) {
		revoked := time.Now()
		pat := newPAT()
		pat.IsActive = true
		pat.RevokedAt = &revoked

		fetchPATFunc = func(context.Context, *generated.Client, string) (*generated.PersonalAccessToken, error) {
			return pat, nil
		}

		_, _, err := isValidPersonalAccessToken(context.Background(), (*generated.Client)(nil), pat.Token, "")
		assert.ErrorIs(t, err, ErrTokenRevoked)
	})

	t.Run("active token passes the revocation check", func(t *testing.T) {
		pat := newPAT()
		pat.IsActive = true

		fetchPATFunc = func(context.Context, *generated.Client, string) (*generated.PersonalAccessToken, error) {
			return pat, nil
		}

		au, id, err := isValidPersonalAccessToken(context.Background(), (*generated.Client)(nil), pat.Token, "")
		assert.NoError(t, err)
		assert.Equal(t, pat.ID, id)
		assert.Equal(t, user.ID, au.SubjectID)
	})

	t.Run("orgs the owner is not a member of are dropped", func(t *testing.T) {
		pat := newPAT()
		pat.IsActive = true
		pat.Edges.Organizations = []*generated.Organization{org, otherOrg}

		fetchPATFunc = func(context.Context, *generated.Client, string) (*generated.PersonalAccessToken, error) {
			return pat, nil
		}

		au, _, err := isValidPersonalAccessToken(context.Background(), (*generated.Client)(nil), pat.Token, "")
		assert.NoError(t, err)
		assert.Equal(t, []string{org.ID}, au.OrganizationIDs)
		assert.Equal(t, auth.MemberRole, au.OrganizationRole)
	})

	t.Run("token with no member orgs is rejected", func(t *testing.T) {
		pat := newPAT()
		pat.IsActive = true
		pat.Edges.Organizations = []*generated.Organization{otherOrg}

		fetchPATFunc = func(context.Context, *generated.Client, string) (*generated.PersonalAccessToken, error) {
			return pat, nil
		}

		_, _, err := isValidPersonalAccessToken(context.Background(), (*generated.Client)(nil), pat.Token, "")
		assert.ErrorIs(t, err, rout.ErrInvalidCredentials)
	})
}
