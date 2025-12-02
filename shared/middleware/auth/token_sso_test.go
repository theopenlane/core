package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	generated "github.com/theopenlane/ent/generated"
	"github.com/theopenlane/shared/models"
)

func TestAPITokenSSOAuthorization(t *testing.T) {
	tok := &generated.APIToken{
		ID:      "t1",
		OwnerID: "org",
		Name:    "token",
		Token:   "tola_token",
	}

	origFetch := fetchAPITokenFunc
	origSSO := isSSOEnforcedFunc
	defer func() { fetchAPITokenFunc = origFetch; isSSOEnforcedFunc = origSSO }()

	fetchAPITokenFunc = func(context.Context, *generated.Client, string) (*generated.APIToken, error) {
		return tok, nil
	}
	isSSOEnforcedFunc = func(context.Context, *generated.Client, string) (bool, error) { return true, nil }
	isSystemAdminFunc = func(context.Context, *generated.Client, string, string) (bool, error) {
		return false, nil
	}

	_, _, err := isValidAPIToken(context.Background(), (*generated.Client)(nil), tok.Token)
	assert.ErrorIs(t, err, ErrTokenSSORequired)

	tok.SSOAuthorizations = models.SSOAuthorizationMap{"org": time.Now()}

	au, id, err := isValidAPIToken(context.Background(), (*generated.Client)(nil), tok.Token)
	assert.NoError(t, err)
	assert.Equal(t, tok.ID, id)
	assert.Equal(t, []string{tok.OwnerID}, au.OrganizationIDs)
}

func TestPATTokenSSOAuthorization(t *testing.T) {
	user := &generated.User{ID: "u1", Email: "u@example.com", DisplayName: "u"}
	org := &generated.Organization{ID: "org", Name: "org"}
	pat := &generated.PersonalAccessToken{
		ID:      "p1",
		OwnerID: user.ID,
		Token:   "tolp_token",
		Edges: generated.PersonalAccessTokenEdges{
			Owner:         user,
			Organizations: []*generated.Organization{org},
		},
	}

	origPAT := fetchPATFunc
	origSSO := isSSOEnforcedFunc
	origAuth := isPATSSOAuthorizedFunc
	defer func() {
		fetchPATFunc = origPAT
		isSSOEnforcedFunc = origSSO
		isPATSSOAuthorizedFunc = origAuth
	}()

	fetchPATFunc = func(context.Context, *generated.Client, string) (*generated.PersonalAccessToken, error) {
		return pat, nil
	}
	isSSOEnforcedFunc = func(context.Context, *generated.Client, string) (bool, error) { return true, nil }
	isPATSSOAuthorizedFunc = func(context.Context, *generated.Client, string, string) (bool, error) {
		if pat.SSOAuthorizations == nil {
			return false, nil
		}
		_, ok := pat.SSOAuthorizations[org.ID]
		return ok, nil
	}
	isSystemAdminFunc = func(context.Context, *generated.Client, string, string) (bool, error) {
		return false, nil
	}

	_, _, err := isValidPersonalAccessToken(context.Background(), (*generated.Client)(nil), pat.Token, "")
	assert.ErrorIs(t, err, ErrTokenSSORequired)

	pat.SSOAuthorizations = models.SSOAuthorizationMap{org.ID: time.Now()}

	au, id, err := isValidPersonalAccessToken(context.Background(), (*generated.Client)(nil), pat.Token, "")
	assert.NoError(t, err)
	assert.Equal(t, pat.ID, id)
	assert.Equal(t, user.ID, au.SubjectID)
	assert.Equal(t, []string{org.ID}, au.OrganizationIDs)
	assert.Empty(t, au.OrganizationID)

	userDefaultOrg := org.ID

	au, id, err = isValidPersonalAccessToken(context.Background(), (*generated.Client)(nil), pat.Token, userDefaultOrg)
	assert.NoError(t, err)
	assert.Equal(t, org.ID, au.OrganizationID)
}
